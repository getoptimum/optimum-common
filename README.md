# optimum-common

[![Tests](https://github.com/getoptimum/optimum-common/actions/workflows/test.yml/badge.svg)](https://github.com/getoptimum/optimum-common/actions/workflows/test.yml)
[![Lint](https://github.com/getoptimum/optimum-common/actions/workflows/lint.yml/badge.svg)](https://github.com/getoptimum/optimum-common/actions/workflows/lint.yml)
[![Security](https://github.com/getoptimum/optimum-common/actions/workflows/security-scan.yml/badge.svg)](https://github.com/getoptimum/optimum-common/actions/workflows/security-scan.yml)
[![Coverage](https://img.shields.io/badge/coverage-81%25-brightgreen)](https://github.com/getoptimum/optimum-common)
[![GoDoc](https://pkg.go.dev/badge/github.com/getoptimum/optimum-common)](https://pkg.go.dev/github.com/getoptimum/optimum-common)

Shared Go SDK for Optimum services. Provides
configuration, networking, logging, concurrency primitives, and other
foundational utilities used across the Optimum ecosystem.

## Installation

```bash
go get github.com/getoptimum/optimum-common
```

Requires **Go 1.26.4** or later.

## Packages

| Package          | Description                                                       |
| ---------------- | ----------------------------------------------------------------- |
| `pkg/chain`      | Blockchain chain helpers for generic usage                        |
| `pkg/config`     | Runtime configuration with env/flag/YAML binding and hot-reload   |
| `pkg/entities`   | Shared domain types for P2P and dynamic config                    |
| `pkg/hash`       | Hashing utilities (SHA-256, xxHash)                               |
| `pkg/identity`   | Node identity key management                                      |
| `pkg/io`         | I/O utilities including rotating file writer                      |
| `pkg/jwks`       | JWKS key cache with disk fallback and background refresh          |
| `pkg/logger`     | Structured logging interface and helpers                          |
| `pkg/maps`       | Generic map utilities                                             |
| `pkg/math`       | Safe integer conversions with overflow protection                 |
| `pkg/net`        | Networking: external IP detection, multiaddr builder, HTTP client |
| `pkg/pointers`   | Generic pointer helpers                                           |
| `pkg/rand`       | Cryptographic random utilities                                    |
| `pkg/slices`     | Generic slice operations (unique, filter, chunk, flatten)         |
| `pkg/sql`        | SQL query builders for bulk insert and upsert                     |
| `pkg/syncx`      | Concurrency: TTL map, read-write map, broadcaster, balancer       |
| `pkg/telemetry`  | Geolocation service client                                        |
| `pkg/test_utils` | Test helpers and fixtures                                         |
| `pkg/version`    | Build version and commit metadata                                 |

## Usage

```go
import (
    "github.com/getoptimum/optimum-common/pkg/config"
    "github.com/getoptimum/optimum-common/pkg/logger"
    commonnet "github.com/getoptimum/optimum-common/pkg/net"
)

// Detect public IP (falls back to interface inspection in hermetic environments)
ipv4, ipv6, err := commonnet.GetExternalIPs()

// Safe integer conversion
import "github.com/getoptimum/optimum-common/pkg/math"
val, err := math.SafeIntToUint32(someInt)
```

## Development

The project uses a Makefile for common tasks:

```bash
make test       # Run unit tests with coverage (60% threshold)
make lint       # Run golangci-lint
make fuzz       # Run fuzz tests
make vulcheck   # Run govulncheck
make bench      # Run benchmarks
make fmt        # Format code
make vet        # Run go vet
make tidy       # Run go mod tidy
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on filing issues,
submitting pull requests, and code standards.

## Security

To report a vulnerability, see [SECURITY.md](SECURITY.md).

## Code of Conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md).
