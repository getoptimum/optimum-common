package test_utils_test

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/test_utils"
	"github.com/stretchr/testify/require"
)

// ensure valid and bindable tcp ports returned
func TestGetFreePortFunctions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		port func(t *testing.T) int
	}{
		{
			name: "GetFreePort",
			port: func(t *testing.T) int {
				t.Helper()
				p, err := test_utils.GetFreePort()
				require.NoError(t, err)
				return p
			},
		},
		{
			name: "GetFreePortT",
			port: test_utils.GetFreePortT,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			port := tc.port(t)
			require.Greater(t, port, 0)

			ensurePortBindable(t, port)
		})
	}
}

func ensurePortBindable(t *testing.T, port int) {
	t.Helper()

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, listener.Close())
	})
}

type panicWriter struct {
	called bool
}

func (pw *panicWriter) Write([]byte) (int, error) {
	pw.called = true
	return 0, fmt.Errorf("unexpected write")
}

func TestNewTestLogWriterLocalMirrorsOutput(t *testing.T) {
	t.Setenv("CI", "")

	var captured bytes.Buffer
	writer := test_utils.NewTestLogWriterWithOutput(t, &captured)

	const message = "local log line"
	n, err := writer.Write([]byte(message))
	require.NoError(t, err)
	require.Equal(t, len(message), n)
	require.Equal(t, message, captured.String())
}

func TestNewTestLogWriterCIBuffersLogs(t *testing.T) {
	t.Setenv("CI", "1")

	pw := &panicWriter{}
	writer := test_utils.NewTestLogWriterWithOutput(t, pw)

	const message = "ci log line"
	n, err := writer.Write([]byte(message))
	require.NoError(t, err)
	require.Equal(t, len(message), n)
	require.False(t, pw.called, "write should not reach the underlying writer when CI is set")
}

type stubLogger struct {
	writers []io.Writer
}

func TestNewTestLoggerIntegratesWithFactory(t *testing.T) {
	t.Parallel()

	logger := test_utils.NewTestLogger[*stubLogger](t, func(writers []io.Writer) *stubLogger {
		return &stubLogger{writers: writers}
	})

	require.Len(t, logger.writers, 1)
	_, ok := logger.writers[0].(*test_utils.TestLogWriter)
	require.True(t, ok)
}
