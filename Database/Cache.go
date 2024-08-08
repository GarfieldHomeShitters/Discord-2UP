package Database

import (
	"sync"
	"time"
)

type CacheItem[V any] struct {
	Value      V
	LastAccess int64
}
type Cache[K comparable, V any] struct {
	items map[K]CacheItem[V]
	mu    sync.RWMutex
	ttl   time.Duration
}

func NewCache[K comparable, V any](ttl time.Duration) *Cache[K, V] {
	return &Cache[K, V]{
		items: make(map[K]CacheItem[V]),
		ttl:   ttl,
	}
}

func (c *Cache[K, V]) Set(key K, value V) {
	currTime := time.Now().Unix()
	c.mu.Lock()
	c.items[key] = CacheItem[V]{
		Value:      value,
		LastAccess: currTime,
	}
	c.mu.Unlock()
}

func (c *Cache[K, V]) Get(key K) (value V, ok bool) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()

	if !ok {
		var blankVal V
		return blankVal, false
	}

	item.LastAccess = time.Now().UnixNano()
	c.mu.Lock()
	c.items[key] = item
	c.mu.Unlock()
	return item.Value, true
}

func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

func (c *Cache[K, V]) Cleanup() {
	currTime := time.Now().UnixNano()
	expiredKeys := make([]K, 0)

	c.mu.RLock()
	for key, item := range c.items {
		if currTime-c.ttl.Nanoseconds() > item.LastAccess {
			expiredKeys = append(expiredKeys, key)
		}
	}
	c.mu.RUnlock()

	for _, key := range expiredKeys {
		c.Delete(key)
	}
}

func (c *Cache[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]K, 0, len(c.items))
	for k := range c.items {
		keys = append(keys, k)
	}
	return keys
}
