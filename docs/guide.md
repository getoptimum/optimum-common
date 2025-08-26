# Version package

The `version` package provides helpers to expose build metadata for an application.
It reads the Go module build information via `debug.ReadBuildInfo` and derives
stable values without requiring `-ldflags`.

## API

- `GetVersion()` – returns the version of the module or `dev` if no
  tag can be determined.
- `GetCommitHash()` – returns the abbreviated commit hash (7 characters) or
  `unknown` when the hash is unavailable (or shorter if truncated available).
- `DeriveVersion(in string)` – parses raw module versions and resolves Go
  pseudo-version patterns into a human-friendly form.

## Usage

```go
import "github.com/getoptimum/optimum-common/version"

func main() {
    fmt.Printf("Version: %s (commit %s)\n", version.GetVersion(), version.GetCommitHash())
}
```

## Pseudo-version handling

`DeriveVersion` strips build metadata and normalises common Go pseudo-version
formats:

| Input example                                    | Output    |
|--------------------------------------------------|-----------|
| `v1.2.3`                                         | `v1.2.3`  |
| `v1.2.3-rc1`                                     | `v1.2.3-rc1` |
| `v1.2.3-0.20240102112233-deadbeef`              | `v1.2.3`  |
| `v0.0.0-20240102112233-deadbeef`                | `dev`     |

Unknown or malformed versions fall back to `dev`, and commit hashes fall back
to `unknown`.