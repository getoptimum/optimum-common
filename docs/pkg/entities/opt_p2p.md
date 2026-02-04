# Package: entities

**File:** `opt_p2p.go`

## Types

### OptimumConfig

```go
type OptimumConfig struct{...}
```

OptimumConfig holds configuration for P2P network settings including
cluster/chain identifiers, message size limits, RLNC sharding parameters,
mesh topology settings, and bootstrap peer addresses.

**Methods:**

- `ApplyDynamicConfig` - ApplyDynamicConfig creates a new OptimumConfig by applying dynamic configuration
- `Clone` - Clone creates a deep copy of the configuration.
- `Validate` - Validate checks that required fields are set and values are within valid ranges.
