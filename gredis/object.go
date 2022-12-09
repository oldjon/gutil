package gredis

import (
	"context"
	"time"
)

type Object interface {
	GetObject(ctx context.Context, key string, obj interface{}) error
	SetObject(ctx context.Context, key string, obj interface{}) error
	SetObjectEx(ctx context.Context, key string, obj interface{}, expiration time.Duration) error
}

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
