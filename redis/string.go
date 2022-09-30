package fredis

import (
	"context"
	"fmt"
	"time"
)

type String interface {
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
}

func (rc *redisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	ret := rc.client.SetNX(ctx, key, value, expiration)
	ok, err := ret.Result()
	if err != nil {
		return false, fmt.Errorf("failed to get setnx result, %w", err)
	}
	return ok, nil
}
