package syncx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestTTLMap_ExpiryIsOnMonotonicClock pins every write path to the map's monotonic clock.
// A wall-clock expiry (Unix nanoseconds) moves with NTP steps and VM resumes, which
// would flush a dedup map at once or keep stale peers alive, so the stored expiry must
// be an offset from the map's epoch that never exceeds the elapsed time plus maxTTL.
func TestTTLMap_ExpiryIsOnMonotonicClock(t *testing.T) {
	const maxTTL = time.Minute

	tests := []struct {
		name  string
		write func(m *TTLMap[string, int])
	}{
		{name: "Put", write: func(m *TTLMap[string, int]) { m.Put("k", 1) }},
		{name: "PutIfNotExists", write: func(m *TTLMap[string, int]) { m.PutIfNotExists("k", 1) }},
		{name: "Upsert", write: func(m *TTLMap[string, int]) {
			m.Upsert("k", func(v int) int { return v + 1 }, 1)
		}},
		{name: "GetAndRefresh", write: func(m *TTLMap[string, int]) {
			m.Put("k", 1)
			m.GetAndRefresh("k")
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given a fresh map.
			m := NewTTLMap[string, int](maxTTL, time.Hour)
			defer m.Close()

			// When the key is written through this path.
			tc.write(m)

			// Then its expiry is measured from the map's epoch, not from the Unix epoch.
			it, ok := m.shardOf("k").m["k"]
			require.True(t, ok)
			require.LessOrEqual(t, it.expiryNano, int64(time.Since(m.epoch))+int64(maxTTL))
			require.Greater(t, it.expiryNano, int64(0))

			// And the entry is still live through the public API.
			_, ok = m.Get("k")
			require.True(t, ok)
		})
	}
}
