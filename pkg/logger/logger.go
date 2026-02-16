package logger

import (
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/getoptimum/optimum-common/pkg/version"
)

// LogMode defines logging verbosity levels.
type LogMode string

const (
	Production LogMode = "production"
	Debug      LogMode = "debug"
	Verbose    LogMode = "verbose"
)

func validateLogMode(mode LogMode) LogMode {
	switch mode {
	case Production, Debug, Verbose:
		return mode
	default:
		return Production
	}
}

// Field represents a typed key-value pair for structured logging.
type Field struct {
	a slog.Attr
}

// AppLogger defines the structured logger interface.
//
// Deprecated: The AppLogger interface is deprecated and will be removed in a
// future release. Use slog.Logger directly instead with NewMultiHandler if
// multiple handlers are needed. The logging package may still exist and provide
// a simple helper for a standardized slog logger creation.
type AppLogger interface {
	Info(message string, fields ...Field)
	Error(message string, err error, fields ...Field)
	Debug(message string, fields ...Field)
	Fatal(message string, err error, fields ...Field)
	With(fields ...Field) AppLogger
}

// SLogger is an implementation of AppLogger backed by slog.
type SLogger struct {
	logger *slog.Logger
}

var _ AppLogger = (*SLogger)(nil)

// NewAppSLogger creates a default AppLogger writing to STDERR.
func NewAppSLogger(mode LogMode, fields ...Field) AppLogger {
	return InitLogger([]io.Writer{os.Stderr}, validateLogMode(mode), fields...)
}

// NewAppSLoggerFromSLog returns a new AppLogger from an existing slog.Logger.
func NewAppSLoggerFromSLog(logger *slog.Logger) AppLogger {
	if logger == nil {
		panic("logger is nil")
	}

	return &SLogger{logger: logger}
}

// InitLogger initializes a multi-output JSON logger with slog. At least one writer
// must be provided.
func InitLogger(writers []io.Writer, mode LogMode, fields ...Field) AppLogger {
	handlers := make([]slog.Handler, len(writers))
	for i, w := range writers {
		handlers[i] = slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: logLevelFromMode(mode),
			ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
				switch a.Key {
				case "time":
					return slog.Int64("timestamp", time.Now().Unix())

				case "gray_log_level":
					return slog.Int64("level", a.Value.Int64())
				}

				return a
			},
		})
	}

	attrs := make([]any, 0, len(fields)+2)
	for _, f := range fields {
		attrs = append(attrs, f.a)
	}

	attrs = append(attrs,
		slog.String("commit", version.GetCommitHash()),
		slog.String("version", version.GetVersion()),
	)

	mh := slog.NewMultiHandler(handlers...)
	return &SLogger{logger: slog.New(mh).With(attrs...)}
}

// Info logs an informational message with optional structured fields.
func (l *SLogger) Info(message string, fields ...Field) {
	params := prepareSlogParams(nil, fields)
	l.logger.Info(message, params...)
}

// Error logs an error message with associated error and fields.
func (l *SLogger) Error(message string, err error, fields ...Field) {
	params := prepareSlogParams(err, fields)
	l.logger.Error(message, params...)
}

// Debug logs a debug message.
func (l *SLogger) Debug(message string, fields ...Field) {
	params := prepareSlogParams(nil, fields)
	l.logger.Debug(message, params...)
}

// Fatal logs an error and exits the application.
func (l *SLogger) Fatal(message string, err error, fields ...Field) {
	params := prepareSlogParams(err, fields)
	l.logger.Error(message, params...)
	l.logger.Info("fatal error; exiting")
	os.Exit(1)
}

// With creates a child logger with additional structured context.
func (l *SLogger) With(fields ...Field) AppLogger {
	return &SLogger{logger: l.logger.With(prepareSlogParams(nil, fields)...)}
}

func prepareSlogParams(err error, fields []Field) []any {
	params := make([]any, 0, len(fields)+2)
	if err != nil {
		params = append(params, slog.String("error", err.Error()))
	}

	for _, f := range fields {
		params = append(params, f.a)
	}

	return params
}

func logLevelFromMode(mode LogMode) slog.Level {
	switch mode {
	case Debug:
		return slog.LevelDebug
	case Verbose:
		return slog.LevelInfo
	case Production:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}
