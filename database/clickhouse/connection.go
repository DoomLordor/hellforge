package clickhouse

import (
	"context"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/doug-martin/goqu/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/DoomLordor/hellforge/helpers"
)

type Connection interface {
	QB(table any) *goqu.SelectDataset
	Write(ctx context.Context) Executor
	Read(ctx context.Context) Executor
}

type connection struct {
	writeConnects helpers.RoundRobin[clickhouse.Conn]
	readConnects  helpers.RoundRobin[clickhouse.Conn]
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
	var readRobin helpers.RoundRobin[clickhouse.Conn]
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
func (c *connection) QB(table any) *goqu.SelectDataset {
	return goqu.From(table).Prepared(true)
}

func (c *connection) getQuerier(ctx context.Context, connects helpers.RoundRobin[clickhouse.Conn]) Executor {
	q := &executor{
		ctx:    ctx,
		runner: connects.Next(),
	}

	if c.tracer != nil {
		q.tracer = c
	}

	return q
}

// Write get Executor for write query
func (c *connection) Write(ctx context.Context) Executor {
	return c.getQuerier(ctx, c.writeConnects)
}

// Read get Executor for read query
func (c *connection) Read(ctx context.Context) Executor {
	return c.getQuerier(ctx, c.readConnects)
}

// traceQuery returns global tracer with default name (if set) otherwise returns noOp trace provider
func (c *connection) traceQuery(ctx context.Context, query string, args ...interface{}) (context.Context, trace.Span) {
	query = helpers.CutString(query, c.cutQueryLen)

	ctx, span := c.tracer.Start(ctx, query)
	if c.withArgs {
		var cutLen uint

		if c.cutArgsLen == 0 {
			cutLen = uint(len(args))
		} else {
			cutLen = min(uint(len(args)), c.cutArgsLen)
		}

		stringSlice := make([]string, 0, cutLen)
		for _, v := range args[:cutLen] {
			stringSlice = append(stringSlice, helpers.DefineString(v))
		}

		span.SetAttributes(attribute.String("args", strings.Join(stringSlice, ",")))
	}

	return ctx, span
}
