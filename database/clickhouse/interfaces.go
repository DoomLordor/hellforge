package clickhouse

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// SQLConverter query builder to sql with args converter
type SQLConverter interface {
	ToSQL() (string, []any, error)
}

type tracer interface {
	traceQuery(ctx context.Context, query string, args ...any) (context.Context, trace.Span)
}
