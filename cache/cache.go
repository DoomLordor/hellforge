package cache

import (
	"sync"
	"time"
)

type Cache[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V)
	SetWithTTL(key K, val V, ttl time.Duration)
}

type value[V any] struct {
	deathTime *time.Time
	value     V
}

type cache[K comparable, V any] struct {
	data       map[K]*value[V]
	mu         *sync.RWMutex
	defaultTTL time.Duration
	clearTTL   time.Duration
}

func NewCache[K comparable, V any](options ...Option[K, V]) Cache[K, V] {
	c := &cache[K, V]{
		data:       make(map[K]*value[V], 100),
		mu:         &sync.RWMutex{},
		defaultTTL: 0,
		clearTTL:   5 * time.Minute,
	}

	for _, opt := range options {
		c = opt(c)
	}

	go c.runClearCacheTTL()

	return c
}

func (c *cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var def V
	v, ok := c.data[key]
	if !ok {
		return def, false
	}

	if v.deathTime.Before(time.Now()) {
		return def, false
	}

	return v.value, true
}

func (c *cache[K, V]) Set(key K, val V) {
	c.set(key, val, c.defaultTTL)
}

func (c *cache[K, V]) SetWithTTL(key K, val V, ttl time.Duration) {
	c.set(key, val, ttl)
}

func (c *cache[K, V]) set(key K, val V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v := &value[V]{value: val}
	if ttl != 0 {
		deathTime := time.Now().Add(ttl)
		v.deathTime = &deathTime
	}

	c.data[key] = v
}

func (c *cache[K, V]) clearCacheTTL() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for key, val := range c.data {
		if val.deathTime != nil && val.deathTime.Before(now) {
			delete(c.data, key)
		}
	}
}

func (c *cache[K, V]) runClearCacheTTL() {
	if c.clearTTL == 0 {
		return
	}

	timer := time.NewTicker(c.clearTTL)
	for {
		<-timer.C
		c.clearCacheTTL()
	}
}
