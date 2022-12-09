package gdb

import (
	"context"
	"time"
)

type ObjectDB interface {
	GetObject(ctx context.Context, key string, obj interface{}) error
	SetObject(ctx context.Context, key string, obj interface{}) error
	GetObjects(ctx context.Context, keys []string, objs []interface{}) error
	SetObjects(ctx context.Context, keys []string, objs []interface{}, expiration time.Duration) error
	IsErrNil(err error) bool
}
