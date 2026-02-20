package clickhouse

import (
	"github.com/ClickHouse/clickhouse-go/v2"
	"go.opentelemetry.io/otel/trace"
)

type config struct {
	writeConnects []clickhouse.Conn
	readConnects  []clickhouse.Conn
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
