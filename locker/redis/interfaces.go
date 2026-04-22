package redis

import (
	"context"
	"time"
)

// redisClient redis client interface
type redisClient interface {
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Del(ctx context.Context, key string) error
}
