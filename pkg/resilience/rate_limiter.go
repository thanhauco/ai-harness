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

func (tb *TokenBucket) Allow(cost float64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()
	if tb.tokens >= cost {
		tb.tokens -= cost
		return true
	}
	return false
}

var ErrRateLimitExceeded = errors.New("rate limit wait canceled or timed out")

func (tb *TokenBucket) Wait(ctx context.Context, cost float64) error {
	for {
		tb.mu.Lock()
		tb.refill()
		if tb.tokens >= cost {
			tb.tokens -= cost
			tb.mu.Unlock()
			return nil
		}

		missing := cost - tb.tokens
		waitTime := time.Duration((missing / tb.rate) * float64(time.Second))
		tb.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
		}
	}
}
