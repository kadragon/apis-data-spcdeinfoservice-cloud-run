package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestInMemoryCache_GetSet(t *testing.T) {
	c := NewInMemoryCache(10 * time.Millisecond)
	defer c.Close()

	// Set & Get before expiration
	c.Set("foo", "bar", 50*time.Millisecond)
	val, found := c.Get("foo")
	if !found {
		t.Fatal("expected to find foo")
	}
	if val.(string) != "bar" {
		t.Fatalf("expected bar, got %v", val)
	}

	// Wait for expiration
	time.Sleep(60 * time.Millisecond)
	_, found = c.Get("foo")
	if found {
		t.Fatal("expected foo to be expired and not found")
	}
}

func TestInMemoryCache_Eviction(t *testing.T) {
	c := NewInMemoryCache(10 * time.Millisecond)
	defer c.Close()

	c.Set("short-lived", "data", 15*time.Millisecond)

	// Wait enough for cleaner goroutine to run
	time.Sleep(30 * time.Millisecond)

	// Since we sleep longer than TTL + clean interval, internal map should not leak the entry.
	// We can check if Get returns found=false.
	_, found := c.Get("short-lived")
	if found {
		t.Fatal("expected item to be evicted")
	}
}

func TestInMemoryCache_Delete(t *testing.T) {
	c := NewInMemoryCache(10 * time.Millisecond)
	defer c.Close()

	c.Set("foo", "bar", 10*time.Minute)
	c.Delete("foo")

	_, found := c.Get("foo")
	if found {
		t.Fatal("expected foo to be deleted")
	}
}

func TestInMemoryCache_ConcurrentAccess(t *testing.T) {
	c := NewInMemoryCache(0)
	defer c.Close()

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			c.Set(fmt.Sprintf("k%d", i), i, time.Minute)
		}(i)
		go func(i int) {
			defer wg.Done()
			c.Get(fmt.Sprintf("k%d", i))
		}(i)
	}
	wg.Wait()
}

func TestInMemoryCache_CloseIdempotent(t *testing.T) {
	c := NewInMemoryCache(10 * time.Millisecond)
	c.Close()
	c.Close() // must not panic
}
