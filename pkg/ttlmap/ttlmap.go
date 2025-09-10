package ttlmap

import (
	"sync"
	"time"
)

type item[V any] struct {
	value      V
	expiryTime time.Time
}

// TTLMap stores values with a time-to-live
// Expired items are removed during cleanup or upon access
type TTLMap[K comparable, V any] struct {
	mu     sync.RWMutex
	m      map[K]*item[V]
	maxTTL time.Duration
}

// NewTTLMap creates a new TTL map that stores items for up to maxTTL
// A background goroutine periodically removes expired items every cleanupInterval
func NewTTLMap[K comparable, V any](maxTTL, cleanupInterval time.Duration) *TTLMap[K, V] {
	m := &TTLMap[K, V]{
		m:      make(map[K]*item[V], 1_000),
		maxTTL: maxTTL,
	}
	ticker := time.NewTicker(cleanupInterval)
	go func() {
		defer ticker.Stop()
		for now := range ticker.C {
			m.mu.Lock()
			for k, v := range m.m {
				if now.After(v.expiryTime) {
					delete(m.m, k)
				}
			}
			m.mu.Unlock()
		}
	}()
	return m
}

// Len returns the number of items in the map
func (m *TTLMap[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.m)
}

// Put stores value under specified key and refreshes its expiry time
func (m *TTLMap[K, V]) Put(k K, v V) {
	m.mu.Lock()
	defer m.mu.Unlock()

	it, ok := m.m[k]
	if !ok {
		it = &item[V]{value: v}
		m.m[k] = it
	} else {
		it.value = v
	}
	it.expiryTime = time.Now().Add(m.maxTTL)
}

// Delete removes value associated with the given key
func (m *TTLMap[K, V]) Delete(k K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.m, k)
}

// Get returns value associated with key
// If item has expired it is removed and zero_value and false are returned
func (m *TTLMap[K, V]) Get(k K) (val V, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	it, ok := m.m[k]
	if !ok || time.Now().After(it.expiryTime) {
		if ok {
			delete(m.m, k)
		}
		return val, false
	}
	return it.value, true
}

// GetAndRefresh returns value associated with the given key and refreshes its expiry time
// If key does not exist, zero_value and false are returned
func (m *TTLMap[K, V]) GetAndRefresh(k K) (val V, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	it, ok := m.m[k]
	if !ok {
		return val, false
	}
	it.expiryTime = time.Now().Add(m.maxTTL)
	return it.value, true
}

// Do executes fn with value associated with key under a read lock
// If key is missing or expired, fn is not executed and false is returned
// Expired items are removed before returning
func (m *TTLMap[K, V]) Do(k K, fn func(v V)) bool {
	m.mu.RLock()
	it, ok := m.m[k]
	if !ok || time.Now().After(it.expiryTime) {
		m.mu.RUnlock()
		m.mu.Lock()
		if it2, ok2 := m.m[k]; ok2 && time.Now().After(it2.expiryTime) {
			delete(m.m, k)
		}
		m.mu.Unlock()
		return false
	}
	val := it.value
	m.mu.RUnlock()
	fn(val)
	return true
}
