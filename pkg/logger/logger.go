package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/getoptimum/optimum-common/pkg/version"
	"github.com/libp2p/go-libp2p/core/peer"
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
	key   string
	value any
}

func (f Field) slog() slog.Attr {
	return slog.Any(f.key, f.value)
}

// AppLogger defines the structured logger interface.
type AppLogger interface {
	Info(message string, fields ...Field)
	Error(message string, err error, fields ...Field)
	Debug(message string, fields ...Field)
	Fatal(message string, err error, fields ...Field)
	With(fields ...Field) AppLogger
}

// SLogger is an implementation of AppLogger backed by slog.
type SLogger struct {
	// logWriters are immutable after construction; no mutex required.
	logWriters []*slog.Logger
}

var _ AppLogger = (*SLogger)(nil)

// NewAppSLogger creates a default AppLogger writing to stdout.
func NewAppSLogger(appHash string, mode LogMode, fields ...Field) AppLogger {
	return InitLogger([]io.Writer{os.Stdout}, appHash, validateLogMode(mode), fields...)
}

// InitLogger initializes a multi-output JSON logger with slog.
func InitLogger(writers []io.Writer, appHash string, mode LogMode, fields ...Field) AppLogger {
	logs := make([]*slog.Logger, 0, len(writers))
	for _, w := range writers {
		handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: logLevelFromMode(mode),
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				switch a.Key {
				case "time":
					return slog.Int64("timestamp", time.Now().Unix())
				case "level":
					return slog.String("_level", strings.ToLower(a.Value.String()))
				case "gray_log_level":
					return slog.Int64("level", a.Value.Int64())
				case "msg":
					return slog.String("short_message", a.Value.String())
				}
				return a
			},
		})
		attrs := make([]any, 0, len(fields)+2)
		if appHash != "" && len(appHash) > 6 {
			attrs = append(attrs, slog.String("version", appHash[:6]))
		}
		for _, f := range fields {
			attrs = append(attrs, f.slog())
		}
		if gitHash := version.GetCommitHash(); gitHash != "" {
			attrs = append(attrs, slog.String("commit", gitHash))
		}

		lw := slog.New(handler).With(attrs...)
		logs = append(logs, lw)
	}
	return &SLogger{logWriters: logs}
}

// Info logs an informational message with optional structured fields.
func (l *SLogger) Info(message string, fields ...Field) {
	params := prepareSlogParams(nil, fields)
	l.processWriters(func(lg *slog.Logger) {
		lg.Info(message, params...)
	})
}

// Error logs an error message with associated error and fields.
func (l *SLogger) Error(message string, err error, fields ...Field) {
	params := prepareSlogParams(err, fields)
	l.processWriters(func(lg *slog.Logger) {
		lg.Error(message, params...)
	})
}

// Debug logs a debug message.
func (l *SLogger) Debug(message string, fields ...Field) {
	params := prepareSlogParams(nil, fields)
	l.processWriters(func(lg *slog.Logger) {
		lg.Debug(message, params...)
	})
}

// Fatal logs an error and exits the application.
func (l *SLogger) Fatal(message string, err error, fields ...Field) {
	params := prepareSlogParams(err, fields)
	l.processWriters(func(lg *slog.Logger) {
		lg.Error(message, params...)
	})
	l.Info("fatal error, exiting")
	os.Exit(1)
}

// With creates a child logger with additional structured context.
func (l *SLogger) With(fields ...Field) AppLogger {
	logs := make([]*slog.Logger, 0, len(l.logWriters))
	for _, lg := range l.logWriters {
		logs = append(logs, lg.With(prepareSlogParams(nil, fields)...))
	}
	return &SLogger{
		logWriters: logs,
	}
}

func prepareSlogParams(err error, fields []Field) []any {
	params := make([]any, 0, len(fields)+2)
	if err != nil {
		params = append(params, slog.String("error", err.Error()))
	}
	for _, f := range fields {
		params = append(params, f.slog())
	}
	return params
}

func (l *SLogger) processWriters(processor func(*slog.Logger)) {
	var wg sync.WaitGroup
	wg.Add(len(l.logWriters))
	for i := range l.logWriters {
		go func(j int) {
			processor(l.logWriters[j])
			wg.Done()
		}(i)
	}
	wg.Wait()
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

// --- Attribute helpers ---

func WithAny(key string, val any) Field {
	return Field{key: key, value: val}
}

func WithString(key, val string) Field {
	return Field{key: key, value: val}
}

func WithUint64(key string, val uint64) Field {
	return Field{key: key, value: val}
}

func WithBool(key string, val bool) Field {
	return Field{key: key, value: val}
}

func WithInt64(key string, val int64) Field {
	return Field{key: key, value: val}
}

func WithInt(key string, val int) Field {
	return Field{key: key, value: val}
}

// --- Generic helpers ---

func WithModule(val string) *Field {
	f := WithString("module", val)
	return &f
}

func WithError(err error) *Field {
	f := WithString("err", err.Error())
	return &f
}

func WithFilePath(filePath string) Field {
	return WithString("file", filePath)
}

func WithService(serviceName string) Field {
	return WithString("service", serviceName)
}

// WithPeer adds the peer ID of the given AddrInfo.
func WithPeer(info peer.AddrInfo) Field {
	return WithString("peer", info.ID.String())
}

// WithPeerAddrs adds comma separated multiaddrs of the given peer.
func WithPeerAddrs(info peer.AddrInfo) Field {
	addrs := make([]string, 0, len(info.Addrs))
	for _, addr := range info.Addrs {
		addrs = append(addrs, addr.String())
	}
	return WithString("peer", strings.Join(addrs, ","))
}

// WithTopic adds a pubsub topic as a string field.
func WithTopic(topic string) Field {
	return WithString("topic", topic)
}

// WithTopicBytes adds a topic represented as byte slice.
func WithTopicBytes(topic []byte) Field {
	return WithString("topic", string(topic))
}
