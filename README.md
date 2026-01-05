# optimum-common

This library will serve as the **shared SDK** for Optimum projects.

## High-level structure

- `optimum-common/` contains a Go module that acts as a shared SDK for other Optimum services,
consolidating logging, configuration, utilities, and version helpers used across projects.

## Key packages in `optimum-common`

- `pkg/config` implements a configurable loader that merges YAML, environment variables, and CLI flags, with deterministic precedence and automatic struct field discovery for overrides.
- `pkg/logger` wraps Go's `slog` with multi-writer support, contextual fields, automatic version/commit tagging, and concurrency-safe fan-out to multiple structured loggers.
- `pkg/utils` offers shared helpers and utility primitives (hashing, random IDs, request helpers, etc.).
- `pkg/test_utils` provides testing adapters like `TestLogWriter` and `NewTestLogger` to capture structured logs during CI runs while mirroring them locally.
- `pkg/version` derives semantic versions and short commit hashes from Go build metadata, normalizes various pseudo-version formats into stable identifiers.

## Versioning and Releases

Use the Makefile helpers to manage release versions:

- `make tag-rc` creates and pushes the next `vX.Y.Z-rcN` release candidate tag.
- `make release` creates a GitHub release using GoReleaser (requires existing tag).

Consumers can use tagged versions in their `go.mod`:

```go
require github.com/getoptimum/optimum-common v0.0.1-rc1
```

## Standards followed

1. Tests employ `require` package, must be table driven where possible
2. Comments in tests follow Given/When/Then structure when possible
3. Comments are capitalized, in full sentences ending with a period.
