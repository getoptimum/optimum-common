# Package: logger

**File:** `logger.go`

## Functions

### InitLogger

```go
func InitLogger(writers []io.Writer, mode LogMode, fields ...Field) AppLogger
```

InitLogger initializes a multi-output JSON logger with slog.

---

### NewAppSLogger

```go
func NewAppSLogger(mode LogMode, fields ...Field) AppLogger
```

NewAppSLogger creates a default AppLogger writing to stdout.

---

## Types

### AppLogger

```go
type AppLogger interface{...}
```

AppLogger defines the structured logger interface.

---

### Field

```go
type Field struct{...}
```

Field represents a typed key-value pair for structured logging.

---

### LogMode

```go
type LogMode string
```

LogMode defines logging verbosity levels.

---

### SLogger

```go
type SLogger struct{...}
```

SLogger is an implementation of AppLogger backed by slog.

**Methods:**

- `Debug` - Debug logs a debug message.
- `Error` - Error logs an error message with associated error and fields.
- `Fatal` - Fatal logs an error and exits the application.
- `Info` - Info logs an informational message with optional structured fields.
- `With` - With creates a child logger with additional structured context.

---

## Constants

### Debug

```go
const Debug = "debug"
```

---

### Production

```go
const Production = "production"
```

---

### Verbose

```go
const Verbose = "verbose"
```
