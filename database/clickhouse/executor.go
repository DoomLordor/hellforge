package clickhouse

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Executor interface {
	RunRaw(query string, args []any, resp any, scanFunc ScanFunc) error
	Run(sq SQLConverter, resp any, scanFunc ScanFunc) error
	Get(sq SQLConverter, resp any) error
	Select(sq SQLConverter, resp any) error
	Exec(sq SQLConverter) error
	BatchStruct(query string, items []any) error
}

type executor struct {
	ctx    context.Context
	tracer tracer
	runner clickhouse.Conn
}

func (e *executor) RunRaw(query string, args []any, resp any, scanFunc ScanFunc) error {
	var span trace.Span
	if e.tracer != nil {
		e.ctx, span = e.tracer.traceQuery(e.ctx, query, args...)
		defer span.End()
	}

	err := scanFunc(e.ctx, e.runner, resp, query, args...)
	if span != nil {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "succeeded")
		}
	}

	return err
}

func (e *executor) Run(sq SQLConverter, resp any, scanFunc ScanFunc) error {
	query, args, err := sq.ToSQL()
	if err != nil {
		return err
	}

	return e.RunRaw(query, args, resp, scanFunc)
}

// Get query for only one row. If no rows are found it returns a pgx.ErrNoRows error.
func (e *executor) Get(sq SQLConverter, resp any) error {
	return e.Run(sq, resp, wrapGet)
}

// Select query for many rows. Accept slice as destination resp. If no rows are found - it returns nil error.
func (e *executor) Select(sq SQLConverter, resp any) error {
	return e.Run(sq, resp, wrapSelect)
}

// Exec query for no result queries (insert/update/delete without "RETURNING any" suffix)
func (e *executor) Exec(sq SQLConverter) error {
	return e.Run(sq, nil, wrapExec)
}

// ExecRaw raw query for no result queries (insert/update/delete without "RETURNING any" suffix)
func (e *executor) ExecRaw(query string, args ...any) error {
	return e.RunRaw(query, args, nil, wrapExec)
}

func (e *executor) BatchStruct(query string, items []any) error {
	return e.RunRaw(query, items, nil, wrapBatch)
}
