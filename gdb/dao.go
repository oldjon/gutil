package gdb

import (
	"context"
	"time"
)

type DB struct {
	ObjectDB // support redis db or mongo db(TODO)
	RedisClient
}

func NewDAO(objDB ObjectDB, redisClient RedisClient) *DB {
	return &DB{
		ObjectDB:    objDB,
		RedisClient: redisClient,
	}
}

func (db *DB) GetObject(ctx context.Context, key string, obj interface{}) error {
	return db.ObjectDB.GetObject(ctx, key, obj)
}

func (db *DB) SetObject(ctx context.Context, key string, obj interface{}) error {
	return db.ObjectDB.SetObject(ctx, key, obj)
}

func (db *DB) SetObjectEx(ctx context.Context, key string, obj interface{}, expiration time.Duration) error {
	return db.ObjectDB.SetObjectEx(ctx, key, obj, expiration)
}

func (db *DB) GetObjects(ctx context.Context, keys []string, objs []interface{}) error {
	return db.ObjectDB.GetObjects(ctx, keys, objs)
}

func (db *DB) SetObjects(ctx context.Context, keys []string, objs []interface{}) error {
	return db.ObjectDB.SetObjects(ctx, keys, objs)
}

func (db *DB) SetObjectsEx(ctx context.Context, keys []string, objs []interface{}, expiration time.Duration) error {
	return db.ObjectDB.SetObjectsEx(ctx, keys, objs, expiration)
}

func (db *DB) IsErrNil(err error) bool {
	return db.ObjectDB.IsErrNil(err)
}
