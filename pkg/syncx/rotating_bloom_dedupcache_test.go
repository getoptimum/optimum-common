package syncx

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRotatingBloomDedupCacheRetainsPreviousGenerationAtBoundary(t *testing.T) {
	const rotation = 10 * time.Nanosecond
	cache := NewRotatingBloomDedupCache(100, 1e-6, rotation)
	first := uint64(1)
	second := uint64(2)
	offset := NewRotatingBloomDedupCache(100, 1e-6, rotation)
	offset.Remember(first, 5)
	offset.Contains(first, int64(rotation))
	require.Equal(t, int64(rotation), offset.generationStart)

	cache.Remember(first, 0)
	require.True(t, cache.Contains(first, int64(rotation)))
	cache.Remember(second, int64(rotation))
	require.True(t, cache.Contains(first, int64(rotation)))
	require.True(t, cache.Contains(second, int64(rotation)))

	cache.Contains(second, 2*int64(rotation))
	for _, value := range cache.active {
		require.Zero(t, value, "the oldest generation was discarded")
	}
	require.True(t, cache.Contains(second, 2*int64(rotation)))
}

func TestRotatingBloomDedupCacheClearsBothFiltersAfterMultipleIntervals(t *testing.T) {
	const rotation = 10 * time.Nanosecond
	cache := NewRotatingBloomDedupCache(100, 1e-6, rotation)
	cache.Remember(1, 0)
	cache.Remember(2, int64(rotation))

	cache.Contains(0, 3*int64(rotation))
	for _, filter := range [][]byte{cache.active, cache.previous} {
		for _, value := range filter {
			require.Zero(t, value)
		}
	}
}

func TestNewRotatingBloomDedupCacheRejectsInvalidParameters(t *testing.T) {
	require.NotPanics(t, func() { NewRotatingBloomDedupCache(1, 1e-20, time.Second) })
	require.NotPanics(t, func() { NewRotatingBloomDedupCache(1, math.SmallestNonzeroFloat64, time.Second) })
	require.Panics(t, func() { NewRotatingBloomDedupCache(0, 1e-6, time.Second) })
	require.Panics(t, func() { NewRotatingBloomDedupCache(-1, 1e-6, time.Second) })
	require.Panics(t, func() { NewRotatingBloomDedupCache(1, 0, time.Second) })
	require.Panics(t, func() { NewRotatingBloomDedupCache(1, -1, time.Second) })
	require.Panics(t, func() { NewRotatingBloomDedupCache(1, 1, time.Second) })
	require.Panics(t, func() { NewRotatingBloomDedupCache(1, math.NaN(), time.Second) })
	require.Panics(t, func() { NewRotatingBloomDedupCache(1, 1e-6, 0) })
	require.Panics(t, func() { NewRotatingBloomDedupCache(1, 1e-6, -time.Second) })
}

func TestRotatingBloomDedupCacheCanProduceFalsePositiveWhenSaturated(t *testing.T) {
	cache := NewRotatingBloomDedupCache(1, 0.99, time.Second)
	first := uint64(1)
	second := uint64(2)

	cache.Remember(first, 0)
	require.Equal(t, uint64(1), cache.bits)
	require.True(t, cache.Contains(second, 0))
}

func TestRotatingBloomDedupCacheSupportsUint64EdgeIDs(t *testing.T) {
	cache := NewRotatingBloomDedupCache(100, 1e-6, time.Hour)
	for _, id := range []uint64{0, 1 << 63, math.MaxUint64} {
		cache.Remember(id, 0)
		require.True(t, cache.Contains(id, 0))
	}
}
