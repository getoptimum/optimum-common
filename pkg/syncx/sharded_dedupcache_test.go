package syncx

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestShardedDedupCacheExpiryBoundaryAndLookupDoesNotRefresh(t *testing.T) {
	cache := NewShardedDedupCache(10 * time.Nanosecond)
	id := EventID{1}

	cache.Remember(id, 100)
	require.True(t, cache.Contains(id, 109))
	require.False(t, cache.Contains(id, 110))
	require.Equal(t, 1, cache.Len(), "Contains does not remove expired entries")
	require.Equal(t, 1, cache.Expire(110, 1))
	require.Equal(t, 0, cache.Len())
}

func TestShardedDedupCacheExpireStopsAtLiveTail(t *testing.T) {
	cache := NewShardedDedupCache(10 * time.Nanosecond)
	first := EventID{1}
	second := EventID{2}

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
	id := EventID{1}

	cache.Remember(id, 0)
	cache.Remember(id, 5)

	require.Equal(t, 0, cache.Expire(10, 1), "the first record is stale")
	require.True(t, cache.Contains(id, 10))
	require.Equal(t, 1, cache.Expire(15, 1))
	require.False(t, cache.Contains(id, 15))
}

func TestShardedDedupCacheExpireBudgetIncludesStaleRecords(t *testing.T) {
	cache := NewShardedDedupCache(10 * time.Nanosecond)
	first := EventID{1}
	second := EventID{2}

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

func TestShardedDedupCacheRoutesAcrossFixedMaps(t *testing.T) {
	cache := NewShardedDedupCache(time.Hour)
	require.Len(t, cache.shards, shardedDedupCacheShards)
	require.Equal(t, 32, shardedDedupCacheShards)

	seen := make([]bool, shardedDedupCacheShards)
	for i := range 10_000 {
		id := EventID{byte(i), byte(i >> 8)}
		shard := cache.shardIndex(id)
		cache.Remember(id, int64(i))
		deadline, ok := cache.shards[shard][id]
		require.True(t, ok)
		require.Equal(t, int64(i)+int64(time.Hour), deadline)
		seen[shard] = true
	}
	for shard, found := range seen {
		require.Truef(t, found, "shard %d did not receive an ID", shard)
	}
}

func TestNewShardedDedupCacheRejectsInvalidTTLAndDeadlineOverflow(t *testing.T) {
	require.Panics(t, func() { NewShardedDedupCache(0) })
	require.Panics(t, func() { NewShardedDedupCache(-time.Nanosecond) })

	cache := NewShardedDedupCache(time.Nanosecond)
	require.Panics(t, func() { cache.Remember(EventID{}, math.MaxInt64) })
}
