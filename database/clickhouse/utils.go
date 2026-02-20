package clickhouse

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ScanFunc func(ctx context.Context, r driver.Conn, dst any, query string, args ...any) error

func wrapGet(ctx context.Context, r driver.Conn, dst any, query string, args ...any) error {
	return r.QueryRow(ctx, query, args...).Scan(dst)
}

func wrapSelect(ctx context.Context, r driver.Conn, dst any, query string, args ...any) error {
	return r.Select(ctx, dst, query, args...)
}

func wrapExec(ctx context.Context, r driver.Conn, _ any, query string, args ...any) error {
	return r.Exec(ctx, query, args...)
}

func wrapBatch(ctx context.Context, r driver.Conn, _ any, query string, args ...any) error {
	batch, err := r.PrepareBatch(ctx, query)
	if err != nil {
		return err
	}

	for _, arg := range args {
		err = batch.AppendStruct(arg)
		if err != nil {
			return err
		}
	}

	return batch.Send()
}
