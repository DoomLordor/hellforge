package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Querier interface {
	RunRaw(query string, args []any, resp any, scanFunc ScanFunc) error
	Run(sq SQLConverter, resp any, scanFunc ScanFunc) error
	Get(sq SQLConverter, resp any) error
	Select(sq SQLConverter, resp any) error
	Exec(sq SQLConverter) error
}

type querier struct {
	ctx    context.Context
	tracer tracer
	runner Runner
}

func (q *querier) RunRaw(query string, args []any, resp any, scanFunc ScanFunc) error {
	var span trace.Span
	if q.tracer != nil {
		q.ctx, span = q.tracer.traceQuery(q.ctx, query, args...)
		defer span.End()
	}

	err := scanFunc(q.ctx, q.runner, resp, query, args...)
	if span != nil {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "succeeded")
		}
	}

	return err
}

func (q *querier) Run(sq SQLConverter, resp any, scanFunc ScanFunc) error {
	query, args, err := sq.ToSQL()
	if err != nil {
		return err
	}

	return q.RunRaw(query, args, resp, scanFunc)
}

// Get query for only one row. If no rows are found it returns a pgx.ErrNoRows error.
func (q *querier) Get(sq SQLConverter, resp any) error {
	return q.Run(sq, resp, wrapGet)
}

// Select query for many rows. Accept slice as destination resp. If no rows are found - it returns nil error.
func (q *querier) Select(sq SQLConverter, resp any) error {
	return q.Run(sq, resp, wrapSelect)
}

// Exec query for no result queries (insert/update/delete without "RETURNING any" suffix)
func (q *querier) Exec(sq SQLConverter) error {
	return q.Run(sq, nil, wrapExec)
}

// ExecRaw raw query for no result queries (insert/update/delete without "RETURNING any" suffix)
func (q *querier) ExecRaw(query string, args ...any) error {
	return q.RunRaw(query, args, nil, wrapExec)
}

// CopyFrom uses the PostgreSQL copy protocol to perform bulk data insertion. It returns the number of rows copied and an error. VIP
func (q *querier) CopyFrom(tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	rowsProcessed, err := q.runner.CopyFrom(q.ctx, tableName, columnNames, rowSrc)
	if err != nil {
		return 0, err
	}

	return rowsProcessed, nil
}
