package cache

import (
	"sync"
	"time"
)

// Item represents an item in the cache.
type Item struct {
	sync.RWMutex
	data    interface{}
	expires time.Time
	ttl     time.Duration
}

func (item *Item) touch(ttl time.Duration) {
	item.Lock()
	item.ttl = ttl
	item.expires = time.Now().Add(ttl)
	item.Unlock()
}

func (item *Item) expired(now time.Time) bool {
	item.RLock()
	defer item.RUnlock()

	return item.expires.Before(now)
}
