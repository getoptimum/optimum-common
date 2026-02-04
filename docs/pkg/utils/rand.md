# Package: utils

**File:** `rand.go`

## Functions

### MsgHash

```go
func MsgHash(topic string, message []byte) string
```

MsgHash returns a deterministic hash of topic + message for message identity (e.g. deduplication).
For time-scoped identity (e.g. same message at different times), use utils.MsgHashWithTimestamp.

---

### RandBetween

```go
func RandBetween(minVal , maxVal int) (int, error)
```

RandBetween returns a uniform random int in [minVal, maxVal)

---

### RandInt

```go
func RandInt() (int, error)
```

RandInt generates a random positive int value

---

### RandInt64

```go
func RandInt64() (int64, error)
```

RandInt64 generates a random non-negative int64 in [0, math.MaxInt64]

---

### Shuffle

```go
func Shuffle[T any](lst []T)
```

Shuffle shuffles a slice of any type
