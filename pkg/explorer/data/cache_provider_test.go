package data

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// MockCacheProvider implements CacheProvider interface for testing
type MockCacheProvider struct {
	cache map[string]interface{}
	ttl   map[string]time.Time
	mu    sync.RWMutex
}

func NewMockCacheProvider() *MockCacheProvider {
	return &MockCacheProvider{
		cache: make(map[string]interface{}),
		ttl:   make(map[string]time.Time),
	}
}

func (mcp *MockCacheProvider) Get(key string) (interface{}, bool) {
	mcp.mu.RLock()
	defer mcp.mu.RUnlock()
	
	if value, exists := mcp.cache[key]; exists {
		if ttl, hasTTL := mcp.ttl[key]; hasTTL && time.Now().After(ttl) {
			delete(mcp.cache, key)
			delete(mcp.ttl, key)
			return nil, false
		}
		return value, true
	}
	return nil, false
}

func (mcp *MockCacheProvider) Set(key string, value interface{}, ttl time.Duration) {
	mcp.mu.Lock()
	defer mcp.mu.Unlock()
	
	mcp.cache[key] = value
	if ttl > 0 {
		mcp.ttl[key] = time.Now().Add(ttl)
	}
}

func (mcp *MockCacheProvider) Delete(key string) {
	mcp.mu.Lock()
	defer mcp.mu.Unlock()
	
	delete(mcp.cache, key)
	delete(mcp.ttl, key)
}

func (mcp *MockCacheProvider) Clear() {
	mcp.mu.Lock()
	defer mcp.mu.Unlock()
	
	mcp.cache = make(map[string]interface{})
	mcp.ttl = make(map[string]time.Time)
}

func (mcp *MockCacheProvider) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_keys": len(mcp.cache),
		"total_ttl":  len(mcp.ttl),
	}
}

func TestCacheProvider_Get(t *testing.T) {
	cache := NewMockCacheProvider()
	
	// Test getting non-existent key
	value, exists := cache.Get("non-existent")
	if exists {
		t.Error("Expected key to not exist")
	}
	if value != nil {
		t.Error("Expected value to be nil")
	}
	
	// Test getting existing key
	expectedValue := "test-value"
	cache.Set("test-key", expectedValue, 0)
	
	value, exists = cache.Get("test-key")
	if !exists {
		t.Error("Expected key to exist")
	}
	if value != expectedValue {
		t.Errorf("Expected value %v, got %v", expectedValue, value)
	}
}

func TestCacheProvider_Set(t *testing.T) {
	cache := NewMockCacheProvider()
	
	// Test setting value without TTL
	testValue := "test-value"
	cache.Set("test-key", testValue, 0)
	
	value, exists := cache.Get("test-key")
	if !exists {
		t.Error("Expected key to exist after setting")
	}
	if value != testValue {
		t.Errorf("Expected value %v, got %v", testValue, value)
	}
	
	// Test setting value with TTL
	cache.Set("ttl-key", "ttl-value", 100*time.Millisecond)
	
	value, exists = cache.Get("ttl-key")
	if !exists {
		t.Error("Expected key to exist immediately after setting")
	}
	
	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)
	
	value, exists = cache.Get("ttl-key")
	if exists {
		t.Error("Expected key to not exist after TTL expiration")
	}
}

func TestCacheProvider_Delete(t *testing.T) {
	cache := NewMockCacheProvider()
	
	// Set a value
	cache.Set("test-key", "test-value", 0)
	
	// Verify it exists
	value, exists := cache.Get("test-key")
	if !exists {
		t.Error("Expected key to exist before deletion")
	}
	
	// Delete the key
	cache.Delete("test-key")
	
	// Verify it no longer exists
	value, exists = cache.Get("test-key")
	if exists {
		t.Error("Expected key to not exist after deletion")
	}
	if value != nil {
		t.Error("Expected value to be nil after deletion")
	}
}

func TestCacheProvider_Clear(t *testing.T) {
	cache := NewMockCacheProvider()
	
	// Set multiple values
	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)
	cache.Set("key3", "value3", 0)
	
	// Verify they exist
	if len(cache.cache) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(cache.cache))
	}
	
	// Clear the cache
	cache.Clear()
	
	// Verify all keys are gone
	if len(cache.cache) != 0 {
		t.Errorf("Expected 0 keys after clear, got %d", len(cache.cache))
	}
	
	// Verify individual keys are gone
	value, exists := cache.Get("key1")
	if exists {
		t.Error("Expected key1 to not exist after clear")
	}
	if value != nil {
		t.Error("Expected value to be nil after clear")
	}
}

func TestCacheProvider_GetStats(t *testing.T) {
	cache := NewMockCacheProvider()
	
	// Get initial stats
	stats := cache.GetStats()
	if stats["total_keys"] != 0 {
		t.Errorf("Expected 0 total keys, got %v", stats["total_keys"])
	}
	
	// Set some values
	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 100*time.Millisecond)
	
	// Get updated stats
	stats = cache.GetStats()
	if stats["total_keys"] != 2 {
		t.Errorf("Expected 2 total keys, got %v", stats["total_keys"])
	}
	if stats["total_ttl"] != 1 {
		t.Errorf("Expected 1 total TTL entries, got %v", stats["total_ttl"])
	}
}

func TestCacheProvider_TTLExpiration(t *testing.T) {
	cache := NewMockCacheProvider()
	
	// Set value with very short TTL
	cache.Set("short-ttl", "value", 10*time.Millisecond)
	
	// Verify it exists immediately
	value, exists := cache.Get("short-ttl")
	if !exists {
		t.Error("Expected key to exist immediately")
	}
	
	// Wait for TTL to expire
	time.Sleep(20 * time.Millisecond)
	
	// Verify it's gone
	value, exists = cache.Get("short-ttl")
	if exists {
		t.Error("Expected key to not exist after TTL expiration")
	}
	if value != nil {
		t.Error("Expected value to be nil after TTL expiration")
	}
}

func TestCacheProvider_ConcurrentAccess(t *testing.T) {
	cache := NewMockCacheProvider()
	done := make(chan bool)
	
	// Start multiple goroutines
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				cache.Set(key, fmt.Sprintf("value-%d-%d", id, j), 0)
				cache.Get(key)
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify final state
	stats := cache.GetStats()
	expectedKeys := 1000 // 10 goroutines * 100 keys each
	if stats["total_keys"] != expectedKeys {
		t.Errorf("Expected %d total keys, got %v", expectedKeys, stats["total_keys"])
	}
}

func TestCacheProvider_EdgeCases(t *testing.T) {
	cache := NewMockCacheProvider()
	
	// Test setting nil value
	cache.Set("nil-key", nil, 0)
	value, exists := cache.Get("nil-key")
	if !exists {
		t.Error("Expected nil key to exist")
	}
	if value != nil {
		t.Error("Expected value to be nil")
	}
	
	// Test setting empty string key
	cache.Set("", "empty-key-value", 0)
	value, exists = cache.Get("")
	if !exists {
		t.Error("Expected empty key to exist")
	}
	if value != "empty-key-value" {
		t.Errorf("Expected value 'empty-key-value', got %v", value)
	}
	
	// Test setting very long key
	longKey := string(make([]byte, 1000))
	cache.Set(longKey, "long-key-value", 0)
	value, exists = cache.Get(longKey)
	if !exists {
		t.Error("Expected long key to exist")
	}
	if value != "long-key-value" {
		t.Errorf("Expected value 'long-key-value', got %v", value)
	}
}
