package redis

import (
	"time"
)

type config struct {
	ttl        time.Duration
	refreshTTL time.Duration
	retryTTL   time.Duration
	errorCount int
}

func newConfig() *config {
	return &config{
		ttl:        6 * time.Second,
		refreshTTL: 2 * time.Second,
		retryTTL:   1 * time.Second,
		errorCount: 3,
	}
}
