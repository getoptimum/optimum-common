# Package: utils

**File:** `rotating_file_writer.go`

## Functions

### NewRotatingFileWriter

```go
func NewRotatingFileWriter(filePath string, maxSize int64) (*RotatingFileWriter, error)
```

NewRotatingFileWriter creates a new RotatingFileWriter
filePath: path to the log file (e.g., "/proxy/logs/proxy-debug.log" or "./logs/gateway-debug.log")
maxSize: maximum file size in bytes before rotation (default: 10MB)

---

## Types

### RotatingFileWriter

```go
type RotatingFileWriter struct{...}
```

RotatingFileWriter is an io.Writer that rotates files when they exceed a size limit

**Methods:**

- `Close` - Close implements io.Closer
- `Write` - Write implements io.Writer

---

## Constants

### DefaultMaxFileSize

```go
const DefaultMaxFileSize = 10 * 1024 * 1024
```

---

### MaxRotatedFiles

```go
const MaxRotatedFiles = 5
```
