package syncx_test

import (
	"math"
	"testing"
	"time"

	"github.com/getoptimum/optimum-common/pkg/syncx"
	"github.com/stretchr/testify/require"
)

func TestDedupCacheExpiryBoundaryAndLookupDoesNotRefresh(t *testing.T) {
	cache := syncx.NewDedupCache(10 * time.Nanosecond)
	id := syncx.EventID{1}

	cache.Remember(id, 100)
	require.True(t, cache.Contains(id, 109))
	require.False(t, cache.Contains(id, 110))
	require.Equal(t, 1, cache.Len(), "Contains does not remove expired entries")
	require.Equal(t, 1, cache.Expire(110, 1))
	require.Equal(t, 0, cache.Len())
}

func TestDedupCacheExpireKeepsLaterAcceptance(t *testing.T) {
	cache := syncx.NewDedupCache(10 * time.Nanosecond)
	id := syncx.EventID{1}

	cache.Remember(id, 0)
	cache.Remember(id, 5)

	require.Equal(t, 0, cache.Expire(10, 1), "the first record is stale")
	require.True(t, cache.Contains(id, 10))
	require.Equal(t, 1, cache.Expire(15, 1))
	require.False(t, cache.Contains(id, 15))
}

func TestDedupCacheExpireHonorsBudget(t *testing.T) {
	cache := syncx.NewDedupCache(10 * time.Nanosecond)
	first := syncx.EventID{1}
	second := syncx.EventID{2}

	cache.Remember(first, 0)
	cache.Remember(second, 0)

	require.Equal(t, 0, cache.Expire(10, 0))
	require.Equal(t, 0, cache.Expire(10, -1))
	require.Equal(t, 2, cache.Len())
	require.Equal(t, 1, cache.Expire(10, 1))
	require.Equal(t, 1, cache.Len())
	require.Equal(t, 1, cache.Expire(10, 1))
	require.Equal(t, 0, cache.Len())
}

func TestDedupCacheStaleRecordsUseBudget(t *testing.T) {
	cache := syncx.NewDedupCache(10 * time.Nanosecond)
	first := syncx.EventID{1}
	second := syncx.EventID{2}

	cache.Remember(first, 0)
	cache.Remember(first, 1)
	cache.Remember(second, 1)

	require.Equal(t, 0, cache.Expire(10, 1))
	require.True(t, cache.Contains(first, 10))
	require.True(t, cache.Contains(second, 10))
	require.Equal(t, 1, cache.Expire(11, 1))
	require.Equal(t, 1, cache.Len())
}

func TestDedupCacheQueueWrapAndGrowthPreserveFIFO(t *testing.T) {
	cache := syncx.NewDedupCache(10 * time.Nanosecond)

	for i := range 32 {
		cache.Remember(syncx.EventID{byte(i)}, 0)
	}
	require.Equal(t, 20, cache.Expire(10, 20))

	for i := 32; i < 65; i++ {
		cache.Remember(syncx.EventID{byte(i)}, 1)
	}

	require.Equal(t, 12, cache.Expire(10, 100))
	require.Equal(t, 33, cache.Len())
	for i := 32; i < 65; i++ {
		require.True(t, cache.Contains(syncx.EventID{byte(i)}, 10))
	}
	require.Equal(t, 33, cache.Expire(11, 100))
	require.Equal(t, 0, cache.Len())
}

func TestNewDedupCacheRejectsInvalidTTLAndDeadlineOverflow(t *testing.T) {
	require.Panics(t, func() { syncx.NewDedupCache(0) })
	require.Panics(t, func() { syncx.NewDedupCache(-time.Nanosecond) })

	cache := syncx.NewDedupCache(time.Nanosecond)
	require.Panics(t, func() { cache.Remember(syncx.EventID{}, math.MaxInt64) })
}
