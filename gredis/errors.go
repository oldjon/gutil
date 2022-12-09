package gredis

import (
	"errors"

	"github.com/go-redis/redis/v8"
)

var (
	ErrNil                  = redis.Nil
	ErrZAddKVLengthNotMatch = errors.New("ERR_ZADD_KV_LENGTH_NOT_MATCH")
)
