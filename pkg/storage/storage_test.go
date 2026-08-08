package storage

import (
	"context"
	"os"
	"testing"
	"time"
)

type dummyResp struct {
	ID      string
	Content string
}

func TestLRUCache_EvictionAndTTL(t *testing.T) {
	cache := NewLRUCache(2, 50*time.Millisecond)

	r1 := &dummyResp{ID: "r1", Content: "one"}
	r2 := &dummyResp{ID: "r2", Content: "two"}
	r3 := &dummyResp{ID: "r3", Content: "three"}

	cache.Set("k1", r1)
	cache.Set("k2", r2)

	// Access k1 to make k2 oldest
	_, _ = cache.Get("k1")

	// Set k3, should evict k2
	cache.Set("k3", r3)

	if _, ok := cache.Get("k2"); ok {
		t.Fatal("expected k2 to be evicted")
	}
	if _, ok := cache.Get("k1"); !ok {
		t.Fatal("expected k1 to be present")
	}

	// Test TTL expiration
	time.Sleep(60 * time.Millisecond)
	if _, ok := cache.Get("k1"); ok {
		t.Fatal("expected k1 to be expired after TTL")
	}
}

func TestMemorySessionStore(t *testing.T) {
	store := NewMemorySessionStore()
	session := &SessionRecord{
		SessionID: "sess_1",
		Messages:  []any{"hello"},
		CreatedAt: time.Now(),
	}

	err := store.Save(context.Background(), session)
	if err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := store.Load(context.Background(), "sess_1")
	if err != nil || loaded.SessionID != "sess_1" {
		t.Fatalf("load error: %v", err)
	}

	// Missing
	_, errMissing := store.Load(context.Background(), "non_existent")
	if errMissing != os.ErrNotExist {
		t.Fatalf("expected ErrNotExist, got %v", errMissing)
	}
}
