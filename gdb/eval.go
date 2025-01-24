package gdb

import (
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

type Scripter interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)
	EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) (interface{}, error)
	ScriptLoad(ctx context.Context, script string) (string, error)
	RunScript(ctx context.Context, script *redis.Script, keys []string, args ...interface{}) (interface{}, error)
}

func (rc *redisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	cmd := rc.client.Eval(ctx, script, keys, args)
	return cmd.Val(), cmd.Err()
}

func (rc *redisClient) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	cmd := rc.client.EvalSha(ctx, sha1, keys, args)
	return cmd.Val(), cmd.Err()
}

func (rc *redisClient) ScriptLoad(ctx context.Context, script string) (string, error) {
	cmd := rc.client.ScriptLoad(ctx, script)
	return cmd.Val(), cmd.Err()
}

func (rc *redisClient) RunScript(ctx context.Context, script *redis.Script, keys []string, args ...interface{},
) (interface{}, error) {
	cmd := script.Run(ctx, rc.client, keys, args...)
	//r := script.EvalSha(ctx, rc.client, keys, args...)
	//if err := r.Err(); err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT ") {
	//	fmt.Println("NOSCRIPT")
	//	r = script.Eval(ctx, rc.client, keys, args...)
	//	return r, r.Err()
	//}
	return cmd.Val(), cmd.Err()
}
