package resilience

import (
	"context"
)

// Policy combines rate limiting, circuit breaker, and retry logic.
type Policy struct {
	Limiter *TokenBucket
	Breaker *CircuitBreaker
	Retry   RetryConfig
}

func (p *Policy) Execute(ctx context.Context, op func(ctx context.Context) error) error {
	if p.Limiter != nil {
		if err := p.Limiter.Wait(ctx, 1.0); err != nil {
			return err
		}
	}

	return ExecuteWithRetry(ctx, p.Retry, func(ctx context.Context) error {
		if p.Breaker != nil {
			return p.Breaker.Execute(func() error {
				return op(ctx)
			})
		}
		return op(ctx)
	}, IsStandardRetryableError)
}
