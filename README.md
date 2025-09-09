# optimum-common
This library will serve as the **shared SDK** for Optimum projects.

Main goal:
Create a reusable internal SDK to standardize configuration loading, logging, and common utilities across all Optimum projects.
## Optimum Common — Progress
- [x] Inject `Version` and `CommitHash` from BuildInfo not -ldflags
    note: embedding version info in requests, version cheching handler
- [ ] Common utility functions (IP detection, hashing, TTLMap, file helpers)
- [ ] Config loader that supports flags, environment variables, and YAML with override priority
- [ ] Standard `AppLogger` interface with structured fields
- [ ] Prepare example usage and integration guide

---


## Tasks better described
### Extracting common utility functions includes:
- Duplicate IP detection logic
    - optimum-p2p/pkg/utils/ip.go
    - optimum-proxy/pkg/utils/ip.go
    - optimum-gateway/internal/utils/ip.go
- Duplicate hashing utilities
    - optimum-p2p/pkg/utils/hash.go
    - optimum-gateway/internal/utils/hash.go
- Multiple TTLMap implementations
    - optimum-p2p/pkg/utils/ttl_map.go
    - optimum-proxy/pkg/utils/ttl_map.go
    - optimum-gateway/internal/utils/ttl_map.go
 - Duplicate file helper functions
    - optimum-p2p/pkg/utils/file.go
    - optimum-proxy/pkg/utils/file.go
    - optimum gateway reuses the p2p implementation?

### Add standard AppLogger interface
 - shared logger abstraction in optimum-common
 - current StringWith structs only carry string values, forcing all fields through string conversion and losing type information
 - both optimum-p2p and optimum-proxy embed nearly identical SLogger implementations

### Add config loader with overrides support
- Add a new configuration loader that reads settings from YAML, environment variables, and command-line flags, applying overrides in the order of flags > environment > YAML
- Document the completed configuration loader capabilities