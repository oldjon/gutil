package gdb

import (
	"context"
	"fmt"
	"testing"
)

func TestEval(t *testing.T) {
	client := newRedisClient(t)
	if client == nil {
		t.Errorf("newRedisClient failed")
		return
	}
	ctx := context.Background()

	script := `
	local key = KEYS[1]
	local arg = ARGV[1]
	local result = redis.call('incrby', key, arg)
	return result
`

	ret, err := client.Eval(ctx, script, []string{"teateval"}, 1)
	if err != nil {
		t.Errorf("TestEval err: %v", err)
	}
	t.Log(fmt.Sprintf("TestEval result: %v", ret))
}
