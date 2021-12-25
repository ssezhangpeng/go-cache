package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	cache := NewTTLCache()

	data, exists := cache.Get("hello")
	assert.False(t, exists)
	assert.Empty(t, data)

	cache.Set("hello", "world", 60)
	data, exists = cache.Get("hello")
	assert.True(t, exists)
	assert.Equal(t, "world", data)
}

func TestExpiration(t *testing.T) {
	cache := &TTLCache{
		cleanupInterval: time.Second,
		items:           map[string]*Item{},
	}
	cache.startCleanupTimer()

	cache.Set("x", "1", 1)
	cache.Set("y", "z", 1)
	cache.Set("z", "3", 1)
	cache.startCleanupTimer()

	count := cache.Count()
	assert.Equal(t, 3, count)

	<-time.After(500 * time.Millisecond)
	cache.mutex.Lock()
	cache.items["y"].touch(time.Second)
	item, exists := cache.items["x"]
	cache.mutex.Unlock()
	assert.True(t, exists)
	assert.Equal(t, "1", item.data)
	assert.False(t, item.expired(time.Now()))

	<-time.After(time.Second)
	cache.mutex.RLock()
	_, exists = cache.items["x"]
	assert.False(t, exists)
	_, exists = cache.items["z"]
	assert.False(t, exists)
	_, exists = cache.items["y"]
	assert.True(t, exists)
	cache.mutex.RUnlock()

	count = cache.Count()
	assert.Equal(t, 1, count)

	<-time.After(600 * time.Millisecond)
	cache.mutex.RLock()
	_, exists = cache.items["y"]
	assert.False(t, exists)
	cache.mutex.RUnlock()

	count = cache.Count()
	assert.Equal(t, 0, count)
}
