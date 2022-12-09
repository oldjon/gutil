package gdb

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type String interface {
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Set(ctx context.Context, key string, value interface{}) error
	SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Incr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	BatchGet(ctx context.Context, keys []string) ([]*redis.StringCmd, error)
	BatchSet(ctx context.Context, keys []string, values []interface{}, expiration time.Duration) error
}

func (rc *redisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	ret := rc.client.SetNX(ctx, key, value, expiration)
	ok, err := ret.Result()
	return ok, err
}

func (rc *redisClient) Set(ctx context.Context, key string, value interface{}) error {
	ret := rc.client.Set(ctx, key, value, -1)
	_, err := ret.Result()
	return err
}

func (rc *redisClient) SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	ret := rc.client.Set(ctx, key, value, expiration)
	_, err := ret.Result()
	return err
}

func (rc *redisClient) Get(ctx context.Context, key string) (string, error) {
	ret := rc.client.Get(ctx, key)
	value, err := ret.Result()
	return value, err
}

func (rc *redisClient) Incr(ctx context.Context, key string) (int64, error) {
	ret := rc.client.Incr(ctx, key)
	value, err := ret.Result()
	return value, err
}

func (rc *redisClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	ret := rc.client.IncrBy(ctx, key, value)
	value, err := ret.Result()
	return value, err
}

func (rc *redisClient) BatchGet(ctx context.Context, keys []string) ([]*redis.StringCmd, error) {
	rets, err := rc.client.Pipelined(ctx, func(pipeliner redis.Pipeliner) error {
		for _, key := range keys {
			pipeliner.Get(ctx, key)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	cmds := make([]*redis.StringCmd, len(rets))
	for i, v := range rets {
		cmds[i] = v.(*redis.StringCmd)
	}
	return cmds, nil
}

func (rc *redisClient) BatchSet(ctx context.Context, keys []string, values []interface{}, expiration time.Duration) error {
	if len(keys) != len(values) {
		return ErrKeyValueCountDismatch
	}
	_, err := rc.client.Pipelined(ctx, func(pipeliner redis.Pipeliner) error {
		for i, key := range keys {
			pipeliner.Set(ctx, key, values[i], expiration)
		}
		return nil
	})
	return err
}
