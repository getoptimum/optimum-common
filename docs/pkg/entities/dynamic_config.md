# Package: entities

**File:** `dynamic_config.go`

## Functions

### FromYAMLFile

```go
func FromYAMLFile(path string) (*DynamicConfig, error)
```

FromYAMLFile loads a DynamicConfig from a YAML file at the specified path.
Returns an error if the file cannot be opened or parsed.

---

### HashRemoteConfig

```go
func HashRemoteConfig(cfg *DynamicConfig) string
```

HashRemoteConfig computes a SHA-256 hash of the configurable fields in DynamicConfig.
Used to detect when remote configuration has changed.
The hash includes: EnableABTesting, ExcludeSelfMessages, and all RLNC/mesh parameters.

---

## Types

### DynamicConfig

```go
type DynamicConfig struct{...}
```

DynamicConfig represents runtime-configurable parameters for P2P network behavior.
These settings can be updated without restarting the node.

**Methods:**

- `ToMap` - ToMap converts the DynamicConfig to a map representation suitable for storage or serialization.
- `Validate` - Validate checks that all configuration values are within valid ranges.
