package cache

import (
	"sync"
	"time"
)

var _ Cache = (*InMemoryCache)(nil)

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
	mu        sync.RWMutex
	items     map[string]cacheItem
	cleanup   time.Duration
	stop      chan struct{}
	closeOnce sync.Once
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

// Set stores a value under key with the given TTL. A ttl of zero or less stores the item with no expiration.
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

// Close signals the background janitor to stop. Safe to call multiple times.
func (c *InMemoryCache) Close() {
	c.closeOnce.Do(func() { close(c.stop) })
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
