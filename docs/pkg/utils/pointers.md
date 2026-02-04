# Package: utils

**File:** `pointers.go`

## Functions

### FromPointer

```go
func FromPointer[T any](value *T) T
```

FromPointer returns the value pointed to by value, or the zero value of T if value is nil.

---

### ToPointer

```go
func ToPointer[T any](value T) *T
```

ToPointer returns a pointer to the provided value.
