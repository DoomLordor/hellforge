package postgres

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
)

type txRunnerKey struct{}

type ScanFunc func(ctx context.Context, r Runner, dst interface{}, query string, args ...interface{}) error

func wrapGet(ctx context.Context, r Runner, dst interface{}, query string, args ...interface{}) error {
	return pgxscan.Get(ctx, r, dst, query, args...)
}

func wrapSelect(ctx context.Context, r Runner, dst interface{}, query string, args ...interface{}) error {
	return pgxscan.Select(ctx, r, dst, query, args...)
}

func wrapExec(ctx context.Context, r Runner, _ interface{}, query string, args ...interface{}) error {
	_, err := r.Exec(ctx, query, args...)
	return err
}

func cutString(s string, minLen uint) string {
	if minLen == 0 {
		return s
	}

	r := []rune(s)
	return string(r[:min(uint(len(r)), minLen)])
}

func defineString(arg interface{}) string {
	switch v := arg.(type) {
	case string:
		return v
	case *string:
		if v != nil {
			return *v
		}
		return ""
	case *int:
		if v != nil {
			return fmt.Sprintf("%v", *v)
		}
		return ""
	case *int32:
		if v != nil {
			return fmt.Sprintf("%v", *v)
		}
		return ""
	case *int64:
		if v != nil {
			return fmt.Sprintf("%v", *v)
		}
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}
