package clickhouse

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ScanFunc func(ctx context.Context, c driver.Conn, dst any, query string, args ...any) error

func wrapGet(ctx context.Context, c driver.Conn, dst any, query string, args ...any) error {
	return c.QueryRow(ctx, query, args...).Scan(dst)
}

func wrapSelect(ctx context.Context, c driver.Conn, dst any, query string, args ...any) error {
	return c.Select(ctx, dst, query, args...)
}

func wrapExec(ctx context.Context, c driver.Conn, _ any, query string, args ...any) error {
	return c.Exec(ctx, query, args...)
}

func wrapBatch(ctx context.Context, c driver.Conn, _ any, query string, args ...any) error {
	batch, err := c.PrepareBatch(ctx, query)
	if err != nil {
		return err
	}

	defer func() {
		_ = batch.Close()
	}()

	for _, arg := range args {
		err = batch.AppendStruct(arg)
		if err != nil {
			return err
		}
	}

	return batch.Send()
}
