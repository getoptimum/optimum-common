# Package: utils

**File:** `balancer.go`

## Functions

### NewRoundRobinBalancer

```go
func NewRoundRobinBalancer[T any](values []T) *RoundRobinBalancer[T]
```

NewRoundRobinBalancer creates a new round-robin balancer with the given values.
The balancer cycles through values in order, returning to the start after the last value.

---

## Types

### RoundRobinBalancer

```go
type RoundRobinBalancer[T any] struct{...}
```

RoundRobinBalancer provides thread-safe round-robin selection from a slice of values.

**Methods:**

- `Next` - Next returns the next value in round-robin order.
- `Values` - Values returns a copy of all values managed by this balancer.
