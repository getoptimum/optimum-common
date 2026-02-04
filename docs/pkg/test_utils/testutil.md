# Package: test_utils

**File:** `testutil.go`

## Functions

### GetFreePortT

```go
func GetFreePortT(t *testing.T) int
```

GetFreePortT is a wrapper around GetFreePort

---

### NewTestLogWriter

```go
func NewTestLogWriter(t testing.TB) *TestLogWriter
```

NewTestLogWriter creates a writer suitable for wiring into structured loggers in tests
Mirrors all output to the provided std writer (stdout by default) during local runs
When the CI_RUN environment variable is set, the output is buffered in-memory and printed only if the test fails

---

### NewTestLogWriterWithOutput

```go
func NewTestLogWriterWithOutput(t testing.TB, std io.Writer) *TestLogWriter
```

NewTestLogWriterWithOutput behaves like NewTestLogWriter but allows overriding the underlying writer used for local runs

---

### NewTestLogger

```go
func NewTestLogger[T any](t testing.TB, factory ([]io.Writer) T) T
```

NewTestLogger wires the TestLogWriter into a repository specific logger factory

Example:

	logger := testutils.NewTestLogger(t, func(w []io.Writer) logger.AppLogger {
	    return logger.InitLogger(w, logger.Debug)
	})

---

### RunConcurrently

```go
func RunConcurrently(nGoroutines , nOps int, fn ())
```

RunConcurrently runs fn from nGoroutines goroutines, nOps times per goroutine, and waits for completion.
Useful for stress-testing concurrent access (e.g. RWMap, TTLMap).

---

## Types

### TestLogWriter

```go
type TestLogWriter struct{...}
```

TestLogWriter captures slog output during CI runs while mirroring logs to stdout locally

**Methods:**

- `Write` - 
