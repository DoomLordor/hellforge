package clickhouse

import (
	"errors"
)

var (
	ErrNoConnects = errors.New("must provide at least one connection")
	ErrNoRunner   = errors.New("no runner available")
)
