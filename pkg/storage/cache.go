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
