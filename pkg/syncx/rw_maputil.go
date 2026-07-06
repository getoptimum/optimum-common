package syncx

import "maps"

import "sync"

// RWMap provides a thread-safe map implementation using read-write locks
type RWMap[K comparable, V any] struct {
	mu       sync.RWMutex
	internal map[K]V
}

// NewRWMap creates a new empty thread-safe map
func NewRWMap[K comparable, V any]() *RWMap[K, V] {
	return &RWMap[K, V]{
		internal: make(map[K]V),
	}
}

// NewRWMapFromStdMap creates a new thread-safe map from an existing standard map
func NewRWMapFromStdMap[K comparable, V any](m map[K]V) *RWMap[K, V] {
	return &RWMap[K, V]{
		internal: m,
	}
}

// Replace swaps the internal map with the provided one and returns the previous map
func (rm *RWMap[K, V]) Replace(m map[K]V) map[K]V {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	current := rm.internal
	rm.internal = m
	return current
}

// Load retrieves a value from the map by its key
func (rm *RWMap[K, V]) Load(key K) (value V, ok bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	value, ok = rm.internal[key]
	return
}

// LoadAll returns a copy of the entire map
func (rm *RWMap[K, V]) LoadAll() map[K]V {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	out := make(map[K]V, len(rm.internal))
	maps.Copy(out, rm.internal)
	return out
}

// LoadAllAndErase returns a copy of the entire map and clears the stored entries
func (rm *RWMap[K, V]) LoadAllAndErase() map[K]V {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	out := make(map[K]V, len(rm.internal))
	maps.Copy(out, rm.internal)
	rm.internal = make(map[K]V)
	return out
}

// Store adds or updates a key-value pair in the map
func (rm *RWMap[K, V]) Store(key K, value V) {
	rm.mu.Lock()
	rm.internal[key] = value
	rm.mu.Unlock()
}

// Delete removes a key-value pair from the map
func (rm *RWMap[K, V]) Delete(key K) {
	rm.mu.Lock()
	delete(rm.internal, key)
	rm.mu.Unlock()
}

// DeleteAll removes all key-value pairs from the map
func (rm *RWMap[K, V]) DeleteAll() {
	rm.mu.Lock()
	rm.internal = make(map[K]V)
	rm.mu.Unlock()
}

// Keys returns a slice of all keys in the map
func (rm *RWMap[K, V]) Keys() []K {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	keys := make([]K, 0, len(rm.internal))
	for k := range rm.internal {
		keys = append(keys, k)
	}
	return keys
}

// Range calls the provided function for each key/value pair in the map
// If the function returns false, iteration stops
func (rm *RWMap[K, V]) Range(f func(key K, value V) bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	for k, v := range rm.internal {
		if !f(k, v) {
			return
		}
	}
}

// Do executes a function on a value under read lock. Returns false if the key is not found
func (rm *RWMap[K, V]) Do(key K, fn func(value V)) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	value, ok := rm.internal[key]
	if !ok {
		return false
	}
	fn(value)
	return true
}

// Upsert inserts a zero value if the key does not exist, or updates the value using the provided function
func (rm *RWMap[K, V]) Upsert(key K, fn func(value V) V, zeroValue V) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	value, ok := rm.internal[key]
	if !ok {
		rm.internal[key] = zeroValue
		return
	}
	rm.internal[key] = fn(value)
}

// DoAndApply executes a function on a value under write lock and stores the result back
func (rm *RWMap[K, V]) DoAndApply(key K, fn func(value V) V) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	value, ok := rm.internal[key]
	if !ok {
		return
	}
	rm.internal[key] = fn(value)
}

// Len returns the number of entries in the map
func (rm *RWMap[K, V]) Len() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return len(rm.internal)
}
