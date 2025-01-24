package grmux

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap/zapcore"
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
	defaultSleepTimeExpandStep = 20 * time.Millisecond // ms
	defaultSleepTimeFloat      = 20 * time.Millisecond // ms
)

const delScript = `
	local key = KEYS[1]
	local value = ARGV[1]
	if redis.call('get', key) == value 
	then
		return redis.call('del', key)
	else 
		return 0
	end
`

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
	client      gdb.RedisClient
	logger      *zap.Logger
	opt         *RedisMuxOption
	delScripter *redis.Script
}

func (rm *RedisMutex) Error(msg string, fields ...zapcore.Field) {
	if rm.logger == nil {
		fmt.Println(msg, fields)
		return
	}
	rm.logger.Error(msg, fields...)
}

func (rm *RedisMutex) Info(msg string, fields ...zapcore.Field) {
	if rm.logger == nil {
		fmt.Println(msg, fields)
		return
	}
	rm.logger.Info(msg, fields...)
}

func (rm *RedisMutex) Warn(msg string, fields ...zapcore.Field) {
	if rm.logger == nil {
		fmt.Println(msg, fields)
		return
	}
	rm.logger.Warn(msg, fields...)
}

func NewRedisMux(cfg env.ModuleConfig, opt *RedisMuxOption, logger *zap.Logger, tracer opentracing.Tracer,
) (*RedisMutex, error) {
	client, err := gdb.NewRedisClientByConfig(cfg, "", tracer)
	if err != nil {
		return nil, err
	}
	if opt == nil {
		opt = &RedisMuxOption{}
	}
	opt.init()

	return &RedisMutex{
		client:      client,
		logger:      logger,
		opt:         opt,
		delScripter: redis.NewScript(delScript),
	}, nil
}

type HandlerFunc func() error

func (rm *RedisMutex) Lock(ctx context.Context, key string, value uint64) bool {
	i := 0
	lockKey := redisMuxPrefix + key
	for ; i < 25; i++ {
		ok, err := rm.client.SetEXNX(ctx, lockKey, value, rm.opt.Expiration)
		if ok {
			return true
		}
		if err != nil {
			rm.Error("RedisMux lock failed with err", zap.String("key", lockKey), zap.Error(err))
		}
		time.Sleep(rm.opt.SleepTimeExpandStep*time.Duration(i+1) +
			rm.opt.SleepTimeFloat*time.Duration(rand.Intn(10))/10)
	}
	if i >= 25 {
		rm.Error("RedisMux lock failed", zap.String("key", lockKey))
	}
	return false
}

func (rm *RedisMutex) Unlock(ctx context.Context, key string, value uint64) bool {
	lockKey := redisMuxPrefix + key

	_, err := rm.client.RunScript(ctx, rm.delScripter, []string{lockKey}, value)
	if err != nil {
		rm.Error("RedisMutex Unlock failed", zap.String("key", lockKey), zap.Error(err))
		return false
	}
	return true
}

func (rm *RedisMutex) Safely(ctx context.Context, key string, handler HandlerFunc) error {
	var (
		now      = time.Now()
		funcCost time.Duration
	)
	defer func() {
		useMS := time.Since(now).Milliseconds()
		if useMS > 3000 {
			rm.Warn("RedisMux lock slow", zap.String("key", key),
				zap.Int64("use millisecond", useMS),
				zap.Int64("handler exec millisecond", funcCost.Milliseconds()))
		}
	}()
	var value = uint64(time.Now().UnixMicro()*10000) + uint64(rand.Intn(10000))

	if !rm.Lock(ctx, key, value) {
		return errors.New("redis mutex lock failed: " + key)
	}

	defer rm.Unlock(ctx, key, value)

	funcStart := time.Now()
	err := handler()
	funcCost = time.Since(funcStart)
	return err
}
