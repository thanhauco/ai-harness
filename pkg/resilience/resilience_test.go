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
