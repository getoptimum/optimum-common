# optimum-common
This library will serve as the **shared SDK** for Optimum projects, consolidating configuration, logging, claims parsing, rate-limiting, and other cross-cutting utilities into one place.


## Optimum Common — TODO

### **Initial Scope**

- **Auth**
    - [x] Centralize JWT claims model (`Claims`)
    - [x] Parsing helpers (`ParseUnverified`, JWKS verifier)
    - [x] Remove duplicate parsing logic from CLI 
    - [x] Remove duplicate parsing logic from Proxy
    - [ ] Unit tests for claim parsing & coercion
 

- **Rate Limiting**
    - [X] Pure functions for per-sec, per-hour, daily, and message-size checks
    - [X] `UsageData` type with pluggable storage (in-memory, file-backed)
    - [X] Shared test suite to ensure consistent enforcement

- **API Models**
    - [ ] Define shared JSON request/response structs (e.g., `PublishRequest`, `SubscribeRequest`)
    - [ ] Replace per-repo definitions in CLI and Proxy

- **Versioning**
    - [ ] Header constant for `X-CLI-Version`
    - [X] Helper to add get version/commit from buildinfo omitting ldflags (`Version`, `CommitHash`)

- **HTTP Client Helpers**
    - [ ] JSON POST/GET helpers with auth header + version injection
    - [ ] Configurable timeouts and retry logic

- **Config Loader**
    - [ ] Support for flags, environment variables, and YAML config files
    - [ ] Override priority: flags > env > YAML
    - [ ] Example usage in CLI and Proxy

- **Logging**
    - [ ] Standard `AppLogger` interface with structured fields
    - [ ] Pluggable backends (Zap, Zerolog, etc.)
    - [ ] Consistent log format across services

- **Utilities**
    - [ ] IP detection (`publicIP`, `outboundIP`, private range detection)
    - [ ] Hashing helpers (XXHash, SHA-256, message ID generation)
    - [ ] TTL map, RW map/slice, broadcaster patterns
    - [ ] File helpers (atomic write, safe temp files)

- **P2P**
    - [ ] Identity key persistence/load for libp2p components
    - [ ] Shared multiaddr builders

- **Telemetry**
    - [ ] Prometheus registry creation helper
    - [ ] HTTP `/metrics` mux setup

---

## **Integration Plan**

1. **Build & publish** `optimum-common` as a standalone Go module.
2. **Replace** duplicated code in:
    - CLI (`mump2p-cli/internal/auth`, rate-limit logic, utils)
    - Proxy (`pkg/proxy/middleware/auth`, rate-limit checks, utils)
    - Gateway (`internal/utils`, IP helpers, broadcaster)
3. **Enforce imports** in all repos so shared logic lives only in `optimum-common`.
4. **Add CI tests** in `optimum-common` to cover all packages — run in downstream repos via `go test ./...` after vendoring.
5. **Document** example usage and integration steps in `/docs`.



# JWT Claims Model & Parsing

## Current Implementation

- **CLI**: `mump2p-cli/internal/auth/token.go`  
  *(TokenClaims, parse-without-verify)*
- **Proxy**: `optimum-proxy/pkg/proxy/middleware/auth.go`  
  *(Claims, parse-without-verify + JWKS verify)*

## Current Issue

Both files currently implement:

- Reading claims without verifying the signature
- Validating and extracting the same custom claims

This leads to duplicated code and potential drift between projects.

## Goal

Introduce a **shared package** that defines:

- **One claims struct** (central model)
- **Two helper functions**:
    1. Parse JWT without verification
    2. Claims validation function

_Optional_: Add dedicated tests (CLI module has some existing tests to migrate).

---

## Description of Solution

### Old Way — `auth.NewTokenParser`

- CLI had its own **TokenParser** implementation
- Proxy had a different claims parsing function
- Changes to claim names, defaults, or parsing quirks had to be applied in multiple places  
  → **High risk of inconsistency**

### New Way — `ocauth.ParseUnverified`

- Parsing logic centralized in `optimum-common/auth`
- Both CLI and Proxy import the same parsing function
- Any change applies automatically across all projects

---

## Why `ParseUnverified` Replaces Per-Repo Token Parsers

### The Old Way
Maintaining separate JWT parsers per project caused:

- **Multiple sources of truth** — Each repo defined its own claims struct and parsing logic
- **Unnecessary object creation** — Stateless `TokenParser` structs existed only to namespace a single function
- **Tight coupling to local config** — Defaults hardcoded per project
- **Incompatible types** — Claims from CLI couldn’t be used directly in Proxy or Gateway
- **Duplicate tests** — Unit tests repeated in each project

### The New Way — `optimum-common/auth.ParseUnverified`

Benefits:

- **Single source of truth** — One claims model and parser
- **No boilerplate** — Directly call a function instead of creating parser objects
- **Configurable defaults** — Set once in `optimum-common`
- **Cross-project compatibility** — Same `Claims` type everywhere
- **Centralized testing** — One test suite covers all usage

---

**Bottom line:**  
By replacing per-repo token parsers with `ParseUnverified`, Optimum removes duplication, reduces maintenance overhead, and guarantees consistent JWT parsing across all projects.
