package client

import (
	"time"

	"go.opentelemetry.io/otel/trace"
)

type Option func(c *config)

func WithTracing(provider trace.TracerProvider) Option {
	return func(c *config) {
		c.provider = provider
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
