package cache

import "time"

type Option[K comparable, V any] func(c *cache[K, V]) *cache[K, V]

func WithTTL[K comparable, V any](ttl time.Duration) Option[K, V] {
	return func(c *cache[K, V]) *cache[K, V] {
		c.defaultTTL = ttl
		return c
	}
}

func WithClearTTL[K comparable, V any](ttl time.Duration) Option[K, V] {
	return func(c *cache[K, V]) *cache[K, V] {
		c.clearTTL = ttl
		return c
	}
}
