package cache

import (
	"context"
	"crypto/sha256"
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/block"
	"github.com/gochain/gochain/pkg/utxo"
)

// CacheLevel represents different levels of caching
type CacheLevel int

const (
	LevelL1 CacheLevel = iota // Fastest, in-memory
	LevelL2                    // Medium, compressed
	LevelL3                    // Slowest, persistent
)

// CacheConfig holds configuration for the advanced cache system
type CacheConfig struct {
	L1Size        int           // L1 cache size (number of items)
	L2Size        int           // L2 cache size (number of items)
	L3Size        int           // L3 cache size (number of items)
	L1TTL         time.Duration // L1 cache TTL
	L2TTL         time.Duration // L2 cache TTL
	L3TTL         time.Duration // L3 cache TTL
	Compression   bool          // Enable compression for L2
	Parallelism   int           // Number of parallel workers
	EvictionPolicy string       // LRU, LFU, or FIFO
}

// DefaultCacheConfig returns sensible defaults for the cache system
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		L1Size:        10000,           // 10K items in L1
		L2Size:        100000,          // 100K items in L2
		L3Size:        1000000,         // 1M items in L3
		L1TTL:         5 * time.Minute, // 5 minutes
		L2TTL:         30 * time.Minute, // 30 minutes
		L3TTL:         24 * time.Hour,  // 24 hours
		Compression:   true,            // Enable compression
		Parallelism:   4,               // 4 parallel workers
		EvictionPolicy: "LRU",          // LRU eviction
	}
}

// CacheItem represents a cached item with metadata
type CacheItem struct {
	Key       string
	Value     interface{}
	Level     CacheLevel
	Created   time.Time
	Accessed  time.Time
	Hits      int64
	Size      int64
	Compressed bool
}

// AdvancedCache is a multi-level, high-performance caching system
type AdvancedCache struct {
	config     *CacheConfig
	l1Cache    *LRUCache
	l2Cache    *LRUCache
	l3Cache    *LRUCache
	stats      *CacheStats
	workers    chan struct{}
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.RWMutex
}

// CacheStats tracks cache performance metrics
type CacheStats struct {
	Hits        int64
	Misses      int64
	Evictions   int64
	Compressions int64
	Decompressions int64
	L1Hits      int64
	L2Hits      int64
	L3Hits      int64
	mu          sync.RWMutex
}

