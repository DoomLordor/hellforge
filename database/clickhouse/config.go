package clickhouse

import (
	"github.com/ClickHouse/clickhouse-go/v2"
	"go.opentelemetry.io/otel/trace"
)

type config struct {
	writeConnects []clickhouse.Conn
	readConnects  []clickhouse.Conn
	provider      trace.TracerProvider
	withArgs      bool
	cutQueryLen   uint
	cutArgsLen    uint
}

func newConfig() *config {
	return &config{
		provider:    nil,
		withArgs:    false,
		cutQueryLen: defaultCuttingSize,
		cutArgsLen:  defaultCuttingSize,
	}
}
