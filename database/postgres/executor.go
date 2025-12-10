package postgres

import (
	"context"
	"strings"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel/codes"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Runner interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)

	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

// SQLConverter query builder to sql with args converter
type SQLConverter interface {
	ToSQL() (string, []interface{}, error)
}

type Executor struct {
	conn   *pgxpool.Pool
	config *config
}

func NewExecutor(conn *pgxpool.Pool, options ...Option) *Executor {
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
func (e *Executor) QB(table any) *goqu.SelectDataset {
	return goqu.From(table).Prepared(true).WithDialect("postgres")
}

func (e *Executor) runner(ctx context.Context) Runner {
	tx, ok := ctx.Value(txRunnerKey{}).(pgx.Tx)
	if ok {
		return tx
	}

	return e.conn
}

func (e *Executor) Scan(ctx context.Context, sq SQLConverter, resp interface{}, scanFunc ScanFunc) error {
	query, args, err := sq.ToSQL()
	if err != nil {
		return err
	}

	var span trace.Span
	if e.config.tracer != nil {
		ctx, span = e.traceQuery(ctx, query, args...)
		defer span.End()
	}

	err = scanFunc(ctx, e.runner(ctx), resp, query, args...)
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
func (e *Executor) Get(ctx context.Context, sq SQLConverter, resp interface{}) error {
	return e.Scan(ctx, sq, resp, wrapGet)
}

// Select query for many rows. Accept slice as destination resp. If no rows are found - it returns nil error.
func (e *Executor) Select(ctx context.Context, sq SQLConverter, resp interface{}) error {
	return e.Scan(ctx, sq, resp, wrapSelect)
}

// Exec query for no result queries (insert/update/delete without "RETURNING any" suffix)
func (e *Executor) Exec(ctx context.Context, sq SQLConverter) error {
	return e.Scan(ctx, sq, nil, wrapExec)
}

// CopyFrom uses the PostgreSQL copy protocol to perform bulk data insertion. It returns the number of rows copied and
// an error.
func (e *Executor) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	rowsProcessed, err := e.runner(ctx).CopyFrom(ctx, tableName, columnNames, rowSrc)
	if err != nil {
		return 0, err
	}

	return rowsProcessed, nil
}

// RunInTransaction runs function f inside db transaction block using specified executor
func (e *Executor) RunInTransaction(ctx context.Context, f func(ctx context.Context) error) (err error) {
	var tx pgx.Tx
	tx, err = e.conn.Begin(ctx)
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, txRunnerKey{}, tx)

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	err = f(ctx)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// returns global tracer with default name (if set) otherwise returns noOp trace provider
func (e *Executor) traceQuery(ctx context.Context, query string, args ...interface{}) (context.Context, trace.Span) {
	query = cutString(query, e.config.cutQueryLen)

	ctx, span := e.config.tracer.Start(ctx, query)
	if e.config.withArgs {
		var cutLen uint

		if e.config.cutArgsLen == 0 {
			cutLen = uint(len(args))
		} else {
			cutLen = min(uint(len(args)), e.config.cutArgsLen)
		}

		stringSlice := make([]string, 0, cutLen)
		for _, v := range args[:cutLen] {
			stringSlice = append(stringSlice, defineString(v))
		}

		span.SetAttributes(attribute.String("args", strings.Join(stringSlice, ",")))
	}

	return ctx, span
}
