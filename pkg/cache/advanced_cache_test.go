package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAdvancedCache(t *testing.T) {
	config := DefaultCacheConfig()
	cache := NewAdvancedCache(config)
	defer cache.Close()

	assert.NotNil(t, cache)
	assert.Equal(t, config.L1Size, cache.l1Cache.Capacity())
	assert.Equal(t, config.L2Size, cache.l2Cache.Capacity())
	assert.Equal(t, config.L3Size, cache.l3Cache.Capacity())
}

func TestAdvancedCache_GetSet(t *testing.T) {
	cache := NewAdvancedCache(nil)
	defer cache.Close()

	// Test basic get/set
	key := "test-key"
	value := "test-value"

	cache.Set(key, value, LevelL1)

	retrieved, found := cache.Get(key)
	assert.True(t, found)
	assert.Equal(t, value, retrieved)
}

func TestAdvancedCache_MultiLevel(t *testing.T) {
	cache := NewAdvancedCache(nil)
	defer cache.Close()

	key := "multi-level-key"
	value := "multi-level-value"

	// Set in L3
	cache.Set(key, value, LevelL3)

	// Should find in L3 and promote to L2
	retrieved, found := cache.Get(key)
	assert.True(t, found)
	assert.Equal(t, value, retrieved)

	// Should now be in L2
	_, found = cache.l2Cache.Get(key)
	assert.True(t, found)
}

func TestAdvancedCache_Promotion(t *testing.T) {
	cache := NewAdvancedCache(nil)
	defer cache.Close()

	key := "promotion-key"
	value := "promotion-value"

	// Set in L3
	cache.Set(key, value, LevelL3)

	// First access - should promote to L2
	cache.Get(key)

	// Second access - should promote to L1
	cache.Get(key)

	// Should now be in L1
	_, found := cache.l1Cache.Get(key)
	assert.True(t, found)
}

func TestAdvancedCache_Compression(t *testing.T) {
	config := DefaultCacheConfig()
	config.Compression = true
	cache := NewAdvancedCache(config)
	defer cache.Close()

	key := "compression-key"
	value := "compression-value"

	// Set in L2 (should be compressed)
	cache.Set(key, value, LevelL2)

	// Get from L2 (should decompress)
	retrieved, found := cache.Get(key)
	assert.True(t, found)
	assert.Equal(t, value, retrieved)

	// Check compression stats
	stats := cache.GetStats()
	assert.Greater(t, stats.Compressions, int64(0))
	assert.Greater(t, stats.Decompressions, int64(0))
}

func TestAdvancedCache_TTL(t *testing.T) {
	cache := NewAdvancedCache(nil)
	defer cache.Close()

	key := "ttl-key"
	value := "ttl-value"
	ttl := 10 * time.Millisecond

	// Set with custom TTL
	cache.SetWithTTL(key, value, ttl)

	// Should be found immediately
	retrieved, found := cache.Get(key)
	assert.True(t, found)
	assert.Equal(t, value, retrieved)

	// Wait for expiration
	time.Sleep(ttl + 5*time.Millisecond)

	// Should not be found
	_, found = cache.Get(key)
	assert.False(t, found)
}

func TestAdvancedCache_Delete(t *testing.T) {
	cache := NewAdvancedCache(nil)
	defer cache.Close()

	key := "delete-key"
	value := "delete-value"

	// Set in multiple levels
	cache.Set(key, value, LevelL1)
	cache.Set(key, value, LevelL2)
	cache.Set(key, value, LevelL3)

	// Delete from all levels
	cache.Delete(key)

	// Should not be found in any level
	_, found := cache.Get(key)
	assert.False(t, found)
}

func TestAdvancedCache_Clear(t *testing.T) {
	cache := NewAdvancedCache(nil)
	defer cache.Close()

	// Add some items
	cache.Set("key1", "value1", LevelL1)
	cache.Set("key2", "value2", LevelL2)
	cache.Set("key3", "value3", LevelL3)

	// Clear all
	cache.Clear()

	// Should not find any items
	_, found := cache.Get("key1")
	assert.False(t, found)
	_, found = cache.Get("key2")
	assert.False(t, found)
	_, found = cache.Get("key3")
	assert.False(t, found)
}

