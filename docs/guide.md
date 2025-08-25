# Version package
Here is a professional Markdown draft for the documentation of the updated version package using debug.ReadBuildInfo() and regex parsing, including a suggestion for a short README snippet for downstream repos. It is ready to be dropped into repo docs or README sections:


# `optimum-common/version`

The `version` package provides a standardized way for all Optimum projects to
embed and expose build metadata (`Version`, `LastCommitHashShort`) without
relying on `-ldflags`.

---

## Overview

Traditionally, Optimum repos injected version information at build time using
Makefile `-ldflags`, for example:

```
VERSION      ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
COMMIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

go build -ldflags "\
-X github.com/getoptimum/optimum-common/version.Version=$(VERSION) \
-X github.com/getoptimum/optimum-common/version.CommitHash=$(COMMIT_HASH)"
```

This introduced boilerplate, duplication across repos, and brittle coupling
between build tooling and runtime code.  
Starting in Go 1.18, the toolchain automatically embeds VCS metadata
(`vcs.revision`, `vcs.time`, `vcs.modified`) into binaries. The updated
version package consumes this metadata via `debug.ReadBuildInfo()` and
derives stable values without any ldflags.

---

## API

### Globals

```
// Version is the last semver-like tag, or "dev" if none.
var Version string = "dev"

// LastCommitHashShort is the 7-char abbreviated commit hash if available.
var LastCommitHashShort string = "unknown"
```

These are initialized automatically at `init()` time.

### Internal Functions

- `deriveVersion(in string) string`  
  Parses the Go module’s Main.Version, stripping build metadata and handling
  pseudo-versions:  
  - Returns the tag/prerelease base for forms like `v1.2.3-rc4.0.<timestamp>-<hash>`  
    or `v1.2.3-0.<timestamp>-<hash>`.  
  - Returns `"dev"` for `v0.0.0-<timestamp>-<hash>` (pure pseudo).  
  - Preserves unknown 8-digit tails like `v1.2.3-20250821-abcdef`.  
  - Returns the input if a valid semver tag (e.g., `v1.2.3`, `v1.2.3-rc4`).  
  - Falls back to `"dev"` for malformed or `(devel)` cases.

- `stripBuildMeta(v string) string`  
  Removes suffixes like `+dirty` or `+meta`.

- `vcsRevision(bi *debug.BuildInfo) string`  
  Returns the full `vcs.revision` hash if present.

- `short(h string) string`  
  Abbreviates a commit hash to 7 characters.

---

## Behavior

| Input                                            | Output Version  | Output LastCommitHashShort |
|-------------------------------------------------|-----------------|---------------------------|
| `v1.2.3`                                        | `v1.2.3`        | `unknown`                 |
| `v1.2.3-rc4`                                   | `v1.2.3-rc4`    | `unknown`                 |
| `v0.0.1-rc4.0.20250821142859-6128ae7a7356`    | `v0.0.1-rc4`    | `6128ae7`                 |
| `v1.2.3-0.20240102112233-abcdef1`              | `v1.2.3`        | `abcdef1`                 |
| `v0.0.0-20250821142859-deadbeef`               | `dev`           | `deadbee`                 |
| `(devel)`                                       | `dev`           | `unknown`                 |

---

## Usage

Import the package and reference the globals:

```
import "github.com/getoptimum/optimum-common/version"

func main() {
fmt.Printf("Version: %s (commit %s)\n",
version.Version, version.LastCommitHashShort)
}
```

At build time, simply build without any ldflags:
```aiignore

Code snippet here
```

---

## Migration Guide

### Before (with ldflags):

```
VERSION      ?= $(shell git describe --tags --abbrev=0)
COMMIT_HASH  ?= $(shell git rev-parse --short HEAD)

go build -ldflags "\
-X github.com/getoptimum/optimum-common/version.Version=$(VERSION) \
-X github.com/getoptimum/optimum-common/version.CommitHash=$(COMMIT_HASH)"
```

### After (no ldflags required):


go build ./cmd/cli

The runtime automatically derives version and commit info from the Go module.

---

## Benefits

- No boilerplate in Makefiles or CI pipelines.
- Consistent parsing of pseudo-versions across all repos.
- Portable across environments (local dev, CI, tagged releases).
- Aligned with Go toolchain (uses embedded VCS metadata).
- Graceful fallbacks (`dev` / `unknown`) when metadata is missing.

---

## Suggested README snippet for downstream repos

```
# Using `optimum-common/version` without ldflags

Previously, version and commit info was injected via ldflags in builds:

```makefile
go build -ldflags " \
    -X github.com/getoptimum/optimum-common/version.Version=$(VERSION) \
    -X github.com/getoptimum/optimum-common/version.CommitHash=$(COMMIT_HASH)"
```

With the updated version package, no ldflags are needed. Simply build as usual:

```bash
go build ./cmd/cli
```

Access version information at runtime:

```go
import "github.com/getoptimum/optimum-common/version"

func main() {
    fmt.Printf("Version: %s (commit %s)\n", version.Version, version.LastCommitHashShort)
}
```

This approach reduces build complexity, aligns with Go toolchain conventions,
and ensures consistent version info across all Optimum repos.

# Authentication Integration

The Optimum projects now share a single authentication implementation provided by
[`optimum-common/auth`](./optimum-common/auth).  The module offers:

- A unified `Claims` model used by all services and clients.
- `ParseUnverified` for reading JWTs without validating the signature (used by
  CLI tools when inspecting locally stored tokens).
- `Verifier` which validates tokens against Auth0 JWKS with optional issuer and
  audience checks.

## Repository changes

### mump2p-cli
- Replaced custom `TokenClaims` and token parser with wrappers around the shared
  module.
- CLI commands now call `auth.ParseUnverified` to inspect tokens.
- Rate limiter operates directly on the common `Claims` type.

### optimum-proxy
- Removed bespoke JWT parsing and JWKS handling in the middleware.
- Middleware and API handlers rely on `auth.Verifier` and `auth.Claims`.
- WebSocket helpers parse tokens via `auth.ParseUnverified` when auth is
  disabled.

### optimum-p2p
- Deleted local JWKS cache and token parser; the service uses
  `auth.NewVerifierFromDomain` for validation and `auth.ParseUnverified` for
  claim extraction.
- Usage tracking works with the shared `Claims` structure.

### optimum-gateway
- No duplicate authentication logic existed.  The project depends on
  `optimum-common` and can use the shared module directly when authentication is
  enabled.

## Example usage

```go
// Create a verifier for services (e.g. proxy or p2p)
v, _ := auth.NewVerifierFromDomain(domain, audience, nil)
claims, err := v.Verify(tokenString)

// Parse without verification (e.g. CLI)
claims, err := auth.ParseUnverified(tokenString)
```

## Migration notes

- Each repository imports `github.com/getoptimum/optimum-common/auth`.
- Existing environment variables for Auth0 domain and audience remain
  unchanged.
- Any custom JWT parsing or claim structs can be removed in favour of the
  shared package.