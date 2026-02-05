package syncx

import (
	"sync/atomic"
)

// RoundRobinBalancer provides thread-safe round-robin selection from a slice of values.
type RoundRobinBalancer[T any] struct {
	values []T
	index  *atomic.Int64
}

// NewRoundRobinBalancer creates a new round-robin balancer with the given values.
// The balancer cycles through values in order, returning to the start after the last value.
func NewRoundRobinBalancer[T any](values []T) *RoundRobinBalancer[T] {
	return &RoundRobinBalancer[T]{
		values: values,
		index:  &atomic.Int64{},
	}
}

// Next returns the next value in round-robin order.
// Thread-safe: uses atomic operations for concurrent access.
func (wb *RoundRobinBalancer[T]) Next() T {
	return wb.values[(wb.index.Add(1)-1)%int64(len(wb.values))]
}

// Values returns a copy of all values managed by this balancer.
func (wb *RoundRobinBalancer[T]) Values() []T {
	return wb.values
}
