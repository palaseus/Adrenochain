package cache

import (
	"container/list"
	"sync"
	"time"
)

// LRUNode represents a node in the LRU cache
type LRUNode struct {
	Key       string
	Value     *CacheItem
	ExpiresAt time.Time
}

// LRUCache implements a Least Recently Used cache with TTL support
type LRUCache struct {
	capacity int
	ttl      time.Duration
	cache    map[string]*list.Element
	list     *list.List
	mu       sync.RWMutex
}

// NewLRUCache creates a new LRU cache with the specified capacity and TTL
func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		ttl:      ttl,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

// Get retrieves an item from the cache
func (lru *LRUCache) Get(key string) (*CacheItem, bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if element, found := lru.cache[key]; found {
		node := element.Value.(*LRUNode)
		
		// Check if item has expired
		if !node.ExpiresAt.IsZero() && time.Now().After(node.ExpiresAt) {
			lru.removeElement(element)
			return nil, false
		}
		
		// Move to front (most recently used)
		lru.list.MoveToFront(element)
		node.Value.Accessed = time.Now()
		node.Value.Hits++
		
		return node.Value, true
	}
	
	return nil, false
}

// Set stores an item in the cache
func (lru *LRUCache) Set(key string, value *CacheItem) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	// Check if key already exists
	if element, found := lru.cache[key]; found {
		// Update existing item
		node := element.Value.(*LRUNode)
		node.Value = value
		node.ExpiresAt = time.Now().Add(lru.ttl)
		lru.list.MoveToFront(element)
		return
	}

	// Create new node
	node := &LRUNode{
		Key:       key,
		Value:     value,
		ExpiresAt: time.Now().Add(lru.ttl),
	}

	// Add to front of list
	element := lru.list.PushFront(node)
	lru.cache[key] = element

	// Check capacity
	if lru.list.Len() > lru.capacity {
		lru.evictOldest()
	}
}

// SetWithTTL stores an item with custom TTL
func (lru *LRUCache) SetWithTTL(key string, value *CacheItem, ttl time.Duration) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	// Check if key already exists
	if element, found := lru.cache[key]; found {
		// Update existing item
		node := element.Value.(*LRUNode)
		node.Value = value
		node.ExpiresAt = time.Now().Add(ttl)
		lru.list.MoveToFront(element)
		return
	}

	// Create new node
	node := &LRUNode{
		Key:       key,
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}

	// Add to front of list
	element := lru.list.PushFront(node)
	lru.cache[key] = element

	// Check capacity
	if lru.list.Len() > lru.capacity {
		lru.evictOldest()
	}
}

// Delete removes an item from the cache
func (lru *LRUCache) Delete(key string) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if element, found := lru.cache[key]; found {
		lru.removeElement(element)
	}
}

// Clear removes all items from the cache
func (lru *LRUCache) Clear() {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	lru.cache = make(map[string]*list.Element)
	lru.list.Init()
}

// Cleanup removes expired items from the cache
func (lru *LRUCache) Cleanup() {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	now := time.Now()
	
	// Iterate from back to front (oldest first)
	for element := lru.list.Back(); element != nil; {
		node := element.Value.(*LRUNode)
		next := element.Prev() // Get next before removing
		
		if !node.ExpiresAt.IsZero() && now.After(node.ExpiresAt) {
			lru.removeElement(element)
		}
		
		element = next
	}
}

// Size returns the current number of items in the cache
func (lru *LRUCache) Size() int {
	lru.mu.RLock()
	defer lru.mu.RUnlock()
	
	return lru.list.Len()
}

// Capacity returns the maximum capacity of the cache
func (lru *LRUCache) Capacity() int {
	return lru.capacity
}

// evictOldest removes the oldest (least recently used) item
func (lru *LRUCache) evictOldest() {
	if lru.list.Len() == 0 {
		return
	}
	
	// Remove from back of list (oldest)
	element := lru.list.Back()
	lru.removeElement(element)
}

// removeElement removes a specific element from the cache
func (lru *LRUCache) removeElement(element *list.Element) {
	node := element.Value.(*LRUNode)
	delete(lru.cache, node.Key)
	lru.list.Remove(element)
}

// GetKeys returns all keys in the cache (for debugging/testing)
func (lru *LRUCache) GetKeys() []string {
	lru.mu.RLock()
	defer lru.mu.RUnlock()
	
	keys := make([]string, 0, len(lru.cache))
	for key := range lru.cache {
		keys = append(keys, key)
	}
	
	return keys
}
