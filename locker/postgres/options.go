package postgres

import "time"

type Option func(c *config)

// WithLeaseDuration sets lease duration
func WithLeaseDuration(d time.Duration) Option {
	return func(c *config) {
		c.leaseDuration = d
	}
}

// WithHeartbeatFrequency set heartbeat frequency
func WithHeartbeatFrequency(d time.Duration) Option {
	return func(c *config) {
		c.heartbeatFrequency = d
	}
}

// WithCustomTable sets lock table name
func WithCustomTable(name string) Option {
	return func(c *config) {
		c.tableName = name
	}
}
