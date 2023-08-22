package gdb

import (
	"context"
	"testing"
	"unsafe"
)

type Foo struct {
	F int
}

type Bar struct {
	B uint
}

func TestHSetObjects(t *testing.T) {
	client := newRedisClient(t)
	ctx := context.Background()

	err := client.HSetObjects(ctx, "TestHSetObjects", "f1", &Foo{1}, "f2", &Bar{2})
	if err != nil {
		t.Errorf("hsetobjects failed: %v", err)
	}
}

func TestHMGetObjects(t *testing.T) {
	client := newRedisClient(t)
	ctx := context.Background()

	err := client.HSetObjects(ctx, "TestHMGetObjects", "f1", &Foo{1}, "f2", &Bar{2})
	if err != nil {
		t.Errorf("hsetobjects failed: %v", err)
	}

	// dst is slice of known type
	var dst1 = make([]*Foo, 2)
	err = client.HMGetObjects(ctx, "TestHMGetObjects", []string{"f1", "f1"}, dst1)
	if err != nil {
		t.Errorf("hmgetobjects failed: %v", err)
	}
	if dst1[0].F != 1 || dst1[1].F != 1 {
		t.Errorf("hmgetobjects failed")
	}

	// dst is slice of interface
	var dst2 = make([]any, 3)
	var fp *Foo
	var bp *Bar
	dst2[0] = fp
	dst2[1] = bp
	dst2[2] = fp
	err = client.HMGetObjects(ctx, "TestHMGetObjects", []string{"f1", "f2", "f3"}, dst2)
	if err != nil {
		t.Errorf("hmgetobjects failed: %v", err)
	}
	f, ok := dst2[0].(*Foo)
	if !ok || f.F != 1 {
		t.Errorf("hmgetobjects failed")
	}

	b, ok := dst2[1].(*Bar)
	if !ok || b.B != 2 {
		t.Errorf("hmgetobjects failed")
	}

	f, ok = dst2[2].(*Foo)
	if !ok || f != nil {
		t.Errorf("hmgetobjects failed")
	}
}

func TestZRangeObjects(t *testing.T) {
	client := newRedisClient(t)
	ctx := context.Background()

	// dst is slice of interface
	var dst = make([]any, 4)
	dst[0] = 1
	dst[1] = Foo{F: 1}
	dst[2] = 2
	dst[3] = Foo{F: 2}
	n, err := client.ZAddObjects(ctx, "TestZRangeObjects", dst...)
	if err != nil {
		t.Errorf("hmgetobjects failed: %v", err)
	}
	t.Log(n, err)

	dst2 := make([]Foo, 0, 0)
	err = client.ZRangeObjects(ctx, "TestZRangeObjects", 0, -1, &dst2)
	if err != nil {
		t.Errorf("hmgetobjects failed: %v", err)
	}
	//t.Log(*dst2[0], *dst2[1])
	t.Log(dst2)

	var dst3 []*Foo
	t.Log(dst3, unsafe.Pointer(&dst3))
	scores, err := client.ZRangeObjectsWithScores(ctx, "TestZRangeObjects", 0, -1, &dst3)
	if err != nil {
		t.Errorf("hmgetobjects failed: %v", err)
	}
	//t.Log(*dst2[0], *dst2[1])
	t.Log(dst3, unsafe.Pointer(&dst3), scores)
}
