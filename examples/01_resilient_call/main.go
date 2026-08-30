package main

import (
	"context"
	"fmt"
	"time"

	"github.com/thanhauco/ai-harness/pkg/harness"
	"github.com/thanhauco/ai-harness/pkg/provider"
	"github.com/thanhauco/ai-harness/pkg/resilience"
	"github.com/thanhauco/ai-harness/pkg/storage"
)

func main() {
	p := provider.NewMockProvider("claude-3-5-sonnet", "Here is a fault-tolerant microservice architecture.")

	runner := harness.NewRunner(harness.RunnerConfig{
		Provider: p,
		Policy: &resilience.Policy{
			Limiter: resilience.NewTokenBucket(10, 5),
			Breaker: resilience.NewCircuitBreaker(resilience.CircuitBreakerConfig{
				FailureThreshold: 3,
				RecoveryTimeout:  2 * time.Second,
			}),
			Retry: resilience.DefaultRetryConfig(),
		},
		Cache: storage.NewLRUCache(100, time.Hour),
	})

	prompt := harness.NewPrompt(harness.NewUserMessage("Design a high-availability event bus"))
	resp, err := runner.ExecutePrompt(context.Background(), prompt)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Response: %s\nTokens: %d, Cached: %v\n", resp.Content, resp.Usage.TotalTokens, resp.Cached)
}
