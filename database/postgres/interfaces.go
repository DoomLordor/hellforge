package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel/trace"
)

type Runner interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)

	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

// SQLConverter query builder to sql with args converter
type SQLConverter interface {
	ToSQL() (string, []any, error)
}

type tracer interface {
	traceQuery(ctx context.Context, query string, args ...any) (context.Context, trace.Span)
}
