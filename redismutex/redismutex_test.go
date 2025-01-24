package grmux

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/oldjon/gutil/gdb"
	gmarshaller "github.com/oldjon/gutil/marshaller"
	"sync"
	"testing"
)

func newRedisClient(t *testing.T) gdb.RedisClient {
	addr := "192.168.221.129:6001"
	op := &gdb.RedisClientOption{
		Addr:       addr,
		Marshaller: &gmarshaller.JsonMarshaller{},
	}
	client, err := gdb.NewRedisClient(op)
	if err != nil {
		t.Errorf("newRedisClient %v", err)
		t.Failed()
	}
	return client
}

func newTestMutex(t *testing.T) (*RedisMutex, error) {
	client := newRedisClient(t)
	if client == nil {
		t.Errorf("newRedisClient failed")
		return nil, nil
	}
	opt := &RedisMuxOption{}
	opt.init()

	return &RedisMutex{
		client:      client,
		opt:         opt,
		delScripter: redis.NewScript(delScript),
	}, nil
}

func TestRedisMutex(t *testing.T) {
	client := newRedisClient(t)
	if client == nil {
		t.Errorf("newRedisClient failed")
		return
	}
	totalNum := 20

	rMux, err := newTestMutex(t)
	if err != nil {
		t.Errorf(fmt.Sprintf("TestRedisMutex new mutex failed %v", err))
		t.Failed()
		return
	}
	ctx := context.Background()

	wg := &sync.WaitGroup{}
	wg.Add(totalNum * 2)
	key1 := "test_with_mux"
	key2 := "test_no_mux"

	_, _ = client.Del(ctx, key1)
	_, _ = client.Del(ctx, key2)

	fWithMux := func(wg *sync.WaitGroup) {
		err = rMux.Safely(ctx, key1, func() error {
			n, err := gdb.ToUint64(client.Get(ctx, key1))
			if err != nil && !errors.Is(err, redis.Nil) {
				return err
			}
			n = n + 1
			return client.Set(ctx, key1, n)
		})
		if err != nil {
			t.Errorf(fmt.Sprintf("TestRedisMutex new mutex failed %v", err))
			t.Failed()
			return
		}
		wg.Done()
	}

	fNoMux := func(wg *sync.WaitGroup) {
		n, err := gdb.ToUint64(client.Get(ctx, key2))
		if err != nil && !errors.Is(err, redis.Nil) {
			return
		}
		n = n + 1
		_ = client.Set(ctx, key2, n)
		wg.Done()
	}

	for i := 0; i < totalNum; i++ {
		go fWithMux(wg)
		go fNoMux(wg)
	}

	wg.Wait()

	nWithMux, err := gdb.ToUint64(client.Get(ctx, key1))
	if err != nil {
		t.Errorf(fmt.Sprintf("TestRedisMutex get test_with_mux failed %v", err))
		t.Failed()
		return
	}
	nNoMux, err := gdb.ToUint64(client.Get(ctx, key2))
	if err != nil {
		t.Errorf(fmt.Sprintf("TestRedisMutex get test_no_mux failed %v", err))
		t.Failed()
		return
	}

	fmt.Println(fmt.Sprintf("with mux, need %d, got %d", totalNum, nWithMux))
	fmt.Println(fmt.Sprintf("without mux, need %d, got %d", totalNum, nNoMux))
	if nWithMux != uint64(totalNum) {
		t.Errorf(fmt.Sprintf("TestRedisMutex redis mux failed %v", nWithMux))
		t.Failed()
	}
}
