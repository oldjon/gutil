package gdb

// Funcs handle the redis data type zset

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type SortedSet interface {
	ZAdd(ctx context.Context, key string, values ...any) (int64, error)
	ZCard(ctx context.Context, key string) (int64, error)
	ZCount(ctx context.Context, key string, min, max string) (int64, error)
	ZIncrBy(ctx context.Context, key string, increment float64, member any) (float64, error)
	ZRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZRangeByScore(ctx context.Context, key string, min, max string) ([]string, error)
	ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error)
	ZRangeByScoreWithScores(ctx context.Context, key string, min, max string) ([]redis.Z, error)
	ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZRevRangeByScore(ctx context.Context, key string, min, max string) ([]string, error)
	ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error)
	ZRevRangeByScoreWithScores(ctx context.Context, key string, min, max string) ([]redis.Z, error)
	ZRank(ctx context.Context, key string, member any) (int64, error)
	ZRevRank(ctx context.Context, key string, member any) (int64, error)
	ZRem(ctx context.Context, key string, members ...any) (int64, error)
	ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error)
	ZRemRangeByScore(ctx context.Context, key string, min, max string) (int64, error)
	ZScore(ctx context.Context, key string, member any) (float64, error)
}

func (rc *redisClient) ZAdd(ctx context.Context, key string, values ...any) (int64, error) {
	if len(values)%2 != 0 {
		panic(PanicScoreValueCountUnmatched)
	}
	var members = make([]*redis.Z, 0, len(values))
	for i := 0; i < len(values); i += 2 {
		s, err := toFloat64(values[i])
		if err != nil {
			return 0, err
		}
		members = append(members, &redis.Z{
			Score:  s,
			Member: values[i+1],
		})
	}
	return rc.client.ZAdd(ctx, key, members...).Result()
}

func (rc *redisClient) ZCard(ctx context.Context, key string) (int64, error) {
	return rc.client.ZCard(ctx, key).Result()
}

func (rc *redisClient) ZCount(ctx context.Context, key string, min, max string) (int64, error) {
	return rc.client.ZCount(ctx, key, min, max).Result()
}

func (rc *redisClient) ZIncrBy(ctx context.Context, key string, increment float64, member any) (float64, error) {
	memStr, _ := toString(member)
	return rc.client.ZIncrBy(ctx, key, increment, memStr).Result()
}

func (rc *redisClient) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return rc.client.ZRange(ctx, key, start, stop).Result()
}

func (rc *redisClient) ZRangeByScore(ctx context.Context, key string, min, max string) ([]string, error) {
	return rc.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{Min: min, Max: max}).Result()
}

func (rc *redisClient) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return rc.client.ZRangeWithScores(ctx, key, start, stop).Result()
}

func (rc *redisClient) ZRangeByScoreWithScores(ctx context.Context, key string, min, max string) ([]redis.Z, error) {
	return rc.client.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{Min: min, Max: max}).Result()
}

func (rc *redisClient) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return rc.client.ZRevRange(ctx, key, start, stop).Result()
}

func (rc *redisClient) ZRevRangeByScore(ctx context.Context, key string, min, max string) ([]string, error) {
	return rc.client.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{Min: min, Max: max}).Result()
}

func (rc *redisClient) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return rc.client.ZRevRangeWithScores(ctx, key, start, stop).Result()
}

func (rc *redisClient) ZRevRangeByScoreWithScores(ctx context.Context, key string, min, max string) ([]redis.Z, error) {
	return rc.client.ZRevRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{Min: min, Max: max}).Result()
}

func (rc *redisClient) ZRank(ctx context.Context, key string, member any) (int64, error) {
	memStr, err := toString(member)
	if err != nil {
		return 0, err
	}
	return rc.client.ZRank(ctx, key, memStr).Result()
}

func (rc *redisClient) ZRevRank(ctx context.Context, key string, member any) (int64, error) {
	memStr, err := toString(member)
	if err != nil {
		return 0, err
	}
	return rc.client.ZRevRank(ctx, key, memStr).Result()
}

func (rc *redisClient) ZRem(ctx context.Context, key string, members ...any) (int64, error) {
	return rc.client.ZRem(ctx, key, members...).Result()
}

func (rc *redisClient) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error) {
	return rc.client.ZRemRangeByRank(ctx, key, start, stop).Result()
}

func (rc *redisClient) ZRemRangeByScore(ctx context.Context, key string, min, max string) (int64, error) {
	return rc.client.ZRemRangeByScore(ctx, key, min, max).Result()
}

func (rc *redisClient) ZScore(ctx context.Context, key string, member any) (float64, error) {
	memStr, err := toString(member)
	if err != nil {
		return 0, err
	}
	return rc.client.ZScore(ctx, key, memStr).Result()
}
