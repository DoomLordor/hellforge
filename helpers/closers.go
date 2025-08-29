package helpers

import (
	"context"
)

func WrapCloseFuncWithoutContext(f func() error) func(context.Context) error {
	return func(_ context.Context) error {
		return f()
	}
}

func WrapCloseFuncWithoutError(f func(context.Context)) func(context.Context) error {
	return func(ctx context.Context) error {
		f(ctx)
		return nil
	}
}

func WrapCloseFuncWithoutContextAndError(f func()) func(context.Context) error {
	return func(_ context.Context) error {
		f()
		return nil
	}
}
