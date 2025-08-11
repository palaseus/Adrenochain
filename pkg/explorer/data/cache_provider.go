package data

import (
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/explorer/service"
)

// InMemoryCache implements CacheProvider interface with an in-memory store
type InMemoryCache struct {
	mu    sync.RWMutex
	cache map[string]*cacheEntry
	stats *cacheStats
}

// cacheEntry represents a cached item with expiration
type cacheEntry struct {
	value      interface{}
	expiration time.Time
}

// cacheStats tracks cache performance metrics
type cacheStats struct {
	mu      sync.RWMutex
	hits    int64
	misses  int64
	size    int
	maxSize int
}

// NewInMemoryCache creates a new in-memory cache
func NewInMemoryCache(maxSize int) *InMemoryCache {
	return &InMemoryCache{
		cache: make(map[string]*cacheEntry),
		stats: &cacheStats{
			maxSize: maxSize,
		},
	}
}

// Get retrieves a value from the cache
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		c.stats.recordMiss()
		return nil, false
	}

	// Check if entry has expired
	if time.Now().After(entry.expiration) {
		// Remove expired entry
		c.mu.RUnlock()
		c.mu.Lock()
		delete(c.cache, key)
		c.stats.decreaseSize()
		c.mu.Unlock()
		c.mu.RLock()
		c.stats.recordMiss()
		return nil, false
	}

	c.stats.recordHit()
	return entry.value, true
}

// Set stores a value in the cache with TTL
func (c *InMemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict items to make room
	if c.stats.getSize() >= c.stats.maxSize {
		c.evictOldest()
	}

	// Create new entry
	entry := &cacheEntry{
		value:      value,
		expiration: time.Now().Add(ttl),
	}

	// Check if this is a new key (to track size)
	_, exists := c.cache[key]
	if !exists {
		c.stats.increaseSize()
	}

	c.cache[key] = entry
}

// Delete removes a key from the cache
func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.cache[key]; exists {
		delete(c.cache, key)
		c.stats.decreaseSize()
	}
}

// Clear removes all entries from the cache
func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*cacheEntry)
	c.stats.resetSize()
}

// GetStats returns cache performance statistics
func (c *InMemoryCache) GetStats() service.CacheStats {
	c.stats.mu.RLock()
	defer c.stats.mu.RUnlock()

	total := c.stats.hits + c.stats.misses
	var hitRate float64
	if total > 0 {
		hitRate = float64(c.stats.hits) / float64(total)
	}

	return service.CacheStats{
		Hits:    c.stats.hits,
		Misses:  c.stats.misses,
		HitRate: hitRate,
		Size:    c.stats.size,
		MaxSize: c.stats.maxSize,
	}
}

// evictOldest removes the oldest entries to make room for new ones
func (c *InMemoryCache) evictOldest() {
	// Simple eviction: remove 20% of entries
	targetSize := c.stats.maxSize * 8 / 10 // 80% of max size

	// Find entries to evict
	var entries []struct {
		key        string
		expiration time.Time
	}

	for key, entry := range c.cache {
		entries = append(entries, struct {
			key        string
			expiration time.Time
		}{key, entry.expiration})
	}

	// Sort by expiration (oldest first)
	// For simplicity, we'll just remove the first entries
	// In a production system, you'd want proper sorting
	toRemove := c.stats.size - targetSize
	if toRemove > 0 {
		for i := 0; i < toRemove && i < len(entries); i++ {
			delete(c.cache, entries[i].key)
			c.stats.decreaseSize()
		}
	}
}

// cacheStats methods

func (s *cacheStats) recordHit() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hits++
}

func (s *cacheStats) recordMiss() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.misses++
}

func (s *cacheStats) getSize() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.size
}

func (s *cacheStats) increaseSize() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.size++
}

func (s *cacheStats) decreaseSize() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.size > 0 {
		s.size--
	}
}

func (s *cacheStats) resetSize() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.size = 0
}
