package clickhouse

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ScanFunc func(ctx context.Context, r driver.Conn, dst interface{}, query string, args ...interface{}) error

func wrapGet(ctx context.Context, r driver.Conn, dst interface{}, query string, args ...interface{}) error {
	return r.QueryRow(ctx, query, args...).Scan(dst)
}

func wrapSelect(ctx context.Context, r driver.Conn, dst interface{}, query string, args ...interface{}) error {
	return r.Select(ctx, dst, query, args...)
}

func wrapExec(ctx context.Context, r driver.Conn, _ interface{}, query string, args ...interface{}) error {
	return r.Exec(ctx, query, args...)
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
