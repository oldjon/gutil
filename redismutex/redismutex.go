package grmux

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/oldjon/gutil/env"
	"github.com/oldjon/gutil/gdb"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

const (
	redisMuxPrefix             = "rmux:"
	defaultExpiration          = 8000 * time.Millisecond // ms
	defaultLockRetryTimes      = 30
	defaultSleepTimeExpandStep = 10 * time.Millisecond // ms
	defaultSleepTimeFloat      = 10 * time.Millisecond // ms
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

type RedisMutex struct {
	client gdb.RedisClient
	logger *zap.Logger
	opt    *RedisMuxOption
}

func NewRedisMux(cfg env.ModuleConfig, opt *RedisMuxOption, logger *zap.Logger, tracer opentracing.Tracer) (*RedisMutex, error) {
	client, err := gdb.NewRedisClientByConfig(cfg, "", tracer)
	if err != nil {
		return nil, err
	}
	if opt == nil {
		opt = &RedisMuxOption{}
	}
	opt.init()
	return &RedisMutex{
		client: client,
		logger: logger,
		opt:    opt,
	}, nil
}

type HandlerFunc func() error

func (rm *RedisMutex) Lock(ctx context.Context, key string) bool {
	i := 0
	lockKey := redisMuxPrefix + key
	for ; i < 25; i++ {
		ok, err := rm.client.SetEXNX(ctx, lockKey, 1, rm.opt.Expiration)
		if ok {
			return true
		}
		if err != nil {
			rm.logger.Error("RedisMux lock failed with err", zap.String("key", lockKey), zap.Error(err))
		}
		time.Sleep(rm.opt.SleepTimeExpandStep*time.Duration(i+1) + rm.opt.SleepTimeFloat*time.Duration(rand.Intn(10))/10)
	}
	if i >= 25 {
		rm.logger.Error("RedisMux lock failed", zap.String("key", lockKey))
	}
	return false
}

func (rm *RedisMutex) Unlock(ctx context.Context, key string) {
	_, _ = rm.client.Del(ctx, redisMuxPrefix+key)
}

func (rm *RedisMutex) Safely(ctx context.Context, key string, handler HandlerFunc) error {
	var (
		now      = time.Now()
		funcCost time.Duration
	)
	defer func() {
		useMS := time.Since(now).Milliseconds()
		if useMS > 1000 {
			rm.logger.Warn("RedisMux lock slow", zap.String("key", key), zap.Int64("use millisecond", useMS),
				zap.Int64("handler exec millisecond", funcCost.Milliseconds()))
		}
	}()
	if !rm.Lock(ctx, key) {
		return errors.New("redis mutex lock failed: " + key)
	}
	funcStart := time.Now()
	err := handler()
	funcCost = time.Since(funcStart)
	if funcCost < rm.opt.Expiration-3*time.Millisecond {
		rm.Unlock(ctx, key)
	} // else expire itself
	return err
}
