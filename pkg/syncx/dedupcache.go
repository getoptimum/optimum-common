package syncx

import "time"

// EventID identifies an event stored in a DedupCache.
type EventID [16]byte

type dedupRecord struct {
	id       EventID
	deadline int64
}

// DedupCache remembers accepted event IDs for a fixed TTL.
//
// It is not safe for concurrent use: one goroutine must own the cache. All
// timestamp arguments are nonnegative, monotonic elapsed nanoseconds from the
// same clock and must not decrease.
type DedupCache struct {
	ttl     int64
	entries map[EventID]int64
	queue   []dedupRecord
	head    int
	size    int
}

// NewDedupCache creates a cache with ttl. It panics if ttl is not positive.
func NewDedupCache(ttl time.Duration) *DedupCache {
	if ttl <= 0 {
		panic("syncx: DedupCache TTL must be positive")
	}

	return &DedupCache{
		ttl:     int64(ttl),
		entries: make(map[EventID]int64),
	}
}

// Contains reports whether id is remembered and has not expired at now.
// It does not refresh the TTL or remove expired entries.
func (c *DedupCache) Contains(id EventID, now int64) bool {
	deadline, ok := c.entries[id]
	return ok && now < deadline
}

// Remember records a successful acceptance of id at now. It panics if now and
// the cache TTL cannot be added without overflowing int64.
func (c *DedupCache) Remember(id EventID, now int64) {
	if now > int64(^uint64(0)>>1)-c.ttl {
		panic("syncx: DedupCache deadline overflows int64")
	}

	deadline := now + c.ttl
	c.entries[id] = deadline
	c.push(dedupRecord{id: id, deadline: deadline})
}

// Expire removes up to budget expired queue records at now and returns the
// number of map entries removed. Stale records count toward budget and cannot
// remove an ID that was remembered again with a later deadline.
func (c *DedupCache) Expire(now int64, budget int) (removed int) {
	for budget > 0 && c.size > 0 {
		record := c.queue[c.head]
		if now < record.deadline {
			break
		}

		c.queue[c.head] = dedupRecord{}
		c.head = (c.head + 1) % len(c.queue)
		c.size--
		budget--

		if deadline, ok := c.entries[record.id]; ok && deadline == record.deadline {
			delete(c.entries, record.id)
			removed++
		}
	}

	return removed
}

// Len returns the number of remembered IDs, including expired IDs that have
// not yet been removed by Expire.
func (c *DedupCache) Len() int {
	return len(c.entries)
}

func (c *DedupCache) push(record dedupRecord) {
	if c.size == len(c.queue) {
		c.grow()
	}

	c.queue[(c.head+c.size)%len(c.queue)] = record
	c.size++
}

func (c *DedupCache) grow() {
	capacity := len(c.queue) * 2
	if capacity == 0 {
		capacity = 16
	}
	queue := make([]dedupRecord, capacity)
	for i := range c.size {
		queue[i] = c.queue[(c.head+i)%len(c.queue)]
	}
	c.queue = queue
	c.head = 0
}
