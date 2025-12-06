package postgres

import (
	"go.opentelemetry.io/otel/trace"
)

type config struct {
	tracer      trace.Tracer
	withArgs    bool
	cutQueryLen uint
	cutArgsLen  uint
}

func newConfig() *config {
	return &config{
		tracer:      nil,
		withArgs:    false,
		cutQueryLen: defaultCuttingSize,
		cutArgsLen:  defaultCuttingSize,
	}
}
