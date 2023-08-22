package gdb

// Funcs handle the redis data type string

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

type String interface {
	SetNX(ctx context.Context, key string, value any) (bool, error)
	SetEXNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error)
	Set(ctx context.Context, key string, value any) error
	SetEX(ctx context.Context, key string, value any, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Incr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	BatchGet(ctx context.Context, keys []string) ([]*redis.StringCmd, error)
	BatchSet(ctx context.Context, keys []string, values []any, expiration time.Duration) error
}

func (rc *redisClient) SetNX(ctx context.Context, key string, value any) (bool, error) {
	return rc.SetEXNX(ctx, key, value, 0)
}

func (rc *redisClient) SetEXNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	return rc.client.SetNX(ctx, key, value, expiration).Result()
}

func (rc *redisClient) Set(ctx context.Context, key string, value any) error {
	return rc.client.Set(ctx, key, value, -1).Err()
}

func (rc *redisClient) SetEX(ctx context.Context, key string, value any, expiration time.Duration) error {
	return rc.client.Set(ctx, key, value, expiration).Err()
}

func (rc *redisClient) Get(ctx context.Context, key string) (string, error) {
	return rc.client.Get(ctx, key).Result()
}

func (rc *redisClient) Incr(ctx context.Context, key string) (int64, error) {
	return rc.client.Incr(ctx, key).Result()
}

func (rc *redisClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return rc.client.IncrBy(ctx, key, value).Result()
}

func (rc *redisClient) BatchGet(ctx context.Context, keys []string) ([]*redis.StringCmd, error) {
	rets, err := rc.client.Pipelined(ctx, func(pipeliner redis.Pipeliner) error {
		for _, key := range keys {
			pipeliner.Get(ctx, key)
		}
		return nil
	})
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	cmds := make([]*redis.StringCmd, len(rets))
	for i, v := range rets {
		cmds[i] = v.(*redis.StringCmd)
	}
	return cmds, nil
}

func (rc *redisClient) BatchSet(ctx context.Context, keys []string, values []any, expiration time.Duration) error {
	if len(keys) != len(values) {
		panic(PanicKeyValueCountUnmatched)
	}
	_, err := rc.client.Pipelined(ctx, func(pipeliner redis.Pipeliner) error {
		for i, key := range keys {
			pipeliner.Set(ctx, key, values[i], expiration)
		}
		return nil
	})
	return err
}
