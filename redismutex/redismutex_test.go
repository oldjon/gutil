package grmux

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/oldjon/gutil/conv"
	"github.com/oldjon/gutil/gdb"
	gmarshaller "github.com/oldjon/gutil/marshaller"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"path/filepath"
	"sync"
	"testing"
)

var addr = "192.168.221.129:6001"

func newLogger() *zap.Logger {
	level := zapcore.DebugLevel

	opts := []zap.Option{
		zap.Development(),
		zap.AddCaller(),
		zap.AddStacktrace(zap.WarnLevel),
	}
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	absPath, err := filepath.Abs("./test.log")
	if err != nil {
		panic(err)
		return nil
	}
	logLevel := zap.NewAtomicLevelAt(level)
	rotateWriter := &lumberjack.Logger{
		Filename:  absPath,
		MaxSize:   10, // MB
		LocalTime: true,
	}
	writer := zapcore.AddSync(rotateWriter)

	wCore := zapcore.NewCore(encoder, writer, logLevel)
	core := zapcore.NewTee(wCore)
	if core == nil {
		return nil
	}

	logger := zap.New(core, opts...)
	return logger
}

func newRedisClient() gdb.RedisClient {
	opt := &gdb.RedisClientOption{
		Mode:       gdb.Single,
		Addr:       addr,
		Marshaller: &gmarshaller.JsonMarshaller{},
	}
	client, err := gdb.NewRedisClient(opt)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return client
}

func newRedisMutex() RedisMutex {
	client := newRedisClient()
	if client == nil {
		fmt.Println("new redis client failed")
		return nil
	}
	logger := newLogger()
	if logger == nil {
		fmt.Println("new zap.Logger failed")
		return nil
	}
	rm, err := NewRedisMux(context.Background(), client, nil, nil, nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return rm
}

func TestRedisMux(t *testing.T) {
	rm := newRedisMutex()
	if rm == nil {
		t.Fatal("newRedisMutex failed")
		return
	}
	rc := newRedisClient()
	if rc == nil {
		t.Fatal("newRedisClient failed")
		return
	}

	wg := new(sync.WaitGroup)
	cnt := 30
	wg.Add(cnt)
	ctx, cancel := context.WithCancel(context.Background())
	rc.Del(context.Background(), "test")

	defer cancel()
	f := func(ctx context.Context, wg *sync.WaitGroup, rm RedisMutex, rc gdb.RedisClient) {
		err := rm.Safely(ctx, "test", func() error {
			v, err := rc.Get(ctx, "test")
			if err != nil && !errors.Is(err, redis.Nil) {
				return err
			}
			vi := conv.StringToUint64(v) + 1
			return rc.Set(ctx, "test", vi)
		})
		if err != nil {
			fmt.Println(err)
		}
		wg.Done()
	}

	for i := 0; i < cnt; i++ {
		go f(ctx, wg, rm, rc)
	}

	wg.Wait()
	v, err := rc.Get(ctx, "test")
	if err != nil {
		t.Fatal(err)
		return
	}
	if conv.StringToInt64(v) != int64(cnt) {
		t.Fatal("result is wrong")
		return
	}
}

func TestClient(t *testing.T) {
	rc := newRedisClient()
	if rc == nil {
		t.Fatal()
		return
	}
	rc.SetNX(context.Background(), "casda111", 1, 0)
}
