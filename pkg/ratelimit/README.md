# Rate limiting
NOTE: limits are byte based not message-count based

The `ratelimit` package offers simple, testable helpers for enforcing message quotas.
It tracks usage through a pluggable `UsageData` store that guarantees atomic updates
by exposing a `WithUsage` callback. Two stores are provided:

- `MemoryUsage` – in-memory counters for single-process applications.
- `FileUsage` – JSON file-backed counters for persistence across runs.

## Example

```go
store := ratelimit.NewMemoryUsage()
now := time.Now()

if err := ratelimit.CheckMessageSize(int64(len(msg)), 1024); err != nil {
    return err
}
if err := ratelimit.CheckPerSecond(store, 5, now); err != nil {
    return err
}
if err := ratelimit.CheckPerHour(store, 100, now); err != nil {
    return err
}
if err := ratelimit.CheckDaily(store, int64(len(msg)), 1<<20, now); err != nil {
    return err
}
```

`CheckDaily` enforces a rolling 24‑hour window: the counter resets 24 hours after
the first event in the current window rather than at calendar midnight.

Use `ratelimit.IsRateLimitError` to distinguish quota failures from other errors:

```go
if err := ratelimit.CheckPerSecond(store, 1, time.Now()); err != nil {
    if ratelimit.IsRateLimitError(err) {
        // handle quota exceeded
    }
}
```

## Examples

Several runnable programs live under [`examples/ratelimit`](../../examples/ratelimit).
Execute them with `go run`:

```bash
go run ./examples/ratelimit/basic        # basic size, per-second and daily checks
go run ./examples/ratelimit/full         # adds per-hour limits
go run ./examples/ratelimit/walkthrough  # step-by-step demonstration
go run ./examples/ratelimit/concurrent   # concurrent stress test
```

## Currently
| Repo                      | Rate‑limiting logic | Notes                                                                                                                |
| ------------------------- |---------------------|----------------------------------------------------------------------------------------------------------------------|
| `optimum-common`          | ✅                   | Provides a shared `ratelimit` package with file and in-memory backends.                                              |
| `mump2p-cli`              | ✅                   | Client-side limiter (`internal/ratelimit`) checks token claims before publishing.                                    |
| `optimum-proxy`           | ✅                   | Middleware (`pkg/proxy/middleware`) enforces JWT-based per-second, hourly, and quota limits.                         |
| `optimum-p2p`             | ✅                   | API layer (`pkg/routes/base.go` and `token_parser.go`) checks usage against token claims before accepting publishes. |
| `optimum-gateway`         | ❌                   | No rate-limiting logic yet. Acts as a network entry point and should enforce limits.                                 |

