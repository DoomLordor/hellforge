package postgres

import (
	"context"
	"strings"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Runner interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
}

// SQLConverter query builder to sql with args converter (accept any squirrel builder interface)
type SQLConverter interface {
	ToSQL() (string, []interface{}, error)
}

type Executor struct {
	db          *pgxpool.Pool
	tracer      trace.Tracer
	withArgs    bool
	cutQueryLen uint
	cutArgsLen  uint
}

func NewExecutor(db *pgxpool.Pool, opts ...Option) *Executor {
	e := &Executor{
		db:          db,
		tracer:      nil,
		withArgs:    false,
		cutQueryLen: defaultCuttingSize,
		cutArgsLen:  defaultCuttingSize,
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// QB sets placeholder format for postgres
func (q *Executor) QB() goqu.DialectWrapper {
	return goqu.Dialect("postgres")
}

func (q *Executor) runner(ctx context.Context) Runner {
	tx, ok := ctx.Value(txRunnerKey{}).(pgx.Tx)
	if ok {
		return tx
	}

	return q.db
}

func (q *Executor) Scan(ctx context.Context, sq SQLConverter, resp interface{}, scanFunc ScanFunc) error {
	query, args, err := sq.ToSQL()
	if err != nil {
		return err
	}

	var span trace.Span
	if q.tracer != nil {
		ctx, span = q.traceQuery(ctx, query, args...)
		defer span.End()
	}

	err = scanFunc(ctx, q.runner(ctx), resp, query, args...)
	if err != nil && span != nil {
		span.SetStatus(codes.Error, err.Error())
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

// RunInTransaction runs function f inside db transaction block using specified executor
func (q *Executor) RunInTransaction(ctx context.Context, f func(ctx context.Context) error) (err error) {
	var tx pgx.Tx
	tx, err = q.db.Begin(ctx)
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
func (q *Executor) traceQuery(ctx context.Context, query string, args ...interface{}) (context.Context, trace.Span) {
	query = cutString(query, q.cutQueryLen)

	ctx, span := q.tracer.Start(ctx, query)
	if q.withArgs {
		var cutLen uint

		if q.cutArgsLen == 0 {
			cutLen = uint(len(args))
		} else {
			cutLen = min(uint(len(args)), q.cutArgsLen)
		}

		stringSlice := make([]string, 0, cutLen)
		for _, v := range args[:cutLen] {
			stringSlice = append(stringSlice, defineString(v))
		}

		span.SetAttributes(attribute.String("args", strings.Join(stringSlice, ",")))
	}

	return ctx, span
}
