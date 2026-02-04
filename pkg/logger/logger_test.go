package logger_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/logger"
	"github.com/getoptimum/optimum-common/pkg/test_utils"
	"github.com/getoptimum/optimum-common/pkg/version"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func parseJSONLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")

	// Preallocate capacity to avoid repeated growth (prealloc)
	out := make([]map[string]any, 0, len(lines))

	for _, ln := range lines {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		var m map[string]any
		require.NoError(t, json.Unmarshal([]byte(ln), &m))
		out = append(out, m)
	}
	return out
}

func getNumber(m map[string]any, key string) float64 {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}

func TestInitLogger_JSONShapeAndMappings(t *testing.T) {
	var buf bytes.Buffer

	l := logger.InitLogger([]io.Writer{&buf}, logger.Debug,
		logger.WithString("svc", "api"),
	)
	l.Info("hello world",
		logger.WithInt("n", 3),
		logger.WithInt64("gray_log_level", 5), // should be remapped
	)

	recs := parseJSONLines(t, &buf)
	require.Len(t, recs, 1)
	r := recs[0]

	require.Equal(t, "hello world", r["short_message"])
	require.Equal(t, "info", r["_level"])
	require.Equal(t, version.GetVersion(), r["version"])
	require.Equal(t, "api", r["svc"])
	require.EqualValues(t, 3, getNumber(r, "n"))
	require.EqualValues(t, 5, getNumber(r, "level"))
}

func TestLogLevelFiltering_PerMode(t *testing.T) {
	var bp bytes.Buffer
	lp := logger.InitLogger([]io.Writer{&bp}, logger.Production)
	lp.Info("info-suppressed")
	lp.Error("err-shown", errors.New("boom"))
	require.Equal(t, 1, len(parseJSONLines(t, &bp)))

	var bv bytes.Buffer
	lv := logger.InitLogger([]io.Writer{&bv}, logger.Verbose)
	lv.Debug("debug-suppressed")
	lv.Info("info-shown")
	require.Equal(t, 1, len(parseJSONLines(t, &bv)))

	var bd bytes.Buffer
	ld := logger.InitLogger([]io.Writer{&bd}, logger.Debug)
	ld.Debug("debug-shown")
	ld.Info("info-shown")
	require.Equal(t, 2, len(parseJSONLines(t, &bd)))
}

func TestLogger_Fatal_ExitsWithCode1AndLogs(t *testing.T) {
	// #nosec G204 -- test re-invokes self with constant arguments; no user input involved
	cmd := exec.Command(os.Args[0], "-test.run", "^TestHelperInvokeFatal$")
	cmd.Env = append(os.Environ(), "INVOKE_FATAL=1")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	require.Error(t, err)

	var ee *exec.ExitError
	require.True(t, errors.As(err, &ee), "expected *exec.ExitError")
	require.Equal(t, 1, ee.ExitCode())

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	require.GreaterOrEqual(t, len(lines), 2)
	found := false
	for _, ln := range lines {
		var m map[string]any
		_ = json.Unmarshal([]byte(ln), &m)
		if m["short_message"] == "fatal error, exiting" {
			found = true
		}
	}
	require.True(t, found)
}

func TestHelperInvokeFatal(t *testing.T) {
	if os.Getenv("INVOKE_FATAL") != "1" {
		t.Skip("helper")
	}
	var buf bytes.Buffer
	l := logger.InitLogger([]io.Writer{&buf, os.Stdout}, logger.Debug)
	l.Fatal("boom", errors.New("fatal")) // os.Exit(1)
}

func Test_SLogger_purelog_with_stdout(t *testing.T) {
	// given
	appLog := test_utils.NewTestLogger(t, func(w []io.Writer) logger.AppLogger {
		return logger.InitLogger(w, logger.Debug)
	})
	appLog.Info("test")
	appLog.Error("source log", fmt.Errorf("123"))

	// when, then
	concurrentlyLogIt(
		appLog.With(
			logger.WithString("a", "b"),
			logger.WithString("c", "d"),
			logger.WithAny("e", "f"),
			logger.WithAny("e2", map[string]string{
				"k1": "v1",
				"k2": "v2",
				"k3": "v3",
			}),
		),
	)
	t.Run("pure check", func(t *testing.T) {
		l := test_utils.NewTestLogger(t, func(w []io.Writer) logger.AppLogger {
			return logger.InitLogger(w, logger.Debug)
		})
		l.Error("source log", fmt.Errorf("123"))
	})
}

func Test_DefaultLogger(t *testing.T) {
	// given
	appLog := test_utils.NewTestLogger(t, func(w []io.Writer) logger.AppLogger {
		return logger.InitLogger(w, logger.Debug)
	})
	appLog.Info("test")
	appLog.Error("source log", fmt.Errorf("123"))

	// when, then
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			concurrentlyLogIt(
				appLog.With(
					logger.WithString("a", "b"),
					logger.WithString("c", "d"),
				),
			)
		}()
		go func() {
			defer wg.Done()
			defaultAppLog := logger.NewAppSLogger("debug")
			concurrentlyLogIt(
				defaultAppLog.With(
					logger.WithString("a", "b"),
					logger.WithString("c", "d"),
				),
			)
		}()
	}
	wg.Wait()
}

func Test_SLogger_multiply_writers(t *testing.T) {
	// given
	appLog := test_utils.NewTestLogger(t, func(w []io.Writer) logger.AppLogger {
		return logger.InitLogger(w, logger.Debug)
	})
	// when, then
	concurrentlyLogIt(appLog)

	t.Run("pure check", func(t *testing.T) {
		l := logger.NewAppSLogger(
			"debug",
			logger.WithString("additional", "value"),
			logger.WithString("additional2", "value"),
		)
		l.Info("test")
	})
}

func concurrentlyLogIt(appLog logger.AppLogger) {
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				appLog.Info("message")
				appLog.Error("message", fmt.Errorf("error"))
				appLog.Error("message", fmt.Errorf("error"), logger.WithString("key", uuid.NewString()))
			}
		}()
	}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			customLogger1 := appLog.With(logger.WithString("keyA", uuid.NewString()))
			customLogger1.Info("customLogger message")
			customLogger1.Error("customLogger message", fmt.Errorf("customLogger error"))
			customLogger1.Error("customLogger message", fmt.Errorf("customLogger error"), logger.WithString("keyB", uuid.NewString()))
		}()
	}
	wg.Wait()
}
