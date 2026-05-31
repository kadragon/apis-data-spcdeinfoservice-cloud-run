package cache

import (
	"sync"
	"time"
)

// Cache defines the caching operations required.
type Cache interface {
	Set(key string, val interface{}, ttl time.Duration)
	Get(key string) (interface{}, bool)
	Delete(key string)
	Close()
}

type cacheItem struct {
	value      interface{}
	expiration int64 // unix nano timestamp
}

func (item *cacheItem) expired() bool {
	if item.expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.expiration
}

// InMemoryCache provides a thread-safe, TTL-based caching layer.
type InMemoryCache struct {
	mu      sync.RWMutex
	items   map[string]cacheItem
	cleanup time.Duration
	stop    chan struct{}
}

// NewInMemoryCache instantiates a thread-safe memory cache.
// A janitor goroutine will run every cleanupInterval if greater than 0.
func NewInMemoryCache(cleanupInterval time.Duration) *InMemoryCache {
	c := &InMemoryCache{
		items:   make(map[string]cacheItem),
		cleanup: cleanupInterval,
		stop:    make(chan struct{}),
	}
	if cleanupInterval > 0 {
		go c.startJanitor()
	}
	return c
}

// Set stores a value inside the cache with a specified TTL.
func (c *InMemoryCache) Set(key string, val interface{}, ttl time.Duration) {
	var exp int64
	if ttl > 0 {
		exp = time.Now().Add(ttl).UnixNano()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = cacheItem{
		value:      val,
		expiration: exp,
	}
}

// Get retrieves a value if it exists and is not expired.
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	if item.expired() {
		return nil, false
	}

	return item.value, true
}

// Delete removes a key immediately from the store.
func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Close shuts down the janitor cleanup goroutine.
func (c *InMemoryCache) Close() {
	if c.stop != nil {
		close(c.stop)
	}
}

func (c *InMemoryCache) startJanitor() {
	ticker := time.NewTicker(c.cleanup)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.evictExpired()
		case <-c.stop:
			return
		}
	}
}

func (c *InMemoryCache) evictExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UnixNano()
	for k, item := range c.items {
		if item.expiration > 0 && now > item.expiration {
			delete(c.items, k)
		}
	}
}
