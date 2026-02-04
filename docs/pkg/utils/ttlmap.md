# Package: utils

**File:** `ttlmap.go`

## Functions

### NewTTLMap

```go
func NewTTLMap[K comparable, V any](maxTTL , cleanupInterval time.Duration) *TTLMap[K, V]
```

NewTTLMap creates a TTL map with a background cleanup goroutine
Call Close() when done to stop the goroutine

---

### NewTTLMapWithContext

```go
func NewTTLMapWithContext[K comparable, V any](ctx context.Context, maxTTL , cleanupInterval time.Duration) *TTLMap[K, V]
```

NewTTLMapWithContext ties the cleanup goroutine to ctx and also supports Close().

---

## Types

### TTLMap

```go
type TTLMap[K comparable, V any] struct{...}
```

TTLMap stores values with a time-to-live
Expired items are removed during cleanup or upon access

**Methods:**

- `Close` - Close stops the background cleanup goroutine and waits for it to exit.
- `Delete` - Delete removes value associated with the given key
- `Do` - Do executes fn with value associated with key under a read lock
- `DoAndApply` - DoAndApply modifies the value associated with the given key using the provided function
- `Get` - 
- `GetAndRefresh` - GetAndRefresh returns value associated with the given key and refreshes its expiry time
- `Len` - Len returns the number of non-expired items currently in the map.
- `Put` - Put stores value under specified key and refreshes its expiry time
- `Upsert` - Upsert inserts a zero value if the key does not exist, or updates the value using the provided function
