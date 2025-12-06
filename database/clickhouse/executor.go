package clickhouse

import (
	"context"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/doug-martin/goqu/v9"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// SQLConverter query builder to sql with args converter (accept any squirrel builder interface)
type SQLConverter interface {
	ToSQL() (string, []interface{}, error)
}

type Executor struct {
	conn   clickhouse.Conn
	config *config
}

func NewExecutor(conn clickhouse.Conn, options ...Option) *Executor {
	cfg := newConfig()

	for _, option := range options {
		option(cfg)
	}

	return &Executor{
		conn:   conn,
		config: cfg,
	}
}

// QB sets placeholder format for postgres
func (q *Executor) QB(table any) *goqu.SelectDataset {
	return goqu.From(table).Prepared(true)
}

func (q *Executor) Scan(ctx context.Context, sq SQLConverter, resp interface{}, scanFunc ScanFunc) error {
	query, args, err := sq.ToSQL()
	if err != nil {
		return err
	}

	var span trace.Span
	if q.config.tracer != nil {
		ctx, span = q.traceQuery(ctx, query, args...)
		defer span.End()
	}

	err = scanFunc(ctx, q.conn, resp, query, args...)
	if span != nil {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "succeeded")
		}
	}

	return err
}

// Get query for only one row. If no rows are found it returns a pgx.ErrNoRows error.
func (q *Executor) Get(ctx context.Context, sq SQLConverter, resp interface{}) error {
	return q.Scan(ctx, sq, resp, wrapGet)
}

// Select query for many rows. Accept slice as destination resp. If no rows are found - it returns nil error.
func (q *Executor) Select(ctx context.Context, sq SQLConverter, resp interface{}) error {
	return q.Scan(ctx, sq, resp, wrapSelect)
}

// Exec query for no result queries (insert/update/delete without "RETURNING any" suffix)
func (q *Executor) Exec(ctx context.Context, sq SQLConverter) error {
	return q.Scan(ctx, sq, nil, wrapExec)
}

func (q *Executor) BatchStruct(ctx context.Context, query string, items []any) error {
	var span trace.Span
	if q.config.tracer != nil {
		ctx, span = q.traceQuery(ctx, query)
		defer span.End()
	}

	batch, err := q.conn.PrepareBatch(ctx, query)
	if err != nil {
		return err
	}

	for _, arg := range items {
		err = batch.AppendStruct(arg)
		if err != nil {
			return err
		}
	}

	err = batch.Send()
	if span != nil {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "succeeded")
		}
	}

	return err
}

// returns global tracer with default name (if set) otherwise returns noOp trace provider
func (q *Executor) traceQuery(ctx context.Context, query string, args ...interface{}) (context.Context, trace.Span) {
	query = cutString(query, q.config.cutQueryLen)

	ctx, span := q.config.tracer.Start(ctx, query)
	if q.config.withArgs {
		var cutLen uint

		if q.config.cutArgsLen == 0 {
			cutLen = uint(len(args))
		} else {
			cutLen = min(uint(len(args)), q.config.cutArgsLen)
		}

		stringSlice := make([]string, 0, cutLen)
		for _, v := range args[:cutLen] {
			stringSlice = append(stringSlice, defineString(v))
		}

		span.SetAttributes(attribute.String("args", strings.Join(stringSlice, ",")))
	}

	return ctx, span
}
