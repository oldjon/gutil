package gdb

import "context"

type Hash interface {
	HGet(ctx context.Context, key, field string) (string, error)
	HSet(ctx context.Context, key string, values ...any) error
	HMSet(ctx context.Context, key string, values ...any) (bool, error)
	HMGet(ctx context.Context, key string, fields ...string) ([]any, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDel(ctx context.Context, key string, fields ...string) (int64, error)
	HLen(ctx context.Context, key string) (int64, error)
}

func (rc *redisClient) HSet(ctx context.Context, key string, values ...any) error {
	return rc.client.HSet(ctx, key, values...).Err()
}

func (rc *redisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return rc.client.HGet(ctx, key, field).Result()
}

func (rc *redisClient) HMSet(ctx context.Context, key string, values ...any) (bool, error) {
	return rc.client.HMSet(ctx, key, values...).Result()
}

func (rc *redisClient) HMGet(ctx context.Context, key string, fields ...string) ([]any, error) {
	return rc.client.HMGet(ctx, key, fields...).Result()
}

func (rc *redisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return rc.client.HGetAll(ctx, key).Result()
}

func (rc *redisClient) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	return rc.client.HDel(ctx, key, fields...).Result()
}

func (rc *redisClient) HLen(ctx context.Context, key string) (int64, error) {
	return rc.client.HDel(ctx, key).Result()
}
