package gdb

import (
	"errors"
)

var (
	ErrZAddKVLengthNotMatch  = errors.New("ERR_ZADD_KV_LENGTH_NOT_MATCH")
	ErrKeyValueCountDismatch = errors.New("ERR_KEY_VALUE_COUNT_DISMATCH")
	ErrKeyIsMissing          = errors.New("ERR_KEY_IS_MISSING")
)
