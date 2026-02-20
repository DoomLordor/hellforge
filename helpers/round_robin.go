package helpers

import (
	"sync"
)

type RoundRobin[T any] interface {
	Next() T
}

type roundRobin[T any] struct {
	mu     sync.Mutex
	index  int
	values []T
}

func NewRoundRobin[T any](values []T) RoundRobin[T] {
	return &roundRobin[T]{
		index:  0,
		values: values,
	}
}

func (r *roundRobin[T]) Next() T {
	r.mu.Lock()
	defer r.mu.Unlock()
	value := r.values[r.index]
	r.index++
	if r.index >= len(r.values) {
		r.index = 0
	}
	return value
}
