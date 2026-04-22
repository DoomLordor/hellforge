package redis

import (
	"go.opentelemetry.io/otel/trace"
)

type config struct {
	provider trace.TracerProvider
	withArgs bool
}

func newConfig() *config {
	return &config{
		provider: nil,
		withArgs: false,
	}
}
