package storage

import (
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

type cacheEntry struct {
	key       string
	value     any
	expiresAt time.Time
}

// LRUCache provides a thread-safe, size-bounded in-memory response cache.
type LRUCache struct {
	mu        sync.RWMutex
	capacity  int
	ttl       time.Duration
	items     map[string]*list.Element
	evictList *list.List
	hits      int64
	misses    int64
}

func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	if capacity <= 0 {
		capacity = 1000
	}
	if ttl <= 0 {
		ttl = 1 * time.Hour
	}
	return &LRUCache{
		capacity:  capacity,
		ttl:       ttl,
		items:     make(map[string]*list.Element),
		evictList: list.New(),
	}
}

// HashKey generates a deterministic SHA-256 hash for any value.
func HashKey(v any) string {
	b, _ := json.Marshal(v)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func (c *LRUCache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, exists := c.items[key]
	if !exists {
		c.misses++
		return nil, false
	}

	entry := elem.Value.(*cacheEntry)
	if time.Now().After(entry.expiresAt) {
		c.removeElement(elem)
		c.misses++
		return nil, false
	}

	c.evictList.MoveToFront(elem)
	c.hits++
	return entry.value, true
}

func (c *LRUCache) Set(key string, val any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.items[key]; exists {
		c.evictList.MoveToFront(elem)
		elem.Value.(*cacheEntry).value = val
		elem.Value.(*cacheEntry).expiresAt = time.Now().Add(c.ttl)
		return
	}

	if c.evictList.Len() >= c.capacity {
		c.evictOldest()
	}

	entry := &cacheEntry{
		key:       key,
		value:     val,
		expiresAt: time.Now().Add(c.ttl),
	}
	elem := c.evictList.PushFront(entry)
	c.items[key] = elem
}

func (c *LRUCache) evictOldest() {
	elem := c.evictList.Back()
	if elem != nil {
		c.removeElement(elem)
	}
}

func (c *LRUCache) removeElement(elem *list.Element) {
	c.evictList.Remove(elem)
	entry := elem.Value.(*cacheEntry)
	delete(c.items, entry.key)
}

func (c *LRUCache) Stats() (hits, misses int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits, c.misses
}
