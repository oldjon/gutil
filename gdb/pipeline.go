package gdb

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"
)

type Pipeliner interface {
	PipelinerObject
	Exec(ctx context.Context) ([]redis.Cmder, error)
	// pipeline zset
	ZRemRangeByScore(ctx context.Context, key string, min, max float64) error
}

func (rc *redisClient) Pipeline() Pipeliner {
	return &pipeline{Pipe: rc.client.Pipeline(), rc: rc}
}

func (rc *redisClient) TxPipeline() Pipeliner {
	return &pipeline{Pipe: rc.client.TxPipeline(), rc: rc}
}

func doNothing(_ redis.Cmder, _ any) error { return nil }

type pipeline struct {
	Pipe        redis.Pipeliner
	rc          *redisClient
	mux         sync.RWMutex
	resHandlers []func(cmd redis.Cmder, obj any) error
	objects     []any
}

func (pipe *pipeline) Exec(ctx context.Context) ([]redis.Cmder, error) {
	cmds, err := pipe.Pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	if len(pipe.resHandlers) == 0 {
		return cmds, nil
	}
	for i, f := range pipe.resHandlers {
		err = f(cmds[i], pipe.objects[i])
		if err != nil {
			return nil, err
		}
	}
	return cmds, nil
}

// pipeline zset
func (pipe *pipeline) ZRemRangeByScore(ctx context.Context, key string, min, max float64) error {
	minStr, err := toString(min)
	if err != nil {
		return err
	}
	maxStr, err := toString(max)
	if err != nil {
		return err
	}
	pipe.mux.Lock()
	result := pipe.Pipe.ZRemRangeByScore(ctx, key, minStr, maxStr)
	pipe.resHandlers = append(pipe.resHandlers, doNothing)
	pipe.objects = append(pipe.objects, nil)
	pipe.mux.Unlock()
	return result.Err()
}
