package redis

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type tracingHook struct {
	tracer   trace.Tracer
	withArgs bool
}

func (h *tracingHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h *tracingHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		traceArgs := []attribute.KeyValue{
			attribute.String("db.system", "redis"),
			attribute.String("db.operation", cmd.Name()),
		}

		args := cmd.Args()
		if h.withArgs && len(args) > 1 {
			key, ok := args[1].(string)
			if ok {
				traceArgs = append(traceArgs, attribute.String("db.redis.key", key))
			}
		}

		ctx, span := h.tracer.Start(
			ctx,
			"redis."+cmd.Name(),
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(traceArgs...),
		)
		defer span.End()

		err := next(ctx, cmd)
		if err != nil && !errors.Is(err, redis.Nil) {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "success")
		}

		return err
	}
}

func (h *tracingHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}
