# Package: config

**File:** `config.go`

Package config provides a tiny loader that populates a
configuration struct from multiple sources with clear precedence:

 1. YAML file (if provided via WithYAML)
 2. Environment variables (optionally prefixed via WithEnvPrefix)
 3. Flags from a flag.FlagSet (default flag.CommandLine unless overridden via WithFlagSet)

Later sources override earlier ones (flags > env > YAML)

## Functions

### Load

```go
func Load(cfg any, opts ...Option) error
```

Load populates cfg by merging configuration from (in order):
YAML (if provided), environment variables, and flags

---

### WithEnvPrefix

```go
func WithEnvPrefix(p string) Option
```

WithEnvPrefix sets a prefix for all environment variables
WithEnvPrefix("APP") causes "HOST" to be read from "APP_HOST"

---

### WithFlagSet

```go
func WithFlagSet(fs *flag.FlagSet) Option
```

WithFlagSet sets the flag.FlagSet that will be consulted for overrides
Flags are evaluated last, giving them the highest precedence over
environment variables and YAML values

---

### WithYAML

```go
func WithYAML(path string) Option
```

WithYAML specifies the path to a YAML file to load first (lowest precedence)
If the file is missing, it is ignored

---

## Types

### Option

```go
type Option (*loader)
```

Option configures the loader used by Load
Helpers (WithFlagSet, WithYAML, and WithEnvPrefix) can be used to customize how
configuration sources are discovered and merged
