package gdb

import (
	"context"
	"strings"

	"github.com/go-redis/redis/v8"
)

type Scripter interface {
	Eval(ctx context.Context, s *Script, keys []string, args ...interface{}) *redis.Cmd
	EvalSha(ctx context.Context, s *Script, keys []string, args ...interface{}) *redis.Cmd
	ScriptExists(ctx context.Context, hashes ...string) ([]bool, error)
	ScriptFlush(ctx context.Context) (string, error)
	ScriptKill(ctx context.Context) (string, error)
	ScriptLoad(ctx context.Context, script string) (string, error)
}

type Script struct {
	src, hash string
}

func NewScript(ctx context.Context, c RedisClient, src string) (*Script, error) {
	hash, err := c.ScriptLoad(ctx, src)
	if err != nil {
		return nil, err
	}
	return &Script{src: src, hash: hash}, nil
}

func (rc *redisClient) Eval(ctx context.Context, s *Script, keys []string, args ...interface{}) *redis.Cmd {
	if s == nil {
		cmd := &redis.Cmd{}
		cmd.SetErr(ErrScriptIsNil)
		return cmd
	}
	return rc.client.Eval(ctx, s.src, keys, args...)
}

func (rc *redisClient) EvalSha(ctx context.Context, s *Script, keys []string, args ...interface{}) *redis.Cmd {
	if s == nil {
		cmd := &redis.Cmd{}
		cmd.SetErr(ErrScriptIsNil)
		return cmd
	}

	cmd := rc.client.EvalSha(ctx, s.hash, keys, args...)
	if cmd.Err() == nil {
		return cmd
	}
	if strings.HasPrefix(cmd.Err().Error(), "NOSCRIPT") {
		return rc.client.Eval(ctx, s.src, keys, args...)
	}
	return cmd
}

func (rc *redisClient) ScriptExists(ctx context.Context, hashes ...string) ([]bool, error) {
	return rc.client.ScriptExists(ctx, hashes...).Result()
}

func (rc *redisClient) ScriptFlush(ctx context.Context) (string, error) {
	return rc.client.ScriptFlush(ctx).Result()
}

func (rc *redisClient) ScriptKill(ctx context.Context) (string, error) {
	return rc.client.ScriptKill(ctx).Result()
}

func (rc *redisClient) ScriptLoad(ctx context.Context, script string) (string, error) {
	return rc.client.ScriptLoad(ctx, script).Result()
}
