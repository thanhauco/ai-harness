package harness

import (
	"context"
	"sync"
	"testing"
)

func TestExecutionState_Concurrency(t *testing.T) {
	state := NewExecutionState()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(2)
		val := i
		go func() {
			defer wg.Done()
			state.Set("counter", val)
		}()
		go func() {
			defer wg.Done()
			_, _ = state.Get("counter")
		}()
	}

	wg.Wait()
}

func TestRunSequential_Cancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	state := NewExecutionState()
	steps := []SequentialStep{
		{
			ID: "step1",
			Execute: func(ctx context.Context, s *ExecutionState) (any, error) {
				return "ok", nil
			},
		},
	}

	err := RunSequential(ctx, state, steps)
	if err == nil {
		t.Fatal("expected cancellation error, got nil")
	}
}
