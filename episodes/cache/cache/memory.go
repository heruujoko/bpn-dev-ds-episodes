// Package cache is a tiny in-memory cache: a map, a mutex, and a TTL per item.
package cache

import (
	"sync"
	"time"
)

// item is one cached value plus the moment it stops being valid.
type item struct {
	value     interface{}
	expiresAt time.Time
}

// Cache is safe for concurrent use.
type Cache struct {
	mu     sync.Mutex // plain Mutex: Get mutates stats, so even reads write
	items  map[string]item
	hits   int64
	misses int64
}

func NewCache() *Cache {
	return &Cache{items: make(map[string]item)}
}

// Get returns the value and true on a hit. Missing or expired keys count as a miss.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	it, ok := c.items[key]
	if !ok {
		c.misses++
		return nil, false
	}
	if time.Now().After(it.expiresAt) {
		delete(c.items, key) // lazy cleanup of the expired entry
		c.misses++
		return nil, false
	}
	c.hits++
	return it.value, true
}

// Set stores value under key; it expires after ttl.
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = item{value: value, expiresAt: time.Now().Add(ttl)}
}

// Delete is explicit invalidation.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear flushes everything (stats are kept).
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]item)
}

// Len returns the number of stored items (expired-but-not-yet-read ones included).
func (c *Cache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}

// Stats returns hit/miss counts and the hit ratio (0..1; 0 when nothing was looked up).
func (c *Cache) Stats() (hits, misses int64, ratio float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if total := c.hits + c.misses; total > 0 {
		ratio = float64(c.hits) / float64(total)
	}
	return c.hits, c.misses, ratio
}

// MemoryUsage is a rough ESTIMATE in bytes: key + value + fixed per-entry overhead.
// ponytail: not real heap usage (Go gives no cheap way); good enough to compare strategies.
func (c *Cache) MemoryUsage() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	var total int64
	for k, it := range c.items {
		total += EntryOverhead + EstimateSize(k) + EstimateSize(it.value)
	}
	return total
}

// EntryOverhead is the assumed fixed cost per entry: map slot + expiry time + interface header.
// This is an assumption, not a measurement; real Redis overhead per key is similar in spirit.
const EntryOverhead = 64

// EstimateSize approximates the bytes a value occupies. Handles the types the demo uses;
// anything else counts as one 8-byte word.
func EstimateSize(v interface{}) int64 {
	switch x := v.(type) {
	case string:
		return 16 + int64(len(x)) // string header + data
	case []string:
		n := int64(24) // slice header
		for _, s := range x {
			n += EstimateSize(s)
		}
		return n
	case map[string]string:
		n := int64(48) // map header
		for k, s := range x {
			n += EstimateSize(k) + EstimateSize(s)
		}
		return n
	default:
		return 8
	}
}
