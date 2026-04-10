package resilience

import (
	"context"
	"math/rand/v2"
	"strings"
	"time"
)

type RetryConfig struct {
	MaxAttempts     int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	Jitter          bool
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:     3,
		InitialInterval: 50 * time.Millisecond,
		MaxInterval:     2 * time.Second,
		Multiplier:      2.0,
		Jitter:          true,
	}
}

func ExecuteWithRetry(ctx context.Context, cfg RetryConfig, op func(ctx context.Context) error, isRetryable func(error) bool) error {
	var err error
	interval := cfg.InitialInterval

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		if err = ctx.Err(); err != nil {
			return err
		}

		err = op(ctx)
		if err == nil {
			return nil
		}

		if isRetryable != nil && !isRetryable(err) {
			return err
		}

		if attempt == cfg.MaxAttempts {
			break
		}

		backoff := interval
		if cfg.Jitter {
			backoff = time.Duration(float64(interval) * (0.5 + rand.Float64()*0.5))
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}

		interval = time.Duration(float64(interval) * cfg.Multiplier)
		if interval > cfg.MaxInterval {
			interval = cfg.MaxInterval
		}
	}

	return err
}

func IsStandardRetryableError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "503") ||
		strings.Contains(msg, "502") ||
		strings.Contains(msg, "connection reset")
}
