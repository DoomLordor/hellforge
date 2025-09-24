package cache

import (
	"time"
)

type value[V any] struct {
	deathTime *time.Time
	value     V
}

func (v *value[V]) isDeath(now time.Time) bool {
	return v.deathTime != nil && v.deathTime.Before(now)
}
