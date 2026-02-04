# Package: utils

**File:** `sliceutil.go`

## Functions

### ChunkSlice

```go
func ChunkSlice[T any](slice []T, chunkSize int) [][]T
```

ChunkSlice splits a slice into chunks of the specified size.
The last chunk may be smaller than chunkSize if the slice length is not divisible by chunkSize.

---

### ContainsInSlice

```go
func ContainsInSlice[T comparable](slice []T, value T) bool
```

ContainsInSlice checks if value exists in slice.
Returns true if found, false otherwise.

---

### ExcludeFromSlice

```go
func ExcludeFromSlice[T comparable](slice , excludeValues []T) []T
```

ExcludeFromSlice returns a new slice with all values from excludeValues removed from slice
not optimized for performance

---

### MapSlice

```go
func MapSlice[T, U any](src []T, converter (T) U) []U
```

MapSlice applies a converter function to each element of src and returns a new slice.
The result slice has the same length as src.

---

### UniqueSlice

```go
func UniqueSlice[T fmt.Stringer](slice []T) []T
```

UniqueSlice returns a new slice containing only unique elements from the input.
Uniqueness is determined by the String() method of each element.
The order of elements is preserved (first occurrence is kept).
