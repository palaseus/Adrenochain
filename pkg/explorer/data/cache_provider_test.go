package data

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getCacheKeys is a helper function to get all keys in the cache for debugging
func getCacheKeys(cache *InMemoryCache) []string {
	keys := make([]string, 0, len(cache.cache))
	for key := range cache.cache {
		keys = append(keys, key)
	}
	return keys
}

func TestNewInMemoryCache(t *testing.T) {
	tests := []struct {
		name     string
		maxSize  int
		expected int
	}{
		{"zero size", 0, 0},
		{"positive size", 100, 100},
		{"large size", 10000, 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewInMemoryCache(tt.maxSize)
			require.NotNil(t, cache)
			assert.Equal(t, tt.expected, cache.stats.maxSize)
			assert.Equal(t, 0, cache.stats.getSize())
			assert.Equal(t, int64(0), cache.stats.hits)
			assert.Equal(t, int64(0), cache.stats.misses)
		})
	}
}

func TestInMemoryCache_Get(t *testing.T) {
	cache := NewInMemoryCache(10)

	t.Run("get non-existent key", func(t *testing.T) {
		value, exists := cache.Get("nonexistent")
		assert.Nil(t, value)
		assert.False(t, exists)
		assert.Equal(t, int64(1), cache.stats.misses)
	})

	t.Run("get existing key", func(t *testing.T) {
		testValue := "test value"
		cache.Set("test", testValue, time.Hour)

		value, exists := cache.Get("test")
		assert.Equal(t, testValue, value)
		assert.True(t, exists)
		assert.Equal(t, int64(1), cache.stats.hits)
	})

	t.Run("get expired key", func(t *testing.T) {
		cache.Set("expired", "expired value", -time.Hour) // Already expired

		value, exists := cache.Get("expired")
		assert.Nil(t, value)
		assert.False(t, exists)
		// Should be removed and counted as miss
		assert.Equal(t, int64(2), cache.stats.misses)
	})

	t.Run("get key with short TTL", func(t *testing.T) {
		cache.Set("short", "short value", 10*time.Millisecond)

		// Get immediately
		value, exists := cache.Get("short")
		assert.Equal(t, "short value", value)
		assert.True(t, exists)

		// Wait for expiration
		time.Sleep(20 * time.Millisecond)

		value, exists = cache.Get("short")
		assert.Nil(t, value)
		assert.False(t, exists)
	})
}

func TestInMemoryCache_Set(t *testing.T) {
	cache := NewInMemoryCache(3)

	t.Run("set new key", func(t *testing.T) {
		cache.Set("key1", "value1", time.Hour)
		assert.Equal(t, 1, cache.stats.getSize())

		value, exists := cache.Get("key1")
		assert.Equal(t, "value1", value)
		assert.True(t, exists)
	})

	t.Run("set existing key", func(t *testing.T) {
		cache.Set("key1", "new value", time.Hour)
		assert.Equal(t, 1, cache.stats.getSize()) // Size shouldn't change

		value, exists := cache.Get("key1")
		assert.Equal(t, "new value", value)
		assert.True(t, exists)
	})

	t.Run("set multiple keys", func(t *testing.T) {
		cache.Set("key2", "value2", time.Hour)
		cache.Set("key3", "value3", time.Hour)
		assert.Equal(t, 3, cache.stats.getSize())
	})

	t.Run("eviction when full", func(t *testing.T) {
		// Cache is now full (3 items: key1, key2, key3)
		// Adding key4 should trigger eviction of one item
		cache.Set("key4", "value4", time.Hour)

		// One of the existing keys should be evicted (implementation dependent)
		// Since key1 was updated, it might not be the one evicted
		// Let's check that we have exactly 3 items total
		assert.Equal(t, 3, cache.stats.getSize())

		// New item should exist
		value, exists := cache.Get("key4")
		assert.Equal(t, "value4", value)
		assert.True(t, exists)

		// Verify we have exactly 3 items
		assert.Equal(t, 3, cache.stats.getSize())
	})

	t.Run("set with zero TTL", func(t *testing.T) {
		cache.Set("zero", "zero value", 0)

		value, exists := cache.Get("zero")
		assert.Nil(t, value)
		assert.False(t, exists)
	})
}

