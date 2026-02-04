# Package: utils

**File:** `maps.go`

## Functions

### MapKeys

```go
func MapKeys[M ..., K comparable, V any](m M) []K
```

MapKeys returns the keys of the map m.
The keys will be in an indeterminate order.

---

### MapValues

```go
func MapValues[M ..., K comparable, V any](m M) []V
```

MapValues returns the values of the map m.
The values will be in an indeterminate order.
