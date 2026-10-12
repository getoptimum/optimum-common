package syncx

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestShardedDedupCacheExpiryBoundaryAndLookupDoesNotRefresh(t *testing.T) {
	cache := NewShardedDedupCache(10 * time.Nanosecond)
	id := uint64(1)

	cache.Remember(id, 100)
	require.True(t, cache.Contains(id, 109))
	require.False(t, cache.Contains(id, 110))
	require.Equal(t, 1, cache.Len(), "Contains does not remove expired entries")
	require.Equal(t, 1, cache.Expire(110, 1))
	require.Equal(t, 0, cache.Len())
}

func TestShardedDedupCacheExpireStopsAtLiveTail(t *testing.T) {
	cache := NewShardedDedupCache(10 * time.Nanosecond)
	first := uint64(1)
	second := uint64(2)

	cache.Remember(first, 0)
	cache.Remember(second, 1)
	require.Equal(t, 0, cache.Expire(9, 10))
	require.Equal(t, 2, cache.Len())
	require.Equal(t, 1, cache.Expire(10, 10))
	require.True(t, cache.Contains(second, 10))
	require.Equal(t, 1, cache.Expire(11, 10))
}

func TestShardedDedupCacheExpireKeepsLaterAcceptance(t *testing.T) {
	cache := NewShardedDedupCache(10 * time.Nanosecond)
	id := uint64(1)

	cache.Remember(id, 0)
	cache.Remember(id, 5)

	require.Equal(t, 0, cache.Expire(10, 1), "the first record is stale")
	require.True(t, cache.Contains(id, 10))
	require.Equal(t, 1, cache.Expire(15, 1))
	require.False(t, cache.Contains(id, 15))
}

func TestShardedDedupCacheExpireBudgetIncludesStaleRecords(t *testing.T) {
	cache := NewShardedDedupCache(10 * time.Nanosecond)
	first := uint64(1)
	second := uint64(2)

	cache.Remember(first, 0)
	cache.Remember(second, 0)
	cache.Remember(first, 1)

	require.Equal(t, 0, cache.Expire(10, 0))
	require.Equal(t, 0, cache.Expire(10, -1))
	require.Equal(t, 0, cache.Expire(10, 1))
	require.Equal(t, 2, cache.Len(), "the stale record consumed the only budget")
	require.True(t, cache.Contains(first, 10))
	require.Equal(t, 1, cache.Expire(10, 1))
	require.Equal(t, 1, cache.Len())
	require.True(t, cache.Contains(first, 10))
}

func TestShardedDedupCacheQueueWrapGrowthAndReuse(t *testing.T) {
	cache := NewShardedDedupCache(10 * time.Nanosecond)
	require.Equal(t, 0, cache.Expire(0, 100))

	for i := range 32 {
		cache.Remember(uint64(i), 0)
	}
	require.Equal(t, 20, cache.Expire(10, 20))

	// Wrap into freed slots, then grow while the oldest record is not at index zero.
	for i := 32; i < 65; i++ {
		cache.Remember(uint64(i), 10)
	}
	require.Equal(t, 12, cache.Expire(10, 100))
	require.Equal(t, 33, cache.Len())
	for i := 32; i < 65; i++ {
		require.True(t, cache.Contains(uint64(i), 10))
	}
	require.Equal(t, 0, cache.Expire(19, 100))
	require.Equal(t, 33, cache.Expire(20, 100))
	require.Zero(t, cache.Len())
	require.Equal(t, 0, cache.Expire(20, 100))

	for i := range 65 {
		cache.Remember(uint64(i), 20)
	}
	require.Equal(t, 0, cache.Expire(29, 100))
	require.Equal(t, 65, cache.Expire(30, 100))
	require.Zero(t, cache.Len())
}

func TestShardedDedupCacheRoutesAcrossFixedMaps(t *testing.T) {
	cache := NewShardedDedupCache(time.Hour)
	require.Len(t, cache.shards, shardedDedupCacheShards)
	require.Equal(t, 32, shardedDedupCacheShards)

	for i := range 256 {
		id := uint64(255-i)<<56 | uint64(i)
		shard := i % 32
		require.Equal(t, shard, cache.shardIndex(id))
		cache.Remember(id, int64(i))
		deadline, ok := cache.shards[shard][id]
		require.True(t, ok)
		require.Equal(t, int64(i)+int64(time.Hour), deadline)
		require.True(t, cache.Contains(id, int64(i)))
	}
	for _, entries := range cache.shards {
		require.Len(t, entries, 8)
	}
	require.Equal(t, 256, cache.Len())
	require.Equal(t, 256, cache.Expire(int64(time.Hour)+256, 256))
	require.Zero(t, cache.Len())
}

func TestNewShardedDedupCacheRejectsInvalidTTLAndDeadlineOverflow(t *testing.T) {
	require.Panics(t, func() { NewShardedDedupCache(0) })
	require.Panics(t, func() { NewShardedDedupCache(-time.Nanosecond) })

	cache := NewShardedDedupCache(time.Nanosecond)
	require.Panics(t, func() { cache.Remember(0, math.MaxInt64) })
}

func TestShardedDedupCacheSupportsUint64EdgeIDs(t *testing.T) {
	cache := NewShardedDedupCache(time.Hour)
	for _, id := range []uint64{0, 1 << 63, math.MaxUint64} {
		cache.Remember(id, 0)
		require.True(t, cache.Contains(id, 0))
	}
}
