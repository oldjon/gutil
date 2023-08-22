package gdb

import (
	"context"
	"errors"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

type TraceHook struct {
	Tracer     opentracing.Tracer
	Instance   string
	RedisMode  string
	Marshaller string
}

func (h *TraceHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	var span opentracing.Span
	command := cmd.Name()
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		span = h.Tracer.StartSpan(command, opentracing.ChildOf(parentSpan.Context()))
	} else {
		span = h.Tracer.StartSpan(command)
	}

	span.SetTag("span.kind", "client")
	span.SetTag("db.system", "redis")
	span.SetTag("redis.mode", h.RedisMode)
	args := cmd.Args()
	if len(args) >= 2 {
		if key, ok := args[1].(string); ok {
			span.SetTag("redis.keys", key)
		}
	}
	span.SetTag("redis.instance", h.Instance)

	newCtx := opentracing.ContextWithSpan(ctx, span)

	return newCtx, nil
}

func (h *TraceHook) AfterProcess(ctx context.Context, cmd redis.Cmder) (err error) {
	span := opentracing.SpanFromContext(ctx)

	// log the error if we have
	err = cmd.Err()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			//  translate redis.Nil to cache missed
			span.LogFields(log.String("message", "nil"))
		} else {
			span.LogFields(log.String("event", "error"), log.String("message", err.Error()))
		}

	}

	span.Finish()
	return nil
}

func (h *TraceHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	var span opentracing.Span
	command := cmds[0].Name()
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		span = h.Tracer.StartSpan(" Pipelined_"+command, opentracing.ChildOf(parentSpan.Context()))
	} else {
		span = h.Tracer.StartSpan(" Pipelined_" + command)
	}

	span.SetTag("span.kind", "client")
	span.SetTag("db.system", "redis")
	span.SetTag("redis.instance", h.Instance)

	var keys = make([]string, 0, len(cmds))
	for _, cmd := range cmds {
		args := cmd.Args()
		if len(args) > 2 {
			if key, ok := args[1].(string); ok {
				keys = append(keys, key)
			}
		}
	}

	if len(keys) > 0 {
		span.SetTag("redis.keys", strings.Join(keys, " "))
		span.SetTag("redis.keyCount", len(keys))
	}

	newCtx := opentracing.ContextWithSpan(ctx, span)

	return newCtx, nil
}

func (h *TraceHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) (err error) {
	span := opentracing.SpanFromContext(ctx)

	for _, cmd := range cmds {
		if cmd.Err() != nil {
			err = cmd.Err()
			break
		}
	}
	// log the error if we have
	if err != nil {
		if errors.Is(err, redis.Nil) {
			//  translate redis.Nil to cache missed
			span.LogFields(log.String("message", "nil"))
		} else {
			span.LogFields(log.String("event", "error"), log.String("message", err.Error()))
		}

	}

	span.Finish()
	return nil
}
