package syncx_test

import (
	"testing"
	"time"

	"github.com/getoptimum/optimum-common/pkg/syncx"
	"github.com/stretchr/testify/require"
)

// The cleanup interval is deliberately far longer than the TTL so the sweeper
// never runs during these tests: they exercise the window where an entry has
// expired but is still present in the shard, which is what every other accessor
// treats as absent.
func TestUpsert_ExpiredEntryIsTreatedAsAbsent(t *testing.T) {
	const (
		maxTTL  = 50 * time.Millisecond
		cleanup = time.Hour
	)

	t.Run("counter restarts from the zero value after expiry", func(t *testing.T) {
		// Given a counter that has been incremented once.
		m := syncx.NewTTLMap[string, int](maxTTL, cleanup)
		defer m.Close()

		m.Upsert("k", func(v int) int { return v + 1 }, 1)
		v, ok := m.Get("k")
		require.True(t, ok)
		require.Equal(t, 1, v)

		// When the entry expires and is upserted again before the sweeper runs.
		// Len is used for the precondition rather than Get, because Get evicts an
		// expired entry and would remove the very state under test.
		time.Sleep(2 * maxTTL)
		require.Equal(t, 1, m.Len(), "entry must still be present but expired")

		m.Upsert("k", func(v int) int { return v + 1 }, 1)

		// Then the stale count is not carried over.
		v, ok = m.Get("k")
		require.True(t, ok)
		require.Equal(t, 1, v)
	})

	t.Run("fn is not applied to an expired value", func(t *testing.T) {
		// Given an expired entry that was never swept.
		m := syncx.NewTTLMap[string, string](maxTTL, cleanup)
		defer m.Close()

		m.Put("k", "stale")
		time.Sleep(2 * maxTTL)

		// When it is upserted.
		called := false
		m.Upsert("k", func(v string) string {
			called = true
			return v + "-updated"
		}, "fresh")

		// Then the zero value is inserted and fn never sees the dead value.
		require.False(t, called, "fn must not be called for an expired entry")
		v, ok := m.Get("k")
		require.True(t, ok)
		require.Equal(t, "fresh", v)
	})
}
