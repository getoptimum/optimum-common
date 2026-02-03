package test_utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/utils"
	"github.com/stretchr/testify/require"
)

// GetFreePortT is a wrapper around GetFreePort
func GetFreePortT(t *testing.T) int {
	t.Helper()
	port, err := utils.GetFreePort()
	require.NoError(t, err, "failed to allocate free TCP port")
	return port
}

// TestLogWriter captures slog output during CI runs while mirroring logs to stdout locally
type TestLogWriter struct {
	std io.Writer

	logs     bytes.Buffer
	logsChan chan []byte

	stop chan struct{}
	once sync.Once
	wg   sync.WaitGroup
}

// NewTestLogWriter creates a writer suitable for wiring into structured loggers in tests
// Mirrors all output to the provided std writer (stdout by default) during local runs
// When the CI_RUN environment variable is set, the output is buffered in-memory and printed only if the test fails
func NewTestLogWriter(t testing.TB) *TestLogWriter {
	t.Helper()
	return NewTestLogWriterWithOutput(t, os.Stdout)
}

// NewTestLogWriterWithOutput behaves like NewTestLogWriter but allows overriding the underlying writer used for local runs
func NewTestLogWriterWithOutput(t testing.TB, std io.Writer) *TestLogWriter {
	t.Helper()

	tl := &TestLogWriter{
		std:      std,
		logsChan: make(chan []byte, 1_024),
		stop:     make(chan struct{}),
	}

	tl.wg.Add(1)
	go tl.process()

	t.Cleanup(func() {
		tl.close()
		if os.Getenv("CI") != "" && t.Failed() {
			fmt.Print(tl.logs.String())
		}
	})

	return tl
}

func (tl *TestLogWriter) Write(p []byte) (int, error) {
	if os.Getenv("CI") != "" {
		buf := append([]byte(nil), p...) // slog reuses buffers; copy to avoid races
		select {
		case tl.logsChan <- buf:
		default:
			// keep the channel bounded while ensuring we never block the logger
			// if the buffer is full we drop the message but keep the overall run alive
		}
		return len(p), nil
	}

	return tl.std.Write(p)
}

func (tl *TestLogWriter) process() {
	defer tl.wg.Done()

	for {
		select {
		case <-tl.stop:
			return
		case data := <-tl.logsChan:
			tl.logs.Write(data)
		}
	}
}

func (tl *TestLogWriter) close() {
	tl.once.Do(func() {
		close(tl.stop)
		tl.wg.Wait()
	})
}

// NewTestLogger wires the TestLogWriter into a repository specific logger factory
//
// Example:
//
//	logger := testutils.NewTestLogger(t, func(w []io.Writer) logger.AppLogger {
//	    return logger.InitLogger(w, logger.Debug)
//	})
func NewTestLogger[T any](t testing.TB, factory func([]io.Writer) T) T {
	t.Helper()

	writer := NewTestLogWriter(t)
	return factory([]io.Writer{writer})
}

// RunConcurrently runs fn from nGoroutines goroutines, nOps times per goroutine, and waits for completion.
// Useful for stress-testing concurrent access (e.g. RWMap, TTLMap).
func RunConcurrently(nGoroutines, nOps int, fn func()) {
	var wg sync.WaitGroup
	for i := 0; i < nGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < nOps; j++ {
				fn()
			}
		}()
	}
	wg.Wait()
}
