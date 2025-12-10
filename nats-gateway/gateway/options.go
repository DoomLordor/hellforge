package gateway

import (
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type Option func(c *config)

func WithTracer(tracer trace.Tracer) Option {
	return func(c *config) {
		c.tracer = tracer
	}
}

func WithAdapters(adapters ...Adapter) Option {
	return func(c *config) {
		c.adapters = append(c.adapters, adapters...)
	}
}

func WithDialOptions(options ...grpc.DialOption) Option {
	return func(c *config) {
		c.dialOpts = append(c.dialOpts, options...)
	}
}

func WithLimit(limit int) Option {
	return func(c *config) {
		c.limit = limit
	}
}

func WithMaxMessageSize(maxMessageSize int) Option {
	return func(c *config) {
		c.maxMessageSize = maxMessageSize
	}
}
