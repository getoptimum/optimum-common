# Logger

Package `logger` provides a wrapper around Go's `slog` for application wide logging.

## Usage

```go
import "github.com/getoptimum/optimum-common/logger"

func main() {
    log := logger.NewAppSLogger("", logger.Debug, logger.WithString("module", "example"))
    log.Info("starting app")
    log.Debug("with structured fields", logger.WithInt("port", 8080))
}
```

## TODO

* pluggable logging backends
* advanced field types (e.g. duration, errors)
* context propagation helpers