package resilience

import (
	"errors"
	"sync"
	"time"
)

type CircuitState int32

const (
	StateClosed CircuitState = iota
	StateHalfOpen
	StateOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateHalfOpen:
		return "HALF_OPEN"
	case StateOpen:
		return "OPEN"
	default:
		return "UNKNOWN"
	}
}

type CircuitBreakerConfig struct {
	FailureThreshold int
	RecoveryTimeout  time.Duration
	HalfOpenMaxCalls int
}

type CircuitBreaker struct {
	mu              sync.Mutex
	cfg             CircuitBreakerConfig
	state           CircuitState
	failures        int
	successes       int
	halfOpenCalls   int
	lastFailureTime time.Time
}

func NewCircuitBreaker(cfg CircuitBreakerConfig) *CircuitBreaker {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 5
	}
	if cfg.RecoveryTimeout <= 0 {
		cfg.RecoveryTimeout = 5 * time.Second
	}
	if cfg.HalfOpenMaxCalls <= 0 {
		cfg.HalfOpenMaxCalls = 2
	}
	return &CircuitBreaker{
		cfg:   cfg,
		state: StateClosed,
	}
}

var ErrCircuitBreakerOpen = errors.New("circuit breaker is open; upstream requests rejected")

func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.checkStateTransition()
	return cb.state
}

func (cb *CircuitBreaker) checkStateTransition() {
	if cb.state == StateOpen && time.Since(cb.lastFailureTime) >= cb.cfg.RecoveryTimeout {
		cb.state = StateHalfOpen
		cb.halfOpenCalls = 0
		cb.successes = 0
	}
}
