package fredis

import (
	"context"
	"fmt"
	"time"
)

type Generic interface {
	Exists(ctx context.Context, key string) (bool, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	Del(ctx context.Context, key string) (uint32, error)
}

func (rc *redisClient) Del(ctx context.Context, key string) (uint32, error) {
	ret := rc.client.Del(ctx, key)
	num, err := ret.Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get del result, %w", err)
	}
	return uint32(num), nil
}

// TTL returns time.Duration
func (rc *redisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	ret := rc.client.TTL(ctx, key)
	duration, err := ret.Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get ttl result, %w", err)
	}

	// run a check on duration
	// In Redis 2.6 or older the command returns -1 if the key does not exist or if the key exist but has no associated expire.
	// Starting with Redis 2.8 the return value in case of error changed:
	//   The command returns -2 if the key does not exist.
	//   The command returns -1 if the key exists but has no associated expire.
	if duration == ttlKeyNotExpireSet {
		return 0, ErrTTLKeyNotExpireSet
	} else if duration == ttlKeyNotExists {
		return 0, ErrTTLKeyNotExist
	}
	return duration, nil
}

func (rc *redisClient) Exists(ctx context.Context, key string) (bool, error) {
	ret := rc.client.Exists(ctx, key)
	n, err := ret.Result()
	return n == 1, err
}
