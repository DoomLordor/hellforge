package clickhouse

import (
	"go.opentelemetry.io/otel/trace"
)

// Option alias for Executor options
type Option func(e *Executor)

// defaultCuttingSize default span name/attributes len cutting size
const defaultCuttingSize = 1000

// WithTracing enable tracing with arguments with default query/args len cutting
func WithTracing(tracer trace.Tracer) Option {
	return func(e *Executor) {
		e.tracer = tracer
	}
}

// WithTracingArgs enable tracing with arguments
func WithTracingArgs() Option {
	return func(q *Executor) {
		q.withArgs = true
	}
}

// WithCutLenTracing enable tracing with cutting query len (0 for disable cutting)
func WithCutLenTracing(queryLen uint) Option {
	return func(q *Executor) {
		q.cutQueryLen = queryLen
	}
}

// WithCutLenArgs enable arguments in tracing with cutting args len (0 for disable cutting)
func WithCutLenArgs(argsLen uint) Option {
	return func(q *Executor) {
		q.cutArgsLen = argsLen
	}
}
