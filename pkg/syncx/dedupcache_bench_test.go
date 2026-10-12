package syncx_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/getoptimum/optimum-common/pkg/syncx"
)

// These benchmarks compare single-owner ID deduplication with identical uint64
// keys. Consecutive counters keep the low bits balanced across shards.
// Both DedupCache timings include reading the caller's monotonic clock.
// Run without -race, for example:
// go test ./pkg/syncx -run '^$' -bench '^BenchmarkIDCache' -benchmem -count=3
func BenchmarkIDCacheLookup(b *testing.B) {
	for _, n := range []int{10_000, 1_000_000} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			ids := make([]syncx.EventID, 3*n)
			for i := range ids {
				ids[i] = uint64(i)
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
						name, keys = "Miss", ids[n:2*n]
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
						name, keys = "Miss", ids[n:2*n]
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
						name, keys = "Miss", ids[n:2*n]
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
			b.Run("RotatingBloomDedupCache", func(b *testing.B) {
				const rotation = 24 * time.Hour
				m := syncx.NewRotatingBloomDedupCache(n, 1e-6, rotation)
				for _, id := range ids[:n] {
					m.Remember(id, 0)
				}
				for _, id := range ids[n : 2*n] {
					m.Remember(id, int64(rotation))
				}
				origin := time.Now()
				for _, kind := range []struct {
					name string
					keys []syncx.EventID
					hit  bool
				}{
					{name: "ActiveHit", keys: ids[n : 2*n], hit: true},
					{name: "PreviousHit", keys: ids[:n], hit: true},
					{name: "Miss", keys: ids[2*n:]},
				} {
					b.Run(kind.name, func(b *testing.B) {
						b.ReportAllocs()
						hits := 0
						b.ResetTimer()
						for i := range b.N {
							if m.Contains(kind.keys[i%n], int64(rotation)+time.Since(origin).Nanoseconds()) {
								hits++
							}
						}
						b.StopTimer()
						if kind.hit && hits != b.N {
							b.Fatalf("unexpected hits: %d of %d", hits, b.N)
						}
						if !kind.hit {
							b.ReportMetric(float64(hits)/float64(b.N), "false-positive-hits/op")
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
				ids[i] = uint64(i)
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
			b.Run("RotatingBloomDedupCache", func(b *testing.B) {
				origin := time.Now()
				b.ReportAllocs()
				b.ResetTimer()
				for range b.N {
					m := syncx.NewRotatingBloomDedupCache(n, 1e-6, time.Hour)
					for _, id := range ids {
						m.Remember(id, time.Since(origin).Nanoseconds())
					}
					if !m.Contains(ids[0], time.Since(origin).Nanoseconds()) {
						b.Fatal("remembered ID was not found")
					}
				}
				b.StopTimer()
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(n), "ns/ID")
			})
		})
	}
}

func BenchmarkRotatingBloomDedupCacheRotation(b *testing.B) {
	const (
		capacity = 1_000_000
		rotation = time.Second
	)
	ids := make([]syncx.EventID, capacity)
	for i := range ids {
		ids[i] = uint64(i)
	}
	b.Run("Steady", func(b *testing.B) {
		m := syncx.NewRotatingBloomDedupCache(capacity, 1e-6, rotation)
		for _, id := range ids {
			m.Remember(id, 0)
		}
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			if !m.Contains(ids[0], 0) {
				b.Fatal("remembered ID was not found")
			}
		}
	})
	b.Run("FullCycle", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			m := syncx.NewRotatingBloomDedupCache(capacity, 1e-6, rotation)
			for _, id := range ids {
				m.Remember(id, 0)
			}
			m.Contains(ids[0], int64(rotation))
			for _, id := range ids {
				m.Remember(id, int64(rotation))
			}
			m.Contains(ids[0], 2*int64(rotation))
		}
	})
}
