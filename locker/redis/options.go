package redis

import (
	"time"
)

// Option for distributed lock settings
type Option func(c *config)

// WithTTL set key expiration time
func WithTTL(t time.Duration) Option {
	return func(c *config) {
		c.ttl = t
	}
}

// WithRefreshTTL set key refresh time
func WithRefreshTTL(ttl time.Duration) Option {
	return func(c *config) {
		c.refreshTTL = ttl
	}
}

// WithRetryTTL set retry key lock time
func WithRetryTTL(ttl time.Duration) Option {
	return func(c *config) {
		c.retryTTL = ttl
	}
}

// WithErrorCount set max error count for holding lock
func WithErrorCount(count int) Option {
	return func(c *config) {
		c.errorCount = count
	}
}
