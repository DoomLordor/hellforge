package client

import (
	"time"

	"go.opentelemetry.io/otel/trace"
)

type config struct {
	tracer         trace.Tracer
	timeout        time.Duration
	maxMessageSize int
}

func newConfig() *config {
	return &config{
		timeout:        10 * time.Second,
		maxMessageSize: 1024 * 1024 * 10,
	}
}
