package harness

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/thanhauco/ai-harness/pkg/resilience"
	"github.com/thanhauco/ai-harness/pkg/storage"
)

type testMockCaller struct {
	calls int
}

func (m *testMockCaller) Name() string { return "mock" }
func (m *testMockCaller) Generate(ctx context.Context, p *Prompt) (*Response, error) {
	m.calls++
	return &Response{
		ID:      "test-1",
		Content: "test response",
		Usage:   TokenUsage{TotalTokens: 10},
	}, nil
}

func TestRunner_CacheHit(t *testing.T) {
	caller := &testMockCaller{}
	cache := storage.NewLRUCache(10, time.Hour)

	runner := NewRunner(RunnerConfig{
		Provider: caller,
		Cache:    cache,
	})

	prompt := NewPrompt(NewUserMessage("what is Go?"))

	// 1st call -> Cache Miss
	resp1, err1 := runner.ExecutePrompt(context.Background(), prompt)
	if err1 != nil {
		t.Fatalf("first call failed: %v", err1)
	}
	if resp1.Cached {
		t.Fatal("first call should not be cached")
	}

	// 2nd call -> Cache Hit
	resp2, err2 := runner.ExecutePrompt(context.Background(), prompt)
	if err2 != nil {
		t.Fatalf("second call failed: %v", err2)
	}
	if !resp2.Cached {
		t.Fatal("second call should be cached")
	}

	if caller.calls != 1 {
		t.Fatalf("expected exactly 1 provider invocation, got %d", caller.calls)
	}
}

type faultyCaller struct {
	attempts int
}

func (f *faultyCaller) Name() string { return "faulty" }
func (f *faultyCaller) Generate(ctx context.Context, p *Prompt) (*Response, error) {
	f.attempts++
	if f.attempts < 2 {
		return nil, errors.New("temporary 503 service unavailable")
	}
	return &Response{
		ID:      "recovered",
		Content: "recovered output",
		Usage:   TokenUsage{TotalTokens: 12},
	}, nil
}

func TestRunner_ResilienceRecovery(t *testing.T) {
	caller := &faultyCaller{}
	policy := &resilience.Policy{
		Retry: resilience.RetryConfig{
			MaxAttempts:     3,
			InitialInterval: 5 * time.Millisecond,
			MaxInterval:     20 * time.Millisecond,
			Multiplier:      1.5,
		},
	}

	runner := NewRunner(RunnerConfig{
		Provider: caller,
		Policy:   policy,
	})

	resp, err := runner.ExecutePrompt(context.Background(), NewPrompt(NewUserMessage("test")))
	if err != nil {
		t.Fatalf("expected runner to recover from transient error: %v", err)
	}

	if resp.Content != "recovered output" {
		t.Fatalf("unexpected content: %s", resp.Content)
	}
	if caller.attempts != 2 {
		t.Fatalf("expected 2 attempts before success, got %d", caller.attempts)
	}
}
