package postgres

import (
	"context"
	"strings"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/DoomLordor/hellforge/helpers"
)

type Connection interface {
	QB(table any) *goqu.SelectDataset
	Write(ctx context.Context) Executor
	Read(ctx context.Context) Executor
	RunInTransaction(ctx context.Context, f func(ctx context.Context) error) (err error)
}

type connection struct {
	writeConnects helpers.RoundRobin[*pgxpool.Pool]
	readConnects  helpers.RoundRobin[*pgxpool.Pool]
	tracer        trace.Tracer
	withArgs      bool
	cutQueryLen   uint
	cutArgsLen    uint
}

func NewConnection(options ...Option) (Connection, error) {
	cfg := newConfig()

	for _, option := range options {
		option(cfg)
	}

	if len(cfg.writeConnects) == 0 && len(cfg.readConnects) == 0 {
		return nil, ErrNoConnects
	}

	writeRobin := helpers.NewRoundRobin(cfg.writeConnects)
	var readRobin helpers.RoundRobin[*pgxpool.Pool]
	if len(cfg.readConnects) == 0 {
		readRobin = writeRobin
	} else {
		readRobin = helpers.NewRoundRobin(cfg.readConnects)
	}

	return &connection{
		writeConnects: writeRobin,
		readConnects:  readRobin,
		tracer:        cfg.tracer,
		withArgs:      cfg.withArgs,
		cutQueryLen:   cfg.cutQueryLen,
		cutArgsLen:    cfg.cutArgsLen,
	}, nil
}

// QB sets placeholder format for postgres
func (e *connection) QB(table any) *goqu.SelectDataset {
	return goqu.From(table).Prepared(true).WithDialect("postgres")
}

func (e *connection) getQuerier(ctx context.Context, connects helpers.RoundRobin[*pgxpool.Pool]) Executor {
	q := &executor{
		ctx: ctx,
	}

	if e.tracer != nil {
		q.tracer = e
	}

	tx, ok := ctx.Value(txRunnerKey{}).(pgx.Tx)
	if ok {
		q.runner = tx
	} else {
		q.runner = connects.Next()
	}

	return q
}

// Write get Executor for write query
func (e *connection) Write(ctx context.Context) Executor {
	return e.getQuerier(ctx, e.writeConnects)
}

// Read get Executor for read query
func (e *connection) Read(ctx context.Context) Executor {
	return e.getQuerier(ctx, e.readConnects)
}

// RunInTransaction runs function f inside db transaction block using specified connection
func (e *connection) RunInTransaction(ctx context.Context, f func(ctx context.Context) error) (err error) {
	conn := e.writeConnects.Next()
	if conn == nil {
		return ErrNoConnects
	}

	var tx pgx.Tx
	tx, err = conn.Begin(ctx)
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

// traceQuery returns global tracer with default name (if set) otherwise returns noOp trace provider
func (e *connection) traceQuery(ctx context.Context, query string, args ...any) (context.Context, trace.Span) {
	query = helpers.CutString(query, e.cutQueryLen)

	ctx, span := e.tracer.Start(ctx, query)
	if e.withArgs {
		var cutLen uint

		if e.cutArgsLen == 0 {
			cutLen = uint(len(args))
		} else {
			cutLen = min(uint(len(args)), e.cutArgsLen)
		}

		stringSlice := make([]string, 0, cutLen)
		for _, v := range args[:cutLen] {
			stringSlice = append(stringSlice, helpers.DefineString(v))
		}

		span.SetAttributes(attribute.String("args", strings.Join(stringSlice, ",")))
	}

	return ctx, span
}
