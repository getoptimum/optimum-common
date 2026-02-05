package syncx

import "sync"

// RWSlice provides a thread-safe slice implementation using read-write locks
type RWSlice[V any] struct {
	mu       sync.RWMutex
	internal []V
}

// NewRWSlice creates a new thread-safe slice
func NewRWSlice[V any]() *RWSlice[V] {
	return &RWSlice[V]{
		internal: make([]V, 0),
	}
}

// Add appends a single value to the slice
func (rw *RWSlice[V]) Add(value V) {
	rw.mu.Lock()
	rw.internal = append(rw.internal, value)
	rw.mu.Unlock()
}

// AddBulk appends a list of values to the slice
func (rw *RWSlice[V]) AddBulk(values []V) {
	rw.mu.Lock()
	rw.internal = append(rw.internal, values...)
	rw.mu.Unlock()
}

// LoadAll returns the current slice contents
func (rw *RWSlice[V]) LoadAll() []V {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	out := make([]V, len(rw.internal))
	copy(out, rw.internal)
	return out
}

// LoadAndErase returns the current slice contents and clears the slice
func (rw *RWSlice[V]) LoadAndErase() []V {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	out := make([]V, len(rw.internal))
	copy(out, rw.internal)
	rw.internal = rw.internal[:0]
	return out
}

// Len returns the number of elements in the slice
func (rw *RWSlice[V]) Len() int {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	return len(rw.internal)
}

// Erase removes all elements from the slice
func (rw *RWSlice[V]) Erase() {
	rw.mu.Lock()
	rw.internal = rw.internal[:0]
	rw.mu.Unlock()
}
