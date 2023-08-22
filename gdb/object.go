package gdb

import (
	"context"
	"time"
)

type ObjectDB interface {
	// GetObject get data from db of the key, and unmarshal into obj.
	// obj should be a struct point, and should be memory allocated
	GetObject(ctx context.Context, key string, obj any) error
	// SetObject set data into db by the key, the data is unmarshalled from obj.
	// obj should be a struct point, and not nil.
	SetObject(ctx context.Context, key string, obj any) error
	// SetObjectEX set data into db by the key with expiration, the data is unmarshalled from obj.
	// obj should be a struct point, and not nil.
	SetObjectEX(ctx context.Context, key string, obj any, expiration time.Duration) error
	// GetObjects get datas from db of all keys, and unmarshal into objs.
	// objs should be a slice of struct points, and slice should be memory allocated.
	GetObjects(ctx context.Context, keys []string, objs any) error
	// SetObjects set datas into db by keys, the datas is unmarshalled from objs.
	// objs should be a slice.
	SetObjects(ctx context.Context, keys []string, objs any) error
	// SetObjectsEX set datas into db by keys with expiration, the datas is unmarshalled from objs.
	// objs should be a slice.
	SetObjectsEX(ctx context.Context, keys []string, objs any, expiration time.Duration) error

	// HSetObjects set filed value pairs into db by the key, value will be marshalled before set.
	HSetObjects(ctx context.Context, key string, values ...any) error
	// HGetObject get data from db with the key and the field, and unmarshal into obj.
	// obj should be a struct point, and should be memory allocated
	HGetObject(ctx context.Context, key string, field string, obj any) error
	// HMGetObjects get datas from db with the key and the fields, and unmarshall into objs.
	// objs should be a slice of struct points, and the slice should be memory allocated.
	HMGetObjects(ctx context.Context, key string, fields []string, objs any) error
	// HGetAllObjects get datas from db with the key and the fields, and unmarshall into objs,
	// field_name will be set into fields, objs should be a point of a slice of same struct or struct points.
	HGetAllObjects(ctx context.Context, key string, fields *[]string, objs any) error

	// ZAddObjects add score and member pairs into db by the key. score should be a number,
	// member will be marshalled before set.
	ZAddObjects(ctx context.Context, key string, values ...any) (int64, error)
	// ZRangeObjects ZRange members from zset of the key, and unmarshall members into objs.
	// objs should be a point of a slice of struct points.
	ZRangeObjects(ctx context.Context, key string, start, stop int64, objs any) error
	// ZRangeObjectsByScore ZRange members from zset of the key, and unmarshall members into objs.
	// objs should be a point of a slice of struct points.
	ZRangeObjectsByScore(ctx context.Context, key string, min, max string, objs any) error
	// ZRangeObjectsWithScores ZRange members with scores from zset of the key, and unmarshall members into objs.
	// objs should be a point of a slice of struct or struct points.
	ZRangeObjectsWithScores(ctx context.Context, key string, start, stop int64, objs any) (scores []float64, err error)
	// ZRangeObjectsByScoreWithScores ZRange members with scores from zset of the key, and unmarshall members into objs.
	// objs should be a point of a slice of struct or struct points.
	ZRangeObjectsByScoreWithScores(ctx context.Context, key string, min, max string, objs any) (scores []float64, err error)
	// ZRevRangeObjects ZRevRange members from zset of the key, and unmarshall members into objs.
	// objs should be a point of a slice of struct points.
	ZRevRangeObjects(ctx context.Context, key string, start, stop int64, objs any) error
	// ZRevRangeObjectsByScore ZRevRangeByScore members from zset of the key, and unmarshall members into objs.
	// objs should be a point of a slice of struct points.
	ZRevRangeObjectsByScore(ctx context.Context, key string, min, max string, objs any) error
	// ZRevRangeObjectsWithScores ZRevRange members with scores from zset of the key, and unmarshall members into objs.
	// objs should be a point of a slice of struct or struct points.
	ZRevRangeObjectsWithScores(ctx context.Context, key string, start, stop int64, objs any) (scores []float64, err error)
	// ZRevRangeObjectsByScoreWithScores ZRevRangeByScore members with scores from zset of the key, and unmarshall members into objs.
	// objs should be a point of a slice of struct or struct points.
	ZRevRangeObjectsByScoreWithScores(ctx context.Context, key string, min, max string, objs any) (scores []float64, err error)
	ZRankObject(ctx context.Context, key string, member any) (int64, error)
	ZRevRankObject(ctx context.Context, key string, member any) (int64, error)
	ZScoreObject(ctx context.Context, key string, member any) (float64, error)
	ZRemObjects(ctx context.Context, key string, members ...any) (int64, error)
}
