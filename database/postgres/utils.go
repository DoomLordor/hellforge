package postgres

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
)

type txRunnerKey struct{}

type ScanFunc func(ctx context.Context, r Runner, dst any, query string, args ...any) error

func wrapGet(ctx context.Context, r Runner, dst any, query string, args ...any) error {
	return pgxscan.Get(ctx, r, dst, query, args...)
}

func wrapSelect(ctx context.Context, r Runner, dst any, query string, args ...any) error {
	return pgxscan.Select(ctx, r, dst, query, args...)
}

func wrapExec(ctx context.Context, r Runner, _ any, query string, args ...any) error {
	_, err := r.Exec(ctx, query, args...)
	return err
}
