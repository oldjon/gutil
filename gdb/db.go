package gdb

import (
	"context"
	"time"
)

type DB struct {
	ObjectDB // support redis db or mongo db(TODO)
	RedisClient
}

func NewDB(redisClient RedisClient) *DB {
	return &DB{
		ObjectDB:    redisClient,
		RedisClient: redisClient,
	}
}

func (db *DB) GetObject(ctx context.Context, key string, obj any) error {
	return db.ObjectDB.GetObject(ctx, key, obj)
}

func (db *DB) SetObject(ctx context.Context, key string, obj any) error {
	return db.ObjectDB.SetObject(ctx, key, obj)
}

func (db *DB) SetObjectEX(ctx context.Context, key string, obj any, expiration time.Duration) error {
	return db.ObjectDB.SetObjectEX(ctx, key, obj, expiration)
}

func (db *DB) GetObjects(ctx context.Context, keys []string, objs any) error {
	return db.ObjectDB.GetObjects(ctx, keys, objs)
}

func (db *DB) SetObjects(ctx context.Context, keys []string, objs any) error {
	return db.ObjectDB.SetObjects(ctx, keys, objs)
}

func (db *DB) SetObjectsEX(ctx context.Context, keys []string, objs any, expiration time.Duration) error {
	return db.ObjectDB.SetObjectsEX(ctx, keys, objs, expiration)
}

func (db *DB) HSetObjects(ctx context.Context, key string, values ...any) error {
	return db.ObjectDB.HSetObjects(ctx, key, values...)
}

func (db *DB) HGetObject(ctx context.Context, key string, field string, obj any) error {
	return db.ObjectDB.HGetObject(ctx, key, field, obj)
}

func (db *DB) HMGetObjects(ctx context.Context, key string, fields []string, objs any) error {
	return db.ObjectDB.HMGetObjects(ctx, key, fields, objs)
}

func (db *DB) HGetAllObjects(ctx context.Context, key string, fields *[]string, objs any) error {
	return db.ObjectDB.HGetAllObjects(ctx, key, fields, objs)
}

func (db *DB) ZAddObjects(ctx context.Context, key string, values ...any) (int64, error) {
	return db.ObjectDB.ZAddObjects(ctx, key, values...)
}

func (db *DB) ZRangeObjects(ctx context.Context, key string, start, stop int64, objs any) error {
	return db.ObjectDB.ZRangeObjects(ctx, key, start, stop, objs)
}

func (db *DB) ZRangeObjectsByScore(ctx context.Context, key string, min, max string, objs any) error {
	return db.ObjectDB.ZRangeObjectsByScore(ctx, key, min, max, objs)
}

func (db *DB) ZRangeObjectsWithScores(ctx context.Context, key string, start, stop int64, objs any) (scores []float64, err error) {
	return db.ObjectDB.ZRangeObjectsWithScores(ctx, key, start, stop, objs)
}

func (db *DB) ZRangeObjectsByScoreWithScores(ctx context.Context, key string, min, max string, objs any) (scores []float64, err error) {
	return db.ObjectDB.ZRangeObjectsByScoreWithScores(ctx, key, min, max, objs)
}

func (db *DB) ZRevRangeObjects(ctx context.Context, key string, start, stop int64, objs any) error {
	return db.ObjectDB.ZRevRangeObjects(ctx, key, start, stop, objs)
}

func (db *DB) ZRevRangeObjectsByScore(ctx context.Context, key string, min, max string, objs any) error {
	return db.ObjectDB.ZRevRangeObjectsByScore(ctx, key, min, max, objs)
}

func (db *DB) ZRevRangeObjectsWithScores(ctx context.Context, key string, start, stop int64, objs any) (scores []float64, err error) {
	return db.ObjectDB.ZRevRangeObjectsWithScores(ctx, key, start, stop, objs)
}

func (db *DB) ZRevRangeObjectsByScoreWithScores(ctx context.Context, key string, min, max string, objs any) (scores []float64, err error) {
	return db.ObjectDB.ZRevRangeObjectsByScoreWithScores(ctx, key, min, max, objs)
}

func (db *DB) ZRankObject(ctx context.Context, key string, member any) (int64, error) {
	return db.ObjectDB.ZRankObject(ctx, key, member)
}

func (db *DB) ZRevRankObject(ctx context.Context, key string, member any) (int64, error) {
	return db.ObjectDB.ZRevRankObject(ctx, key, member)
}

func (db *DB) ZScoreObject(ctx context.Context, key string, member any) (float64, error) {
	return db.ObjectDB.ZScoreObject(ctx, key, member)
}

func (db *DB) ZRemObjects(ctx context.Context, key string, values ...any) (int64, error) {
	return db.ObjectDB.ZRemObjects(ctx, key, values...)
}

func (db *DB) IsErrNil(err error) bool {
	return db.RedisClient.IsErrNil(err)
}
