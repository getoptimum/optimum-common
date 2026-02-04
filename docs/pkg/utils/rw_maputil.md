# Package: utils

**File:** `rw_maputil.go`

## Functions

### NewRWMap

```go
func NewRWMap[K comparable, V any]() *RWMap[K, V]
```

NewRWMap creates a new empty thread-safe map

---

### NewRWMapFromStdMap

```go
func NewRWMapFromStdMap[K comparable, V any](m map[K]V) *RWMap[K, V]
```

NewRWMapFromStdMap creates a new thread-safe map from an existing standard map

---

## Types

### RWMap

```go
type RWMap[K comparable, V any] struct{...}
```

RWMap provides a thread-safe map implementation using read-write locks

**Methods:**

- `Delete` - Delete removes a key-value pair from the map
- `DeleteAll` - DeleteAll removes all key-value pairs from the map
- `Do` - Do executes a function on a value under read lock. Returns false if the key is not found
- `DoAndApply` - DoAndApply executes a function on a value under write lock and stores the result back
- `Keys` - Keys returns a slice of all keys in the map
- `Len` - Len returns the number of entries in the map
- `Load` - Load retrieves a value from the map by its key
- `LoadAll` - LoadAll returns a copy of the entire map
- `LoadAllAndErase` - LoadAllAndErase returns a copy of the entire map and clears the stored entries
- `Range` - Range calls the provided function for each key/value pair in the map
- `Replace` - Replace swaps the internal map with the provided one and returns the previous map
- `Store` - Store adds or updates a key-value pair in the map
- `Upsert` - Upsert inserts a zero value if the key does not exist, or updates the value using the provided function
