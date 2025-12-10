package client

import (
	"time"

	"go.opentelemetry.io/otel/trace"
)

type Option func(c *config)

func WithTracer(tracer trace.Tracer) Option {
	return func(c *config) {
		c.tracer = tracer
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *config) {
		c.timeout = timeout
	}
}

func WithMaxMessageSize(maxMessageSize int) Option {
	return func(c *config) {
		c.maxMessageSize = maxMessageSize
	}
}
