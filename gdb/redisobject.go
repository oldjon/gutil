package gdb

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

func (rc *redisClient) GetObject(ctx context.Context, key string, obj interface{}) error {
	v, err := rc.Get(ctx, key)
	if err != nil {
		return err
	}
	return rc.objMarshaller.Unmarshal([]byte(v), obj)
}

func (rc *redisClient) SetObject(ctx context.Context, key string, obj interface{}) error {
	bys, err := rc.objMarshaller.Marshal(obj)
	if err != nil {
		return err
	}
	return rc.Set(ctx, key, bys)
}

func (rc *redisClient) SetObjectEx(ctx context.Context, key string, obj interface{}, expiration time.Duration) error {
	bys, err := rc.objMarshaller.Marshal(obj)
	if err != nil {
		return err
	}
	return rc.SetEx(ctx, key, bys, expiration)
}

func (rc *redisClient) IsErrNil(err error) bool {
	return errors.Is(err, redis.Nil)
}

func (rc *redisClient) GetObjects(ctx context.Context, keys []string, objs []interface{}) error {
	if len(keys) == 0 {
		return ErrKeyIsMissing
	}

	if len(keys) != len(objs) {
		return ErrKeyValueCountDismatch
	}

	cmds, err := rc.BatchGet(ctx, keys)
	if err != nil {
		return err
	}
	for i, v := range cmds {
		if rc.IsErrNil(v.Err()) {
			objs[i] = nil
		} else if v.Err() != nil {
			return err
		}
		err = rc.objMarshaller.Unmarshal([]byte(v.String()), objs[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (rc *redisClient) SetObjects(ctx context.Context, keys []string, objs []interface{}) error {
	if len(keys) == 0 {
		return ErrKeyIsMissing
	}
	if len(keys) != len(objs) {
		return ErrKeyValueCountDismatch
	}
	var err error
	datas := make([]interface{}, len(objs))
	for i, obj := range objs {
		datas[i], err = rc.objMarshaller.Marshal(obj)
		if err != nil {
			return err
		}
	}
	return rc.BatchSet(ctx, keys, datas, -1)
}

func (rc *redisClient) SetObjectsEx(ctx context.Context, keys []string, objs []interface{}, expiration time.Duration) error {
	if len(keys) == 0 {
		return ErrKeyIsMissing
	}
	if len(keys) != len(objs) {
		return ErrKeyValueCountDismatch
	}
	var err error
	datas := make([]interface{}, len(objs))
	for i, obj := range objs {
		datas[i], err = rc.objMarshaller.Marshal(obj)
		if err != nil {
			return err
		}
	}
	return rc.BatchSet(ctx, keys, datas, expiration)
}
