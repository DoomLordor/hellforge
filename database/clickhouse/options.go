package clickhouse

import (
	"github.com/ClickHouse/clickhouse-go/v2"
	"go.opentelemetry.io/otel/trace"
)

// Option alias for Executor options
type Option func(c *config)

// defaultCuttingSize default span name/attributes len cutting size
const defaultCuttingSize = 1000

// WithWriteConnects add write connects
func WithWriteConnects(connects ...clickhouse.Conn) Option {
	return func(c *config) {
		c.writeConnects = append(c.writeConnects, connects...)
	}
}

// WithReadConnects add read connects
func WithReadConnects(connects ...clickhouse.Conn) Option {
	return func(c *config) {
		c.readConnects = append(c.readConnects, connects...)
	}
}

// WithTracing enable tracing with arguments with default query/args len cutting
func WithTracing(provider trace.TracerProvider) Option {
	return func(c *config) {
		c.provider = provider
	}
}

// WithTracingArgs enable tracing with arguments
func WithTracingArgs() Option {
	return func(c *config) {
		c.withArgs = true
	}
}

// WithCutLenTracing enable tracing with cutting query len (0 for disable cutting)
func WithCutLenTracing(queryLen uint) Option {
	return func(c *config) {
		c.cutQueryLen = queryLen
	}
}

// WithCutLenArgs enable arguments in tracing with cutting args len (0 for disable cutting)
func WithCutLenArgs(argsLen uint) Option {
	return func(c *config) {
		c.cutArgsLen = argsLen
	}
}
