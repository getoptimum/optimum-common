package syncx_test

import (
	"testing"
	"time"

	"github.com/getoptimum/optimum-common/pkg/syncx"
)

// BenchmarkIDCacheChurn measures continuous single-owner traffic with cleanup.
// Run each cache for approximately 30 seconds using the standard Go benchmark:
// go test ./pkg/syncx -run '^$' -bench '^BenchmarkIDCacheChurn$' -benchmem -benchtime=30s -count=1
//
// One operation processes a new ID and its duplicate, remembering only misses.
// Exact caches use a real 20 ms TTL; Bloom rotates every 20 ms and retains two
// generations. TTLMap sweeps every millisecond; both DedupCaches expire up to
// 256 records every 64 pairs in the owner's goroutine.
// Faster implementations retain more IDs at this unthrottled rate; entries
// reports the final retained count, including unswept expired entries.
// Bloom's capacity is 1,000,000 IDs per generation, not a fixed live entry count.
// It reports maximum generation load and observed false-positive hits.
// Cache construction is excluded from timing and allocation metrics.
// This measures cache throughput, not the ADR's fixed-rate end-to-end workload.
func BenchmarkIDCacheChurn(b *testing.B) {
	const ttl = 20 * time.Millisecond
	b.Run("TTLMap", func(b *testing.B) {
		m := syncx.NewTTLMap[syncx.EventID, struct{}](ttl, time.Millisecond)
		defer m.Close()
		duplicates := 0
		b.ReportAllocs()
		b.ResetTimer()
		for i := range b.N {
			id := uint64(i)
			for range 2 {
				if _, ok := m.Get(id); ok {
					duplicates++
				} else {
					m.Put(id, struct{}{})
				}
			}
		}
		b.StopTimer()
		m.Close()
		b.ReportMetric(float64(m.Len()), "entries")
		b.ReportMetric(float64(duplicates)/float64(b.N), "duplicates/op")
	})
	b.Run("DedupCache", func(b *testing.B) {
		origin := time.Now()
		m := syncx.NewDedupCache(ttl)
		duplicates := 0
		b.ReportAllocs()
		b.ResetTimer()
		for i := range b.N {
			id := uint64(i)
			for copy := range 2 {
				now := time.Since(origin).Nanoseconds()
				if copy == 0 && i%64 == 0 {
					m.Expire(now, 256)
				}
				if m.Contains(id, now) {
					duplicates++
				} else {
					m.Remember(id, now)
				}
			}
		}
		b.StopTimer()
		b.ReportMetric(float64(m.Len()), "entries")
		b.ReportMetric(float64(duplicates)/float64(b.N), "duplicates/op")
	})
	b.Run("ShardedDedupCache", func(b *testing.B) {
		origin := time.Now()
		m := syncx.NewShardedDedupCache(ttl)
		duplicates := 0
		b.ReportAllocs()
		b.ResetTimer()
		for i := range b.N {
			id := uint64(i)
			for copy := range 2 {
				now := time.Since(origin).Nanoseconds()
				if copy == 0 && i%64 == 0 {
					m.Expire(now, 256)
				}
				if m.Contains(id, now) {
					duplicates++
				} else {
					m.Remember(id, now)
				}
			}
		}
		b.StopTimer()
		b.ReportMetric(float64(m.Len()), "entries")
		b.ReportMetric(float64(duplicates)/float64(b.N), "duplicates/op")
	})
	b.Run("RotatingBloomDedupCache", func(b *testing.B) {
		const capacity = 1_000_000
		origin := time.Now()
		m := syncx.NewRotatingBloomDedupCache(capacity, 1e-6, ttl)
		duplicates := 0
		falsePositiveHits := 0
		lastGeneration := int64(-1)
		generationInputs := 0
		maxGenerationInputs := 0
		b.ReportAllocs()
		b.ResetTimer()
		for i := range b.N {
			id := uint64(i)
			for copy := range 2 {
				now := time.Since(origin).Nanoseconds()
				if copy == 0 {
					generation := now / int64(ttl)
					if generation != lastGeneration {
						if generationInputs > maxGenerationInputs {
							maxGenerationInputs = generationInputs
						}
						generationInputs = 0
						lastGeneration = generation
					}
					generationInputs++
				}
				if m.Contains(id, now) {
					duplicates++
					if copy == 0 {
						falsePositiveHits++
					}
				} else {
					m.Remember(id, now)
				}
			}
		}
		b.StopTimer()
		if generationInputs > maxGenerationInputs {
			maxGenerationInputs = generationInputs
		}
		b.ReportMetric(float64(duplicates)/float64(b.N), "duplicates/op")
		b.ReportMetric(float64(falsePositiveHits)/float64(b.N), "false-positive-hits/op")
		b.ReportMetric(float64(maxGenerationInputs)/float64(capacity), "max-generation-load")
	})
}
