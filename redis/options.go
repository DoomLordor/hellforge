package redis

import (
	"go.opentelemetry.io/otel/trace"
)

// Option alias for redis cluster client options
type Option func(*config)

// WithTracing enable tracing
func WithTracing(provider trace.TracerProvider) Option {
	return func(c *config) {
		c.provider = provider
	}
}

// WithArgs enable arguments in tracing
func WithArgs() Option {
	return func(c *config) {
		c.withArgs = true
	}
}
