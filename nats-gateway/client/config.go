package client

import (
	"time"

	"go.opentelemetry.io/otel/trace"
)

type config struct {
	provider       trace.TracerProvider
	timeout        time.Duration
	maxMessageSize int
}

func newConfig() *config {
	return &config{
		timeout:        10 * time.Second,
		maxMessageSize: 1024 * 1024 * 10,
	}
}