// NewAdvancedCache creates a new advanced cache instance
func NewAdvancedCache(config *CacheConfig) *AdvancedCache {
	if config == nil {
		config = DefaultCacheConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	cache := &AdvancedCache{
		config:  config,
		stats:   &CacheStats{},
		workers: make(chan struct{}, config.Parallelism),
		ctx:     ctx,
		cancel:  cancel,
	}

	// Initialize cache levels
	cache.l1Cache = NewLRUCache(config.L1Size, config.L1TTL)
	cache.l2Cache = NewLRUCache(config.L2Size, config.L2TTL)
	cache.l3Cache = NewLRUCache(config.L3Size, config.L3TTL)

	// Start background workers
	cache.startWorkers()

	return cache
}

// Get retrieves an item from the cache
func (ac *AdvancedCache) Get(key string) (interface{}, bool) {
	ac.stats.mu.Lock()
	defer ac.stats.mu.Unlock()

	// Try L1 first (fastest)
	if item, found := ac.l1Cache.Get(key); found {
		ac.stats.Hits++
		ac.stats.L1Hits++
		return item.Value, true
	}

	// Try L2
	if item, found := ac.l2Cache.Get(key); found {
		ac.stats.Hits++
		ac.stats.L2Hits++
		// Promote to L1
		ac.promoteToL1(key, item)
		return item.Value, true
	}

	// Try L3
	if item, found := ac.l3Cache.Get(key); found {
		ac.stats.Hits++
		ac.stats.L3Hits++
		// Promote to L2
		ac.promoteToL2(key, item)
		return item.Value, true
	}

	ac.stats.Misses++
	return nil, false
}

// Set stores an item in the cache
func (ac *AdvancedCache) Set(key string, value interface{}, level CacheLevel) {
	item := &CacheItem{
		Key:      key,
		Value:    value,
		Level:    level,
		Created:  time.Now(),
		Accessed: time.Now(),
		Hits:     0,
		Size:     ac.calculateSize(value),
	}

	switch level {
	case LevelL1:
		ac.l1Cache.Set(key, item)
	case LevelL2:
		if ac.config.Compression {
			ac.compressItem(item)
		}
		ac.l2Cache.Set(key, item)
	case LevelL3:
		ac.l3Cache.Set(key, item)
	}
}

// SetWithTTL stores an item with custom TTL
func (ac *AdvancedCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	item := &CacheItem{
		Key:      key,
		Value:    value,
		Level:    LevelL1, // Default to L1
		Created:  time.Now(),
		Accessed: time.Now(),
		Hits:     0,
		Size:     ac.calculateSize(value),
	}

	ac.l1Cache.SetWithTTL(key, item, ttl)
}

// Delete removes an item from all cache levels
func (ac *AdvancedCache) Delete(key string) {
	ac.l1Cache.Delete(key)
	ac.l2Cache.Delete(key)
	ac.l3Cache.Delete(key)
}

// Clear clears all cache levels
func (ac *AdvancedCache) Clear() {
	ac.l1Cache.Clear()
	ac.l2Cache.Clear()
	ac.l3Cache.Clear()
}

// GetStats returns current cache statistics
func (ac *AdvancedCache) GetStats() *CacheStats {
	ac.stats.mu.RLock()
	defer ac.stats.mu.RUnlock()

	stats := *ac.stats
	return &stats
}

// Close shuts down the cache and cleans up resources
func (ac *AdvancedCache) Close() {
	ac.cancel()
	ac.l1Cache.Clear()
	ac.l2Cache.Clear()
	ac.l3Cache.Clear()
}

// promoteToL1 promotes an item from L2 to L1
func (ac *AdvancedCache) promoteToL1(key string, item *CacheItem) {
	// Decompress if needed
	if item.Compressed {
		ac.decompressItem(item)
	}
	
	// Create a copy for L1
	l1Item := &CacheItem{
		Key:      item.Key,
		Value:    item.Value,
		Level:    LevelL1,
		Created:  time.Now(),
		Accessed: time.Now(),
		Hits:     item.Hits,
		Size:     item.Size,
	}
	
	ac.l1Cache.Set(key, l1Item)
}

// promoteToL2 promotes an item from L3 to L2
func (ac *AdvancedCache) promoteToL2(key string, item *CacheItem) {
	// Create a copy for L2
	l2Item := &CacheItem{
		Key:      item.Key,
		Value:    item.Value,
		Level:    LevelL2,
		Created:  time.Now(),
		Accessed: time.Now(),
		Hits:     item.Hits,
		Size:     item.Size,
	}
	
	if ac.config.Compression {
		ac.compressItem(l2Item)
	}
	
	ac.l2Cache.Set(key, l2Item)
}

// compressItem compresses an item's value
func (ac *AdvancedCache) compressItem(item *CacheItem) {
	// Simple compression simulation - in real implementation, use gzip or similar
	item.Compressed = true
	ac.stats.Compressions++
}

// decompressItem decompresses an item's value
func (ac *AdvancedCache) decompressItem(item *CacheItem) {
	// Simple decompression simulation
	item.Compressed = false
	ac.stats.Decompressions++
}

// calculateSize estimates the size of a value in bytes
func (ac *AdvancedCache) calculateSize(value interface{}) int64 {
	// Simple size estimation - in real implementation, use reflection or serialization
	switch v := value.(type) {
	case []byte:
		return int64(len(v))
	case string:
		return int64(len(v))
	case *block.Block:
		return 1024 // Estimate block size
	case *utxo.UTXO:
		return 256 // Estimate UTXO size
	default:
		return 128 // Default estimate
	}
}

// startWorkers starts background workers for cache maintenance
func (ac *AdvancedCache) startWorkers() {
	for i := 0; i < ac.config.Parallelism; i++ {
		go ac.worker()
	}
}

// worker runs background cache maintenance tasks
func (ac *AdvancedCache) worker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ac.ctx.Done():
			return
		case <-ticker.C:
			ac.performMaintenance()
		}
	}
}

// performMaintenance performs periodic cache maintenance
func (ac *AdvancedCache) performMaintenance() {
	// Clean expired items
	ac.l1Cache.Cleanup()
	ac.l2Cache.Cleanup()
	ac.l3Cache.Cleanup()

	// Balance cache levels if needed
	ac.balanceCacheLevels()
}

// balanceCacheLevels redistributes items between cache levels
func (ac *AdvancedCache) balanceCacheLevels() {
	// Move frequently accessed items to higher levels
	// Move rarely accessed items to lower levels
	// This is a simplified implementation
}

// generateCacheKey generates a cache key from multiple components
func (ac *AdvancedCache) generateCacheKey(components ...string) string {
	h := fnv.New64a()
	for _, component := range components {
		h.Write([]byte(component))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// generateHashKey generates a SHA256 hash key
func (ac *AdvancedCache) generateHashKey(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}
