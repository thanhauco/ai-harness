package resilience

import (
	"context"
	"errors"
	"sync"
	"time"
)

// TokenBucket implements an atomic, lock-managed token bucket rate limiter.
type TokenBucket struct {
	mu         sync.Mutex
	rate       float64
	capacity   float64
	tokens     float64
	lastRefill time.Time
}

func NewTokenBucket(rate, capacity float64) *TokenBucket {
	if rate <= 0 {
		rate = 10.0
	}
	if capacity <= 0 {
		capacity = rate
	}
	return &TokenBucket{
		rate:       rate,
		capacity:   capacity,
		tokens:     capacity,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.rate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastRefill = now
}
