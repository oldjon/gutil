package gdb

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type PipelinerObject interface {
	GetObject(ctx context.Context, key string, obj any) error
	HSetObjects(ctx context.Context, key string, values ...any) error
	ZAddObjects(ctx context.Context, key string, values ...any) error
	ZRemObjects(ctx context.Context, key string, values ...any) error
}

// pipeline string object
func (pipe *pipeline) GetObject(ctx context.Context, key string, obj any) error {
	if obj == nil {
		panic(PanicValueDstNeedBeAllocated)
	}
	pipe.mux.Lock()
	result := pipe.Pipe.Get(ctx, key)
	pipe.resHandlers = append(pipe.resHandlers, func(cmd redis.Cmder, obj any) error {
		c := cmd.(*redis.StringCmd) // let it panic
		if c.Err() != nil {
			return nil
		}
		return pipe.rc.objMarshaller.Unmarshal([]byte(c.Val()), obj)
	})
	pipe.objects = append(pipe.objects, obj)
	pipe.mux.Unlock()
	return result.Err()
}

// pipeline hash object
func (pipe *pipeline) HSetObjects(ctx context.Context, key string, values ...any) error {
	var l = make([]any, 0, len(values))
	if len(values)%2 != 0 {
		panic(PanicFieldValueCountUnmatched)
	}
	for i := 0; i < len(values); i += 2 {
		l = append(l, values[i]) // key
		bys, err := pipe.rc.objMarshaller.Marshal(values[i+1])
		if err != nil {
			return err
		}
		l = append(l, bys) // value
	}
	pipe.mux.Lock()
	result := pipe.Pipe.HSet(ctx, key, l...)
	pipe.resHandlers = append(pipe.resHandlers, doNothing)
	pipe.objects = append(pipe.objects, nil)
	pipe.mux.Unlock()
	return result.Err()
}

// pipeline zset object
func (pipe *pipeline) ZAddObjects(ctx context.Context, key string, values ...any) error {
	if len(values)%2 != 0 {
		panic(PanicScoreValueCountUnmatched)
	}
	var members = make([]*redis.Z, 0, len(values))
	for i := 0; i < len(values); i += 2 {
		s, err := toFloat64(values[i])
		if err != nil {
			return err
		}
		bys, err := pipe.rc.objMarshaller.Marshal(values[i+1])
		if err != nil {
			return err
		}
		members = append(members, &redis.Z{
			Score:  s,
			Member: string(bys),
		})
	}
	pipe.mux.Lock()
	result := pipe.Pipe.ZAdd(ctx, key, members...)
	pipe.resHandlers = append(pipe.resHandlers, doNothing)
	pipe.objects = append(pipe.objects, nil)
	pipe.mux.Unlock()
	return result.Err()
}

func (pipe *pipeline) ZRemObjects(ctx context.Context, key string, values ...any) error {
	for i, v := range values {
		bys, err := pipe.rc.objMarshaller.Marshal(v)
		if err != nil {
			return err
		}
		values[i] = string(bys)
	}
	pipe.mux.Lock()
	result := pipe.Pipe.ZRem(ctx, key, values...)
	pipe.resHandlers = append(pipe.resHandlers, doNothing)
	pipe.objects = append(pipe.objects, nil)
	pipe.mux.Unlock()
	return result.Err()
}
