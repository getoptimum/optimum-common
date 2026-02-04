# Package: utils

**File:** `rw_slice.go`

## Functions

### NewRWSlice

```go
func NewRWSlice[V any]() *RWSlice[V]
```

NewRWSlice creates a new thread-safe slice

---

## Types

### RWSlice

```go
type RWSlice[V any] struct{...}
```

RWSlice provides a thread-safe slice implementation using read-write locks

**Methods:**

- `Add` - Add appends a single value to the slice
- `AddBulk` - AddBulk appends a list of values to the slice
- `Erase` - Erase removes all elements from the slice
- `Len` - Len returns the number of elements in the slice
- `LoadAll` - LoadAll returns the current slice contents
- `LoadAndErase` - LoadAndErase returns the current slice contents and clears the slice
