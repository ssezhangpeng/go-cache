package cache

import (
	"sync"
	"time"
)

// Cache is a cache interface which should supports ttl.
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttlSeconds int) error
}

var _ Cache = (*TTLCache)(nil)

// TTLCache is a simple cache which supports ttl.
type TTLCache struct {
	cleanupInterval time.Duration

	mutex sync.RWMutex
	items map[string]*Item
}

// Set is a thread-safe way to add new items to the map
func (cache *TTLCache) Set(key string, value interface{}, ttlSeconds int) error {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	item := &Item{data: value}
	item.touch(time.Duration(ttlSeconds) * time.Second)
	cache.items[key] = item

	return nil
}

// Get is a thread-safe way to lookup items, and extends the item's expire time.
func (cache *TTLCache) Get(key string) (value interface{}, found bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	item, exists := cache.items[key]
	if !exists || item.expired(time.Now()) {
		return "", false
	}

	// refresh
	item.touch(item.ttl)
	return item.data, true
}

// Count returns the number of items in the cache.
func (cache *TTLCache) Count() int {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	count := len(cache.items)
	return count
}

func (cache *TTLCache) cleanup() {
	now := time.Now()
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	for key, item := range cache.items {
		if item.expired(now) {
			delete(cache.items, key)
		}
	}
}

func (cache *TTLCache) startCleanupTimer() {
	ticker := time.NewTicker(cache.cleanupInterval)
	go func() {
		for range ticker.C {
			cache.cleanup()
		}
	}()
}

// NewTTLCache creates a ttl cache.
func NewTTLCache() *TTLCache {
	cache := &TTLCache{
		cleanupInterval: time.Minute,
		items:           map[string]*Item{},
	}
	cache.startCleanupTimer()
	return cache
}
