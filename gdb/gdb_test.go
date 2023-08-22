package gdb

import (
	gmarshaller "github.com/oldjon/gutil/marshaller"
	"testing"
)

func newRedisClient(t *testing.T) RedisClient {
	addr := "127.0.0.1:6101"
	op := &RedisClientOption{
		Addr:       addr,
		Marshaller: &gmarshaller.JsonMarshaller{},
	}
	client, err := NewRedisClient(op)
	if err != nil {
		t.Failed()
	}
	return client
}
