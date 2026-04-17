package resilience

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestCircuitBreaker_Tripping(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 3,
		RecoveryTimeout:  100 * time.Millisecond,
		HalfOpenMaxCalls: 1,
	})

	errDummy := errors.New("upstream failure")

	for i := 0; i < 3; i++ {
		_ = cb.Execute(func() error {
			return errDummy
		})
	}

	if cb.State() != StateOpen {
		t.Fatalf("expected state OPEN, got %s", cb.State())
	}

	// Immediate call should be rejected
	err := cb.Execute(func() error {
		return nil
	})
	if !errors.Is(err, ErrCircuitBreakerOpen) {
		t.Fatalf("expected ErrCircuitBreakerOpen, got %v", err)
	}
}

func TestCircuitBreaker_Recovery(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 2,
		RecoveryTimeout:  20 * time.Millisecond,
		HalfOpenMaxCalls: 2,
	})

	_ = cb.Execute(func() error { return errors.New("fail") })
	_ = cb.Execute(func() error { return errors.New("fail") })

	if cb.State() != StateOpen {
		t.Fatalf("expected OPEN state")
	}

	// Wait for recovery timeout
	time.Sleep(30 * time.Millisecond)

	// In half-open state, two successful probes should reset to closed
	_ = cb.Execute(func() error { return nil })
	_ = cb.Execute(func() error { return nil })

	if cb.State() != StateClosed {
		t.Fatalf("expected CLOSED state after probe successes, got %s", cb.State())
	}
}

func TestTokenBucket_Burst(t *testing.T) {
	tb := NewTokenBucket(10, 5)

	// Consume entire burst capacity of 5 tokens
	for i := 0; i < 5; i++ {
		if !tb.Allow(1) {
			t.Fatalf("expected token %d to be allowed", i+1)
		}
	}

	// 6th token must fail immediately
	if tb.Allow(1) {
		t.Fatal("expected 6th token to exceed burst capacity")
	}
}

func TestTokenBucket_WaitCancellation(t *testing.T) {
	tb := NewTokenBucket(1, 1)
	_ = tb.Allow(1) // drain bucket

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := tb.Wait(ctx, 10) // needs 10 seconds, timeout is 20ms
	if err == nil {
		t.Fatal("expected timeout cancellation error")
	}
}

func TestExecuteWithRetry_SuccessAfterTransient(t *testing.T) {
	attempts := 0
	cfg := RetryConfig{
		MaxAttempts:     3,
		InitialInterval: 5 * time.Millisecond,
		MaxInterval:     20 * time.Millisecond,
		Multiplier:      1.5,
		Jitter:          false,
	}

	err := ExecuteWithRetry(context.Background(), cfg, func(ctx context.Context) error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary 503")
		}
		return nil
	}, IsStandardRetryableError)

	if err != nil {
		t.Fatalf("expected eventual success, got error: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestExecuteWithRetry_NonRetryableAborts(t *testing.T) {
	attempts := 0
	cfg := DefaultRetryConfig()
	fatalErr := errors.New("invalid api key 401")

	err := ExecuteWithRetry(context.Background(), cfg, func(ctx context.Context) error {
		attempts++
		return fatalErr
	}, IsStandardRetryableError)

	if !errors.Is(err, fatalErr) {
		t.Fatalf("expected fatalErr, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt before abort, got %d", attempts)
	}
}

func TestPolicy_ConcurrencyStress(t *testing.T) {
	policy := &Policy{
		Limiter: NewTokenBucket(1000, 1000),
		Breaker: NewCircuitBreaker(CircuitBreakerConfig{
			FailureThreshold: 50,
			RecoveryTimeout:  10 * time.Millisecond,
		}),
		Retry: DefaultRetryConfig(),
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = policy.Execute(context.Background(), func(ctx context.Context) error {
				return nil
			})
		}()
	}
	wg.Wait()
}
