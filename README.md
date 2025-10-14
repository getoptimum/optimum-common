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
NOTE: CI will run GoReleaser on tagged commits and publish the release automatically.
Use the Makefile helpers to manage release versions:
- `make tag-rc` creates and pushes the next `vX.Y.Z-rcN` git tag based on existing tags.
- `make release` builds and publishes artifacts for the current tag using GoReleaser.