func TestAdvancedCache_Stats(t *testing.T) {
	cache := NewAdvancedCache(nil)
	defer cache.Close()

	// Add some items and access them
	cache.Set("key1", "value1", LevelL1)
	cache.Set("key2", "value2", LevelL2)
	cache.Set("key3", "value3", LevelL3)

	// Access items to generate hits
	cache.Get("key1") // L1 hit
	cache.Get("key2") // L2 hit + promotion
	cache.Get("key3") // L3 hit + promotion

	// Get stats
	stats := cache.GetStats()

	assert.Greater(t, stats.Hits, int64(0))
	assert.Greater(t, stats.L1Hits, int64(0))
	assert.Greater(t, stats.L2Hits, int64(0))
	assert.Greater(t, stats.L3Hits, int64(0))
}

func TestAdvancedCache_KeyGeneration(t *testing.T) {
	cache := NewAdvancedCache(nil)
	defer cache.Close()

	// Test cache key generation
	components := []string{"block", "123", "hash"}
	key := cache.generateCacheKey(components...)
	assert.NotEmpty(t, key)

	// Test hash key generation
	data := []byte("test data")
	hashKey := cache.generateHashKey(data)
	assert.NotEmpty(t, hashKey)
	assert.Len(t, hashKey, 64) // SHA256 hex string
}

func TestAdvancedCache_ParallelAccess(t *testing.T) {
	cache := NewAdvancedCache(nil)
	defer cache.Close()

	// Test concurrent access
	const numGoroutines = 10
	const numOperations = 100

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("parallel-%d-%d", id, j)
				value := fmt.Sprintf("value-%d-%d", id, j)

				cache.Set(key, value, LevelL1)
				cache.Get(key)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify cache is still functional
	stats := cache.GetStats()
	assert.Greater(t, stats.Hits, int64(0))
}

func TestAdvancedCache_Eviction(t *testing.T) {
	// Create cache with small capacity to trigger eviction
	config := &CacheConfig{
		L1Size: 2,
		L2Size: 2,
		L3Size: 2,
		L1TTL:  time.Minute,
		L2TTL:  time.Minute,
		L3TTL:  time.Hour,
	}

	cache := NewAdvancedCache(config)
	defer cache.Close()

	// Fill L1 cache
	cache.Set("key1", "value1", LevelL1)
	cache.Set("key2", "value2", LevelL1)
	cache.Set("key3", "value3", LevelL1) // Should evict key1

	// key1 should be evicted
	_, found := cache.Get("key1")
	assert.False(t, found)

	// key2 and key3 should still be there
	_, found = cache.Get("key2")
	assert.True(t, found)
	_, found = cache.Get("key3")
	assert.True(t, found)
}

func TestAdvancedCache_SizeCalculation(t *testing.T) {
	cache := NewAdvancedCache(nil)
	defer cache.Close()

	// Test different value types
	byteValue := []byte("test bytes")
	stringValue := "test string"

	// Set values
	cache.Set("bytes", byteValue, LevelL1)
	cache.Set("string", stringValue, LevelL1)

	// Get stats to verify size calculation
	stats := cache.GetStats()
	assert.NotNil(t, stats)
}

func BenchmarkAdvancedCache_Get(b *testing.B) {
	cache := NewAdvancedCache(nil)
	defer cache.Close()

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("bench-key-%d", i)
		value := fmt.Sprintf("bench-value-%d", i)
		cache.Set(key, value, LevelL1)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench-key-%d", i%1000)
			cache.Get(key)
			i++
		}
	})
}

func BenchmarkAdvancedCache_Set(b *testing.B) {
	cache := NewAdvancedCache(nil)
	defer cache.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench-set-key-%d", i)
			value := fmt.Sprintf("bench-set-value-%d", i)
			cache.Set(key, value, LevelL1)
			i++
		}
	})
}
