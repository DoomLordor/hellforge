package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/trace"
)

type config struct {
	writeConnects []*pgxpool.Pool
	readConnects  []*pgxpool.Pool
	tracer        trace.Tracer
	withArgs      bool
	cutQueryLen   uint
	cutArgsLen    uint
}

func newConfig() *config {
	return &config{
		tracer:      nil,
		withArgs:    false,
		cutQueryLen: defaultCuttingSize,
		cutArgsLen:  defaultCuttingSize,
	}
}
