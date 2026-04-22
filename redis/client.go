package redis

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// ClusterClient redis client interface
type ClusterClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) (string, error)
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error)
	LRange(ctx context.Context, key string, start int64, stop int64) ([]string, error)
	LPush(ctx context.Context, key string, values ...any) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Incr(ctx context.Context, key string) (int64, error)
	Del(ctx context.Context, key string) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	PTTL(ctx context.Context, key string) (time.Duration, error)
	Close() error
	// GetClient for raw client accessibility (used for non wrapped methods access)
	GetClient() *redis.ClusterClient
}

// clusterClient base wrapped cluster client
type clusterClient struct {
	redis *redis.ClusterClient
}

var (
	Nil = redis.Nil
	// ErrNotSet wrap redis not set err
	ErrNotSet = errors.New("value not set")
	// ErrTTLNotSet represent redis ttl key not set err
	ErrTTLNotSet = errors.New("ttl not set")
)

// NewClusterClient redis cluster client constructor
func NewClusterClient(ctx context.Context, clusters, password string, options ...Option) (ClusterClient, error) {
	client := redis.NewClusterClient(
		&redis.ClusterOptions{
			Addrs:    strings.Split(clusters, ","),
			Password: password,
		},
	)

	cfg := newConfig()
	for _, opt := range options {
		opt(cfg)
	}

	if cfg.provider != nil {
		client.AddHook(&tracingHook{
			tracer:   cfg.provider.Tracer("redis-client"),
			withArgs: cfg.withArgs,
		})
	}

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	wrappedClient := &clusterClient{
		redis: client,
	}

	return wrappedClient, nil
}

// Get obtain data
func (c *clusterClient) Get(ctx context.Context, key string) (string, error) {
	return c.redis.Get(ctx, key).Result()
}

// LPush set list values
func (c *clusterClient) LPush(ctx context.Context, key string, values ...any) (int64, error) {
	return c.redis.LPush(ctx, key, values...).Result()
}

// LRange get list values
func (c *clusterClient) LRange(ctx context.Context, key string, start int64, stop int64) (res []string, err error) {
	return c.redis.LRange(ctx, key, start, stop).Result()
}

// Expire set expiration for key
func (c *clusterClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	_, err := c.redis.Expire(ctx, key, expiration).Result()
	return err
}

// Set one value
func (c *clusterClient) Set(ctx context.Context, key string, value any, expiration time.Duration) (string, error) {
	return c.redis.Set(ctx, key, value, expiration).Result()
}

// SetNX with expiration if key already not exists
func (c *clusterClient) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	args := redis.SetArgs{Mode: string(redis.NX)}
	switch {
	case expiration > 0:
		args.TTL = expiration
	case expiration == redis.KeepTTL:
		args.KeepTTL = true
	}

	res, err := c.redis.SetArgs(ctx, key, value, args).Result()
	if err != nil {
		return false, err
	}

	return res == "OK", nil
}

// Incr value
func (c *clusterClient) Incr(ctx context.Context, key string) (int64, error) {
	return c.redis.Incr(ctx, key).Result()
}

// Del key
func (c *clusterClient) Del(ctx context.Context, key string) error {
	_, err := c.redis.Del(ctx, key).Result()
	return err
}

// TTL return key expiration time in seconds
func (c *clusterClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.redis.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if ttl == -1 {
		return 0, ErrTTLNotSet
	}

	if ttl == -2 {
		return 0, ErrNotSet
	}

	return ttl, nil
}

// PTTL return key expiration time in millisecond
func (c *clusterClient) PTTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.redis.PTTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if ttl == -1 {
		return 0, ErrTTLNotSet
	}

	if ttl == -2 {
		return 0, ErrNotSet
	}

	return ttl, nil
}

// Close connection
func (c *clusterClient) Close() error {
	return c.redis.Close()
}

// GetClient return client
func (c *clusterClient) GetClient() *redis.ClusterClient {
	return c.redis
}