func TestInMemoryCache_Delete(t *testing.T) {
	cache := NewInMemoryCache(10)

	t.Run("delete non-existent key", func(t *testing.T) {
		cache.Delete("nonexistent")
		assert.Equal(t, 0, cache.stats.getSize())
	})

	t.Run("delete existing key", func(t *testing.T) {
		cache.Set("delete", "delete value", time.Hour)
		assert.Equal(t, 1, cache.stats.getSize())

		cache.Delete("delete")
		assert.Equal(t, 0, cache.stats.getSize())

		value, exists := cache.Get("delete")
		assert.Nil(t, value)
		assert.False(t, exists)
	})

	t.Run("delete multiple keys", func(t *testing.T) {
		cache.Set("key1", "value1", time.Hour)
		cache.Set("key2", "value2", time.Hour)
		cache.Set("key3", "value3", time.Hour)
		assert.Equal(t, 3, cache.stats.getSize())

		cache.Delete("key1")
		cache.Delete("key3")
		assert.Equal(t, 1, cache.stats.getSize())

		value, exists := cache.Get("key2")
		assert.Equal(t, "value2", value)
		assert.True(t, exists)
	})
}

func TestInMemoryCache_Clear(t *testing.T) {
	cache := NewInMemoryCache(10)

	t.Run("clear empty cache", func(t *testing.T) {
		cache.Clear()
		assert.Equal(t, 0, cache.stats.getSize())
		assert.Equal(t, int64(0), cache.stats.hits)
		assert.Equal(t, int64(0), cache.stats.misses)
	})

	t.Run("clear populated cache", func(t *testing.T) {
		cache.Set("key1", "value1", time.Hour)
		cache.Set("key2", "value2", time.Hour)
		assert.Equal(t, 2, cache.stats.getSize())

		cache.Clear()
		assert.Equal(t, 0, cache.stats.getSize())

		value, exists := cache.Get("key1")
		assert.Nil(t, value)
		assert.False(t, exists)
	})
}

func TestInMemoryCache_GetStats(t *testing.T) {
	cache := NewInMemoryCache(100)

	t.Run("initial stats", func(t *testing.T) {
		stats := cache.GetStats()
		assert.Equal(t, int64(0), stats.Hits)
		assert.Equal(t, int64(0), stats.Misses)
		assert.Equal(t, 0, stats.Size)
		assert.Equal(t, 100, stats.MaxSize)
		assert.Equal(t, 0.0, stats.HitRate)
	})

	t.Run("stats after operations", func(t *testing.T) {
		// Add some data
		cache.Set("key1", "value1", time.Hour)
		cache.Set("key2", "value2", time.Hour)

		// Get some hits
		cache.Get("key1")
		cache.Get("key2")

		// Get some misses
		cache.Get("nonexistent1")
		cache.Get("nonexistent2")

		stats := cache.GetStats()
		assert.Equal(t, int64(2), stats.Hits)
		assert.Equal(t, int64(2), stats.Misses)
		assert.Equal(t, 2, stats.Size)
		assert.Equal(t, 100, stats.MaxSize)
		assert.Equal(t, 0.5, stats.HitRate) // 2 hits / 4 total = 0.5
	})

	t.Run("stats after clear", func(t *testing.T) {
		cache.Clear()

		stats := cache.GetStats()
		assert.Equal(t, int64(0), stats.Hits)
		assert.Equal(t, int64(0), stats.Misses)
		assert.Equal(t, 0, stats.Size)
		assert.Equal(t, 100, stats.MaxSize)
		assert.Equal(t, 0.0, stats.HitRate)
	})
}

func TestInMemoryCache_Eviction(t *testing.T) {
	cache := NewInMemoryCache(3)

	t.Run("eviction order", func(t *testing.T) {
		// Fill cache
		cache.Set("key1", "value1", time.Hour)
		cache.Set("key2", "value2", time.Hour)
		cache.Set("key3", "value3", time.Hour)
		assert.Equal(t, 3, cache.stats.getSize())

		// Add new item, should evict some items
		cache.Set("key4", "value4", time.Hour)

		// After eviction, size should be reduced
		assert.True(t, cache.stats.getSize() <= 3, "Cache size should be reduced after eviction")

		// At least one of the original keys should be evicted
		_, key1Exists := cache.Get("key1")
		_, key2Exists := cache.Get("key2")
		_, key3Exists := cache.Get("key3")

		// Not all original keys should exist
		assert.False(t, key1Exists && key2Exists && key3Exists, "At least one key should be evicted")

		// New key should exist
		value, exists := cache.Get("key4")
		assert.Equal(t, "value4", value)
		assert.True(t, exists)
	})

	t.Run("eviction with expired items", func(t *testing.T) {
		cache.Clear()

		// Add items with different TTLs
		cache.Set("expired1", "expired1", -time.Hour)
		cache.Set("expired2", "expired2", -time.Hour)
		cache.Set("valid", "valid", time.Hour)

		// Try to get expired items (should be removed)
		cache.Get("expired1")
		cache.Get("expired2")

		// Size should be 1 (only valid item)
		assert.Equal(t, 1, cache.stats.getSize())

		// Valid item should exist
		value, exists := cache.Get("valid")
		assert.Equal(t, "valid", value)
		assert.True(t, exists)
	})
}

