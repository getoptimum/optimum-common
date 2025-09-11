# Logger

Package `logger` provides a wrapper around Go's `slog` for application wide logging.

## Usage

```go
package main

import (
    "fmt"

    "github.com/getoptimum/optimum-common/pkg/logger"
)

func main() {
    log := logger.NewAppSLogger("", logger.Debug, logger.WithString("module", "example"))
    log.Info("starting app", logger.WithInt("port", 8080))
    log.Error("something failed", fmt.Errorf("bad"), logger.WithString("user", "alice"))
}
```

Produces output similar to:

```json
{"timestamp":1700000000,"_level":"info","short_message":"starting app","module":"example","port":8080}
{"timestamp":1700000001,"_level":"error","short_message":"something failed","module":"example","error":"bad","user":"alice"}
```

`Production`  raises the minimum level to `WARN` 
`Debug` enables debug messages:

```go
logger.NewAppSLogger("", logger.Production).Info("hidden") // not logged
logger.NewAppSLogger("", logger.Debug).Debug("visible")   // logged
```