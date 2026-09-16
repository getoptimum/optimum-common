package sql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubBackoff collapses the real (multi-second) schedule so tests stay fast.
// Returns a restore func rather than using t.Cleanup so callers can stay explicit.
func stubBackoff(t *testing.T) {
	t.Helper()
	orig := readReplicaBackoff
	readReplicaBackoff = func(int) time.Duration { return time.Microsecond }
	t.Cleanup(func() { readReplicaBackoff = orig })
}

func TestRetryReadReplica(t *testing.T) {
	stubBackoff(t)

	calls := 0
	err := retryReadReplica(context.Background(), func() error {
		calls++
		if calls < 2 {
			return &pq.Error{Code: "40001"}
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 2, calls)

	calls = 0
	sentinel := errors.New("other error")
	err = retryReadReplica(context.Background(), func() error {
		calls++
		return sentinel
	})
	require.ErrorIs(t, err, sentinel)
	assert.Equal(t, 1, calls)
}

// A conflict that never clears must exhaust the full schedule and surface the last error,
// rather than giving up after the old ~600ms budget.
func TestRetryReadReplicaExhausts(t *testing.T) {
	stubBackoff(t)

	calls := 0
	conflict := &pq.Error{Code: "40001"}
	err := retryReadReplica(context.Background(), func() error {
		calls++
		return conflict
	})
	require.ErrorIs(t, err, conflict)
	assert.Equal(t, readReplicaRetryMax+1, calls)
}

// A canceled context must abort the backoff immediately instead of sleeping out the
// remaining schedule -- this is what keeps a long backoff safe for callers on a tick.
func TestRetryReadReplicaHonoursContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	calls := 0
	conflict := &pq.Error{Code: "40001"}
	start := time.Now()
	err := retryReadReplica(ctx, func() error {
		calls++
		cancel() // conflict persists, but the caller goes away after the first attempt
		return conflict
	})
	require.ErrorIs(t, err, conflict)
	assert.Equal(t, 1, calls)
	assert.Less(t, time.Since(start), readReplicaRetryBase, "should not have slept the backoff")
}

// The schedule must grow and stay capped; jitter keeps it within +/-20% of the nominal value.
func TestReadReplicaBackoffSchedule(t *testing.T) {
	for attempt, nominal := range []time.Duration{
		250 * time.Millisecond,
		500 * time.Millisecond,
		time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		8 * time.Second, // capped
	} {
		got := readReplicaBackoff(attempt)
		lo := time.Duration(float64(nominal) * (1 - readReplicaRetryJitter))
		hi := time.Duration(float64(nominal) * (1 + readReplicaRetryJitter))
		assert.GreaterOrEqual(t, got, lo, "attempt %d below jitter range", attempt)
		assert.LessOrEqual(t, got, hi, "attempt %d above jitter range", attempt)
	}
}