func TestInMemoryCache_Concurrency(t *testing.T) {
	cache := NewInMemoryCache(1000)
	done := make(chan bool)

	// Start multiple goroutines
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)

				cache.Set(key, value, time.Hour)
				cache.Get(key)
				cache.Delete(key)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Cache should still be functional
	stats := cache.GetStats()
	assert.True(t, stats.Size >= 0)
	assert.True(t, stats.Size <= 1000)
}

func TestInMemoryCache_EdgeCases(t *testing.T) {
	cache := NewInMemoryCache(1)

	t.Run("nil value", func(t *testing.T) {
		cache.Set("nil", nil, time.Hour)
		value, exists := cache.Get("nil")
		assert.Nil(t, value)
		assert.True(t, exists)
	})

	t.Run("empty string key", func(t *testing.T) {
		cache.Set("", "empty key value", time.Hour)
		value, exists := cache.Get("")
		assert.Equal(t, "empty key value", value)
		assert.True(t, exists)
	})

	t.Run("very long key", func(t *testing.T) {
		longKey := string(make([]byte, 10000))
		cache.Set(longKey, "long key value", time.Hour)
		value, exists := cache.Get(longKey)
		assert.Equal(t, "long key value", value)
		assert.True(t, exists)
	})

	t.Run("negative max size", func(t *testing.T) {
		negativeCache := NewInMemoryCache(-1)
		negativeCache.Set("test", "value", time.Hour)

		// Should handle gracefully
		stats := negativeCache.GetStats()
		assert.Equal(t, -1, stats.MaxSize)
	})
}

func TestCacheStats_Methods(t *testing.T) {
	stats := &cacheStats{maxSize: 100}

	t.Run("size operations", func(t *testing.T) {
		assert.Equal(t, 0, stats.getSize())

		stats.increaseSize()
		assert.Equal(t, 1, stats.getSize())

		stats.increaseSize()
		assert.Equal(t, 2, stats.getSize())

		stats.decreaseSize()
		assert.Equal(t, 1, stats.getSize())

		stats.decreaseSize()
		assert.Equal(t, 0, stats.getSize())

		// Should not go below 0
		stats.decreaseSize()
		assert.Equal(t, 0, stats.getSize())
	})

	t.Run("hit/miss recording", func(t *testing.T) {
		stats.resetSize()

		stats.recordHit()
		assert.Equal(t, int64(1), stats.hits)

		stats.recordHit()
		assert.Equal(t, int64(2), stats.hits)

		stats.recordMiss()
		assert.Equal(t, int64(1), stats.misses)

		stats.recordMiss()
		assert.Equal(t, int64(2), stats.misses)
	})

	t.Run("reset size", func(t *testing.T) {
		stats.increaseSize()
		stats.increaseSize()
		assert.Equal(t, 2, stats.getSize())

		stats.resetSize()
		assert.Equal(t, 0, stats.getSize())
	})
}

func TestInMemoryCache_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	cache := NewInMemoryCache(10000)

	// Benchmark set operations
	start := time.Now()
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("perf_key_%d", i)
		value := fmt.Sprintf("perf_value_%d", i)
		cache.Set(key, value, time.Hour)
	}
	setDuration := time.Since(start)

	// Benchmark get operations
	start = time.Now()
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("perf_key_%d", i)
		cache.Get(key)
	}
	getDuration := time.Since(start)

	// Performance assertions - use more realistic thresholds
	assert.True(t, setDuration < 200*time.Millisecond, "Set operations took too long: %v", setDuration)
	assert.True(t, getDuration < 100*time.Millisecond, "Get operations took too long: %v", getDuration)

	t.Logf("Performance: Set 10k items in %v, Get 10k items in %v", setDuration, getDuration)
}
