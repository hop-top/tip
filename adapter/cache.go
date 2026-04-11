package adapter

import (
	"sync"
	"time"
)

// DefaultCacheTTL is 24 hours for adapter describe cache.
const DefaultCacheTTL = 24 * time.Hour

type cacheEntry struct {
	resp      *DescribeResponse
	cachedAt  time.Time
}

// DescribeCache caches adapter describe responses in memory with TTL.
type DescribeCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	ttl     time.Duration
	now     func() time.Time // overridable for testing
}

// NewDescribeCache creates a cache with default TTL.
func NewDescribeCache() *DescribeCache {
	return &DescribeCache{
		entries: make(map[string]cacheEntry),
		ttl:     DefaultCacheTTL,
		now:     time.Now,
	}
}

// NewDescribeCacheWithTTL creates a cache with custom TTL.
func NewDescribeCacheWithTTL(ttl time.Duration) *DescribeCache {
	return &DescribeCache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
		now:     time.Now,
	}
}

// CachedDescribe returns a describe response for the adapter at path.
// On first call, it invokes CallDescribe and caches the result.
// Subsequent calls return from cache until TTL expires.
func (c *DescribeCache) CachedDescribe(adapterPath string) (*DescribeResponse, error) {
	c.mu.RLock()
	if e, ok := c.entries[adapterPath]; ok {
		if c.now().Sub(e.cachedAt) < c.ttl {
			c.mu.RUnlock()
			return e.resp, nil
		}
	}
	c.mu.RUnlock()

	resp, err := CallDescribe(adapterPath)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.entries[adapterPath] = cacheEntry{
		resp:     resp,
		cachedAt: c.now(),
	}
	c.mu.Unlock()

	return resp, nil
}

// Len returns the number of entries in cache (for testing).
func (c *DescribeCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// SetNow overrides the time function (for testing).
func (c *DescribeCache) SetNow(fn func() time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = fn
}
