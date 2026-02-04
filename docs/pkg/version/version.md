# Package: version

**File:** `version.go`

## Functions

### DeriveVersion

```go
func DeriveVersion(in string) string
```

DeriveVersion normalizes a version string from build info.
Handles various version formats including semantic versions, pseudo-versions, and dev builds.
Returns "dev" for unrecognized formats or pure pseudo-versions without a base tag.

---

### GetCommitHash

```go
func GetCommitHash() string
```

GetCommitHash returns the short commit hash (7 characters) from build info.
Returns "unknown" if commit information is not available.

---

### GetVersion

```go
func GetVersion() string
```

GetVersion returns the application version derived from build info.
Returns "dev" if version information is not available.
