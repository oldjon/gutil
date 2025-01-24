package gdb

import (
	gmarshaller "github.com/oldjon/gutil/marshaller"
	"testing"
)

func newRedisClient(t *testing.T) RedisClient {
	addr := "192.168.221.129:6001"
	op := &RedisClientOption{
		Addr:       addr,
		Marshaller: &gmarshaller.JsonMarshaller{},
	}
	client, err := NewRedisClient(op)
	if err != nil {
		t.Errorf("newRedisClient %v", err)
		t.Failed()
	}
	return client
}
