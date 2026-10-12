package syncx

import "time"

const (
	shardedDedupCacheShards    = 32
	shardedDedupCacheShardMask = shardedDedupCacheShards - 1
)

type shardedDedupRecord struct {
	id       EventID
	deadline int64
	shard    int
}

// ShardedDedupCache remembers accepted event IDs for a fixed TTL.
//
// It is not safe for concurrent use: one goroutine must own the cache. All
// timestamp arguments are nonnegative, monotonic elapsed nanoseconds from the
// same clock and must not decrease. IDs are routed across exactly 32 maps by
// their low five bits, which should be uniformly distributed.
type ShardedDedupCache struct {
	ttl    int64
	shards [shardedDedupCacheShards]map[EventID]int64
	queue  []shardedDedupRecord
	head   int
	size   int
}

// NewShardedDedupCache creates a cache with ttl. It panics if ttl is not positive.
func NewShardedDedupCache(ttl time.Duration) *ShardedDedupCache {
	if ttl <= 0 {
		panic("syncx: DedupCache TTL must be positive")
	}

	c := &ShardedDedupCache{
		ttl: int64(ttl),
	}
	for i := range c.shards {
		c.shards[i] = make(map[EventID]int64, 64)
	}
	return c
}

// Contains reports whether id is remembered and has not expired at now.
// It does not refresh the TTL or remove expired entries.
func (c *ShardedDedupCache) Contains(id EventID, now int64) bool {
	deadline, ok := c.shards[c.shardIndex(id)][id]
	return ok && now < deadline
}

// Remember records a successful acceptance of id at now. It panics if now and
// the cache TTL cannot be added without overflowing int64.
func (c *ShardedDedupCache) Remember(id EventID, now int64) {
	if now > int64(^uint64(0)>>1)-c.ttl {
		panic("syncx: DedupCache deadline overflows int64")
	}

	if c.size == len(c.queue) {
		capacity := len(c.queue) * 2
		if capacity == 0 {
			capacity = 16
		}
		queue := make([]shardedDedupRecord, capacity)
		for i := range c.size {
			queue[i] = c.queue[(c.head+i)%len(c.queue)]
		}
		c.queue = queue
		c.head = 0
	}

	deadline := now + c.ttl
	shard := c.shardIndex(id)
	c.shards[shard][id] = deadline
	c.queue[(c.head+c.size)%len(c.queue)] = shardedDedupRecord{id: id, deadline: deadline, shard: shard}
	c.size++
}

// Expire removes up to budget expired records at now and returns the number of
// map entries removed. Stale records count toward budget and cannot remove an
// ID that was remembered again with a later deadline.
func (c *ShardedDedupCache) Expire(now int64, budget int) (removed int) {
	for budget > 0 && c.size > 0 {
		record := c.queue[c.head]
		if now < record.deadline {
			break
		}

		c.queue[c.head] = shardedDedupRecord{}
		c.head = (c.head + 1) % len(c.queue)
		c.size--
		budget--

		entries := c.shards[record.shard]
		if deadline, ok := entries[record.id]; ok && deadline == record.deadline {
			delete(entries, record.id)
			removed++
		}
	}
	return removed
}

// Len returns the number of remembered IDs, including expired IDs that have
// not yet been removed by Expire.
func (c *ShardedDedupCache) Len() int {
	total := 0
	for i := range c.shards {
		total += len(c.shards[i])
	}
	return total
}

func (_ *ShardedDedupCache) shardIndex(id EventID) int {
	return int(id & shardedDedupCacheShardMask)
}
