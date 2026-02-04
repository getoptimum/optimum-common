# Package: utils

**File:** `math.go`

## Functions

### MustSafeIntToUint32

```go
func MustSafeIntToUint32(i int) uint32
```

MustSafeIntToUint32 invokes SafeIntToUint32 and panics if an error is returned.

---

### MustSafeUint64ToInt64

```go
func MustSafeUint64ToInt64(u uint64) int64
```

MustSafeUint64ToInt64 invokes SafeUint64ToInt64 and panics if an error is returned.

---

### SafeAddUint64Ptr

```go
func SafeAddUint64Ptr(counter *uint64, values ...int) error
```

SafeAddUint64Ptr adds the values to the counter pointer, and returns an error
if the counter would overflow.

---

### SafeIntToUint32

```go
func SafeIntToUint32(i int) (uint32, error)
```

SafeIntToUint32 converts an int to a uint32, returning an error if the value
is negative or exceeds the maximum uint32 value.

---

### SafeUint64ToInt64

```go
func SafeUint64ToInt64(u uint64) (int64, error)
```

SafeUint64ToInt64 converts a uint64 to an int64, returning an error if the
value is too large to fit in an int64.
