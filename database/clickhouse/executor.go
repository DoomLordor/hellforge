package clickhouse

import (
	"context"
	"errors"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/doug-martin/goqu/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/DoomLordor/hellforge/helpers"
)

type Executor interface {
	QB(table any) *goqu.SelectDataset
	Write(ctx context.Context) Querier
	Read(ctx context.Context) Querier
}

type executor struct {
	writeConnects helpers.RoundRobin[clickhouse.Conn]
	readConnects  helpers.RoundRobin[clickhouse.Conn]
	tracer        trace.Tracer
	withArgs      bool
	cutQueryLen   uint
	cutArgsLen    uint
}

func NewExecutor(options ...Option) (Executor, error) {
	cfg := newConfig()

	for _, option := range options {
		option(cfg)
	}

	if len(cfg.writeConnects) == 0 {
		return nil, errors.New("must provide at least one connection")
	}

	writeRobin := helpers.NewRoundRobin(cfg.writeConnects)
	var readRobin helpers.RoundRobin[clickhouse.Conn]
	if len(cfg.readConnects) == 0 {
		readRobin = writeRobin
	}

	return &executor{
		writeConnects: writeRobin,
		readConnects:  readRobin,
		tracer:        cfg.tracer,
		withArgs:      cfg.withArgs,
		cutQueryLen:   cfg.cutQueryLen,
		cutArgsLen:    cfg.cutArgsLen,
	}, nil
}

// QB sets placeholder format for postgres
func (e *executor) QB(table any) *goqu.SelectDataset {
	return goqu.From(table).Prepared(true)
}

func (e *executor) getQuerier(ctx context.Context, connects helpers.RoundRobin[clickhouse.Conn]) Querier {
	q := &querier{
		ctx:    ctx,
		runner: connects.Next(),
	}

	if e.tracer != nil {
		q.tracer = e
	}

	return q
}

// Write get Querier for write query
func (e *executor) Write(ctx context.Context) Querier {
	return e.getQuerier(ctx, e.writeConnects)
}

// Read get Querier for read query
func (e *executor) Read(ctx context.Context) Querier {
	return e.getQuerier(ctx, e.readConnects)
}

// traceQuery returns global tracer with default name (if set) otherwise returns noOp trace provider
func (e *executor) traceQuery(ctx context.Context, query string, args ...interface{}) (context.Context, trace.Span) {
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
