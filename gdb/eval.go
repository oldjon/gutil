package gdb

import "golang.org/x/net/context"

type Eval interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)
}

func (rc *redisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	cmd := rc.client.Eval(ctx, script, keys, args)
	return cmd.Val(), cmd.Err()
}
