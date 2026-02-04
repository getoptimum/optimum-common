# Package: config

**File:** `rotator.go`

## Functions

### NewConfigRotator

```go
func NewConfigRotator(ctx context.Context, log logger.AppLogger, baseOptCfg *entities.OptimumConfig, chainID , clusterID string, updater (config *entities.DynamicConfig), opts ...RotatorOption) *Rotator
```

NewConfigRotator creates a new config rotator that periodically fetches and applies
dynamic configuration from a remote endpoint.
The rotator starts a background goroutine that fetches config at the specified interval.
If chainID or clusterID is empty, fetching is disabled.
The updater function is called whenever a new config is successfully fetched and applied.

---

### WithRenewInterval

```go
func WithRenewInterval(d time.Duration) RotatorOption
```

WithRenewInterval sets the interval between config fetch attempts. If not set, RenewInterval is used.

---

## Types

### Rotator

```go
type Rotator struct{...}
```

Rotator holds the default and current configuration and allows atomic updates.
need dynamically update config in a thread-safe manner for all p2p nodes in the cluster

**Methods:**

- `Get` - Get returns the current configuration.
- `RenewConfig` - RenewConfig atomically updates the current configuration by applying the dynamic config.

---

### RotatorOption

```go
type RotatorOption (*Rotator)
```

RotatorOption configures a Rotator (e.g. WithRenewInterval).

---

## Variables

### RenewInterval

```go
var RenewInterval = 1 * time.Minute
```
