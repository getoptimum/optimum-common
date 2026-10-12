package syncx

import (
	"hash/maphash"
	"math"
	"time"
)

// RotatingBloomDedupCache remembers event IDs in two rotating Bloom filters.
// Each filter holds approximately capacity IDs; combinedFalsePositiveRate is an
// approximate capacity-based probabilistic bound across both retained filters.
// For the intended 20,000,000 IDs per 10-second generation and a 1e-6 combined
// rate, the two filters use about 151 MB and 21 hash probes per operation.
//
// It is not safe for concurrent use: one goroutine must own the cache. All
// timestamp arguments are nonnegative, monotonic elapsed nanoseconds from the
// same clock and must not decrease.
type RotatingBloomDedupCache struct {
	active   []byte
	previous []byte
	bits     uint64
	hashes   uint64
	seed1    maphash.Seed
	seed2    maphash.Seed

	rotation        int64
	generationStart int64
}

// NewRotatingBloomDedupCache creates a rotating Bloom deduplication cache.
// It panics if capacity or rotation is not positive, or if
// combinedFalsePositiveRate is not strictly between zero and one.
func NewRotatingBloomDedupCache(capacity int, combinedFalsePositiveRate float64, rotation time.Duration) *RotatingBloomDedupCache {
	if capacity <= 0 {
		panic("syncx: RotatingBloomDedupCache capacity must be positive")
	}
	if combinedFalsePositiveRate <= 0 || combinedFalsePositiveRate >= 1 || math.IsNaN(combinedFalsePositiveRate) {
		panic("syncx: RotatingBloomDedupCache combined false-positive rate must be between zero and one")
	}
	if rotation <= 0 {
		panic("syncx: RotatingBloomDedupCache rotation must be positive")
	}

	perFilterLogRate := math.Log(combinedFalsePositiveRate) - math.Log1p(math.Sqrt(1-combinedFalsePositiveRate))
	bitCount := math.Ceil(-float64(capacity) * perFilterLogRate / (math.Ln2 * math.Ln2))
	if bitCount < 1 || bitCount >= math.Exp2(64) || bitCount/8 > float64(^uint(0)>>1) {
		panic("syncx: RotatingBloomDedupCache filter is too large")
	}
	bits := uint64(bitCount)
	bytes := int((bits + 7) / 8)
	hashes := uint64(math.Round(float64(bits) / float64(capacity) * math.Ln2))
	if hashes == 0 {
		hashes = 1
	}

	return &RotatingBloomDedupCache{
		active:   make([]byte, bytes),
		previous: make([]byte, bytes),
		bits:     bits,
		hashes:   hashes,
		seed1:    maphash.MakeSeed(),
		seed2:    maphash.MakeSeed(),
		rotation: int64(rotation),
	}
}

// Contains reports whether id might have been remembered in either retained
// generation. It rotates the filters when now reaches a generation boundary.
func (c *RotatingBloomDedupCache) Contains(id EventID, now int64) bool {
	c.rotate(now)
	first := maphash.Comparable(c.seed1, id)
	second := maphash.Comparable(c.seed2, id)
	return c.contains(c.active, first, second) || c.contains(c.previous, first, second)
}

// Remember records id in the active generation. It rotates the filters when
// now reaches a generation boundary.
func (c *RotatingBloomDedupCache) Remember(id EventID, now int64) {
	c.rotate(now)
	c.remember(c.active, id)
}

func (c *RotatingBloomDedupCache) rotate(now int64) {
	elapsed := now - c.generationStart
	if elapsed < c.rotation {
		return
	}

	generations := elapsed / c.rotation
	if generations >= 2 {
		clear(c.active)
		clear(c.previous)
		c.generationStart += generations * c.rotation
		return
	}

	c.active, c.previous = c.previous, c.active
	clear(c.active)
	c.generationStart += c.rotation
}

func (c *RotatingBloomDedupCache) contains(filter []byte, first, second uint64) bool {
	for i := range c.hashes {
		bit := (first + i*second) % c.bits
		if filter[bit/8]&(1<<uint(bit%8)) == 0 {
			return false
		}
	}
	return true
}

func (c *RotatingBloomDedupCache) remember(filter []byte, id EventID) {
	first := maphash.Comparable(c.seed1, id)
	second := maphash.Comparable(c.seed2, id)
	for i := range c.hashes {
		bit := (first + i*second) % c.bits
		filter[bit/8] |= 1 << uint(bit%8)
	}
}
