package gdb

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type SortedSet interface {
	ZAdd(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
}

func (rc *redisClient) ZAdd(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	ret := rc.client.ZAdd(ctx, key, members...)
	n, err := ret.Result()
	return n, err
}
