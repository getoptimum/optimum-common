package syncx_test

import (
	"encoding/binary"
	"strconv"
	"testing"
	"time"

	"github.com/getoptimum/optimum-common/pkg/syncx"
)

// These benchmarks compare single-owner ID deduplication with identical 16-byte
// keys. Both DedupCache timings include reading the caller's monotonic clock.
// Run without -race, for example:
// go test ./pkg/syncx -run '^$' -bench '^BenchmarkIDCache' -benchmem -count=3
func BenchmarkIDCacheLookup(b *testing.B) {
	for _, n := range []int{10_000, 1_000_000} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			ids := make([]syncx.EventID, 2*n)
			for i := range ids {
				binary.LittleEndian.PutUint64(ids[i][:8], uint64(i))
			}
			b.Run("TTLMap", func(b *testing.B) {
				m := syncx.NewTTLMap[syncx.EventID, struct{}](time.Hour, time.Hour)
				defer m.Close()
				for _, id := range ids[:n] {
					m.Put(id, struct{}{})
				}
				for _, hit := range []bool{true, false} {
					name, keys := "Hit", ids[:n]
					if !hit {
						name, keys = "Miss", ids[n:]
					}
					b.Run(name, func(b *testing.B) {
						b.ReportAllocs()
						hits := 0
						b.ResetTimer()
						for i := range b.N {
							if _, ok := m.Get(keys[i%n]); ok {
								hits++
							}
						}
						b.StopTimer()
						if (hit && hits != b.N) || (!hit && hits != 0) {
							b.Fatalf("unexpected hits: %d of %d", hits, b.N)
						}
					})
				}
			})
			b.Run("DedupCache", func(b *testing.B) {
				origin := time.Now()
				m := syncx.NewDedupCache(time.Hour)
				for _, id := range ids[:n] {
					m.Remember(id, time.Since(origin).Nanoseconds())
				}
				for _, hit := range []bool{true, false} {
					name, keys := "Hit", ids[:n]
					if !hit {
						name, keys = "Miss", ids[n:]
					}
					b.Run(name, func(b *testing.B) {
						b.ReportAllocs()
						hits := 0
						b.ResetTimer()
						for i := range b.N {
							if m.Contains(keys[i%n], time.Since(origin).Nanoseconds()) {
								hits++
							}
						}
						b.StopTimer()
						if (hit && hits != b.N) || (!hit && hits != 0) {
							b.Fatalf("unexpected hits: %d of %d", hits, b.N)
						}
					})
				}
			})
			b.Run("ShardedDedupCache", func(b *testing.B) {
				origin := time.Now()
				m := syncx.NewShardedDedupCache(time.Hour)
				for _, id := range ids[:n] {
					m.Remember(id, time.Since(origin).Nanoseconds())
				}
				for _, hit := range []bool{true, false} {
					name, keys := "Hit", ids[:n]
					if !hit {
						name, keys = "Miss", ids[n:]
					}
					b.Run(name, func(b *testing.B) {
						b.ReportAllocs()
						hits := 0
						b.ResetTimer()
						for i := range b.N {
							if m.Contains(keys[i%n], time.Since(origin).Nanoseconds()) {
								hits++
							}
						}
						b.StopTimer()
						if (hit && hits != b.N) || (!hit && hits != 0) {
							b.Fatalf("unexpected hits: %d of %d", hits, b.N)
						}
					})
				}
			})
		})
	}
}

// One operation fills a fresh cache with n distinct IDs, including constructor,
// growth and TTLMap shutdown costs. ns/ID normalizes time; B/op and allocs/op
// describe the whole batch, not one ID or retained heap size.
func BenchmarkIDCacheFill(b *testing.B) {
	for _, n := range []int{10_000, 1_000_000} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			ids := make([]syncx.EventID, n)
			for i := range ids {
				binary.LittleEndian.PutUint64(ids[i][:8], uint64(i))
			}
			b.Run("TTLMap", func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				for range b.N {
					m := syncx.NewTTLMap[syncx.EventID, struct{}](time.Hour, time.Hour)
					for _, id := range ids {
						m.Put(id, struct{}{})
					}
					m.Close()
					if got := m.Len(); got != n {
						b.Fatalf("entries = %d, want %d", got, n)
					}
				}
				b.StopTimer()
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(n), "ns/ID")
			})
			b.Run("DedupCache", func(b *testing.B) {
				origin := time.Now()
				b.ReportAllocs()
				b.ResetTimer()
				for range b.N {
					m := syncx.NewDedupCache(time.Hour)
					for _, id := range ids {
						m.Remember(id, time.Since(origin).Nanoseconds())
					}
					if got := m.Len(); got != n {
						b.Fatalf("entries = %d, want %d", got, n)
					}
				}
				b.StopTimer()
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(n), "ns/ID")
			})
			b.Run("ShardedDedupCache", func(b *testing.B) {
				origin := time.Now()
				b.ReportAllocs()
				b.ResetTimer()
				for range b.N {
					m := syncx.NewShardedDedupCache(time.Hour)
					for _, id := range ids {
						m.Remember(id, time.Since(origin).Nanoseconds())
					}
					if got := m.Len(); got != n {
						b.Fatalf("entries = %d, want %d", got, n)
					}
				}
				b.StopTimer()
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(n), "ns/ID")
			})
		})
	}
}

// One operation processes a new ID and its duplicate, remembering only misses.
// All caches use a real 20 ms TTL so normal benchmark runs exercise cleanup.
// TTLMap sweeps every millisecond; both DedupCaches expire up to 256 records
// every 64 pairs. Faster implementations retain more IDs at this unthrottled
// rate; entries reports the final retained count, including unswept expired entries.
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
			var id syncx.EventID
			binary.LittleEndian.PutUint64(id[:8], uint64(i))
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
			var id syncx.EventID
			binary.LittleEndian.PutUint64(id[:8], uint64(i))
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
			var id syncx.EventID
			binary.LittleEndian.PutUint64(id[:8], uint64(i))
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
}
