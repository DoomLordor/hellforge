package postgres

import (
	"time"
)

type config struct {
	tryCreateTable     bool
	leaseDuration      time.Duration
	heartbeatFrequency time.Duration
	tableName          string
}

func newConfig() *config {
	return &config{
		tryCreateTable:     true,
		leaseDuration:      5 * time.Second,
		heartbeatFrequency: 1 * time.Second,
		tableName:          "lock",
	}
}
