package grmux

import (
	"context"
	"errors"
	"github.com/oldjon/gutil/gdb"
	"github.com/opentracing/opentracing-go"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

var (
	errRedisLockBusy = errors.New("redis lock busy")
)

const (
	redisMuxPrefix             = "rmux:"
	defaultExpiration          = 5 * time.Second // ms
	defaultLockRetryTimes      = 30
	defaultSleepTimeExpandStep = 10 * time.Millisecond // ms
	defaultSleepTimeFloat      = 10 * time.Millisecond // ms

	unlockScript = `if redis.call("get",KEYS[1]) == ARGV[1] then return redis.call("del",KEYS[1]) else return 0 end`
)

type RedisMuxOption struct {
	Expiration          time.Duration
	RetryTimes          int
	SleepTimeExpandStep time.Duration
	SleepTimeFloat      time.Duration
}

func (ro *RedisMuxOption) init() {
	if ro.Expiration == 0 {
		ro.Expiration = defaultExpiration
	}
	if ro.RetryTimes == 0 {
		ro.RetryTimes = defaultLockRetryTimes
	}
	if ro.SleepTimeExpandStep == 0 {
		ro.SleepTimeExpandStep = defaultSleepTimeExpandStep
	}
	if ro.SleepTimeFloat == 0 {
		ro.SleepTimeFloat = defaultSleepTimeFloat
	}
}

type RedisMutex interface {
	Lock(ctx context.Context, key string, uniqueValue int) error
	Unlock(ctx context.Context, key string, uniqueValue int) error
	Safely(ctx context.Context, key string, handler HandlerFunc) error
}

type redisMutex struct {
	client    gdb.RedisClient
	logger    *zap.Logger
	opt       *RedisMuxOption
	delayTime func(tryTimes int) time.Duration
	delScript *gdb.Script
}

func NewRedisMux(ctx context.Context, client gdb.RedisClient, opt *RedisMuxOption, logger *zap.Logger, tracer opentracing.Tracer) (RedisMutex,
	error) {
	if opt == nil {
		opt = &RedisMuxOption{}
	}
	opt.init()

	// new Script
	s, err := gdb.NewScript(ctx, client, unlockScript)
	if err != nil {
		return nil, err
	}

	return &redisMutex{
		client: client,
		logger: logger,
		opt:    opt,
		delayTime: func(tryTimes int) time.Duration {
			return opt.SleepTimeExpandStep*time.Duration(tryTimes+1) +
				opt.SleepTimeFloat*time.Duration(rand.Intn(10))/10
		},
		delScript: s,
	}, nil
}

type HandlerFunc func() error

func (rm *redisMutex) Lock(ctx context.Context, key string, uniqueValue int) error {
	i := 0
	var timer *time.Timer
	lockKey := redisMuxPrefix + key
	for ; i < rm.opt.RetryTimes; i++ {
		if i != 0 {
			if timer == nil {
				timer = time.NewTimer(rm.delayTime(i))
				defer timer.Stop() // nolint
			} else {
				timer.Reset(rm.delayTime(i))
			}
			select {
			case <-ctx.Done():
			case <-timer.C:
			}
		}
		ok, err := rm.client.SetNX(ctx, lockKey, uniqueValue, rm.opt.Expiration)
		if ok {
			return nil
		}
		if err != nil {
			rm.logger.Error("RedisMux lock failed with err", zap.String("key", lockKey), zap.Error(err))
		}
	}
	if i >= rm.opt.RetryTimes {
		rm.logger.Error("RedisMux lock failed", zap.String("key", lockKey))
	}
	return errRedisLockBusy
}

func (rm *redisMutex) Unlock(ctx context.Context, key string, uniqueValue int) error {
	lockKey := redisMuxPrefix + key
	err := rm.client.EvalSha(ctx, rm.delScript, []string{lockKey}, uniqueValue).Err()
	if err != nil {
		rm.logger.Error("RedisMux unlock failed with err", zap.String("key", key), zap.Error(err))
	}
	return err
}

func (rm *redisMutex) Safely(ctx context.Context, key string, handler HandlerFunc) error {
	var (
		now      = time.Now()
		funcCost time.Duration
	)
	defer func() {
		useMS := time.Since(now).Milliseconds()
		if useMS > 1000 {
			rm.logger.Warn("RedisMux lock slow", zap.String("key", key),
				zap.Int64("use millisecond", useMS),
				zap.Int64("handler exec millisecond", funcCost.Milliseconds()))
		}
	}()
	uniqueValue := int(time.Now().UnixNano()%10000000000*10000) + rand.Intn(10000)
	err := rm.Lock(ctx, key, uniqueValue)
	if err != nil {
		return errors.New("redis mutex lock failed: " + key)
	}
	funcStart := time.Now()
	err = handler()
	funcCost = time.Since(funcStart)
	return rm.Unlock(ctx, key, uniqueValue)
}
