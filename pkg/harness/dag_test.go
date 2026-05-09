package harness

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestDAG_AddStep(t *testing.T) {
	dag := NewDAG()
	err := dag.AddStep(Step{
		ID: "step1",
		Execute: func(ctx context.Context, s *ExecutionState) (any, error) {
			return "step1 done", nil
		},
	})
	if err != nil {
		t.Fatalf("unexpected error adding step: %v", err)
	}

	// Duplicate check
	errDup := dag.AddStep(Step{ID: "step1"})
	if errDup == nil {
		t.Fatal("expected duplicate error, got nil")
	}
}

func TestDAG_TopologicalSort_Cycle(t *testing.T) {
	dag := NewDAG()
	_ = dag.AddStep(Step{ID: "A", Dependencies: []string{"B"}})
	_ = dag.AddStep(Step{ID: "B", Dependencies: []string{"A"}})

	_, err := dag.TopologicalSort()
	if err == nil {
		t.Fatal("expected cycle detection error")
	}
}

func TestDAG_TopologicalSort_Diamond(t *testing.T) {
	// A -> B, C -> D
	dag := NewDAG()
	_ = dag.AddStep(Step{ID: "A"})
	_ = dag.AddStep(Step{ID: "B", Dependencies: []string{"A"}})
	_ = dag.AddStep(Step{ID: "C", Dependencies: []string{"A"}})
	_ = dag.AddStep(Step{ID: "D", Dependencies: []string{"B", "C"}})

	tiers, err := dag.TopologicalSort()
	if err != nil {
		t.Fatalf("unexpected sort error: %v", err)
	}

	if len(tiers) != 3 {
		t.Fatalf("expected 3 tiers, got %d", len(tiers))
	}
	if len(tiers[0]) != 1 || tiers[0][0] != "A" {
		t.Fatalf("tier 0 mismatch: %v", tiers[0])
	}
	if len(tiers[1]) != 2 {
		t.Fatalf("tier 1 expected 2 parallel steps, got %d", len(tiers[1]))
	}
	if len(tiers[2]) != 1 || tiers[2][0] != "D" {
		t.Fatalf("tier 2 mismatch: %v", tiers[2])
	}
}

func TestDAG_ExecuteParallel(t *testing.T) {
	dag := NewDAG()
	state := NewExecutionState()

	var running atomic.Int32
	var maxObserved atomic.Int32

	for i := 0; i < 4; i++ {
		stepID := fmt.Sprintf("parallel_%d", i)
		_ = dag.AddStep(Step{
			ID: stepID,
			Execute: func(ctx context.Context, s *ExecutionState) (any, error) {
				cur := running.Add(1)
				for {
					oldMax := maxObserved.Load()
					if cur <= oldMax || maxObserved.CompareAndSwap(oldMax, cur) {
						break
					}
				}
				time.Sleep(20 * time.Millisecond)
				running.Add(-1)
				return "ok", nil
			},
		})
	}

	summary, err := dag.Execute(context.Background(), state, 4)
	if err != nil {
		t.Fatalf("unexpected execute error: %v", err)
	}

	if summary.Completed != 4 {
		t.Fatalf("expected 4 completed, got %d", summary.Completed)
	}
	if maxObserved.Load() < 2 {
		t.Fatalf("expected parallel execution (>=2 concurrent), observed %d", maxObserved.Load())
	}
}

func TestDAG_FailurePropagation(t *testing.T) {
	dag := NewDAG()
	state := NewExecutionState()

	_ = dag.AddStep(Step{
		ID: "step1",
		Execute: func(ctx context.Context, s *ExecutionState) (any, error) {
			return nil, errors.New("step1 boom")
		},
	})
	_ = dag.AddStep(Step{
		ID:           "step2",
		Dependencies: []string{"step1"},
		Execute: func(ctx context.Context, s *ExecutionState) (any, error) {
			return "should not run", nil
		},
	})

	summary, err := dag.Execute(context.Background(), state, 2)
	if err == nil {
		t.Fatal("expected failure error, got nil")
	}
	if summary.Completed != 0 {
		t.Fatalf("expected 0 completed steps, got %d", summary.Completed)
	}
}

func TestDAG_Hooks(t *testing.T) {
	dag := NewDAG()
	state := NewExecutionState()

	var beforeCount atomic.Int32
	var afterCount atomic.Int32

	dag.SetHooks(
		func(id string, s *ExecutionState) { beforeCount.Add(1) },
		func(id string, out any, err error, dur time.Duration) { afterCount.Add(1) },
	)

	_ = dag.AddStep(Step{
		ID: "stepA",
		Execute: func(ctx context.Context, s *ExecutionState) (any, error) {
			return "doneA", nil
		},
	})

	_, err := dag.Execute(context.Background(), state, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if beforeCount.Load() != 1 || afterCount.Load() != 1 {
		t.Fatalf("hooks count mismatch: before=%d, after=%d", beforeCount.Load(), afterCount.Load())
	}
}

func TestDAG_SkipIf(t *testing.T) {
	dag := NewDAG()
	state := NewExecutionState()
	state.Set("skip_next", true)

	_ = dag.AddStep(Step{
		ID: "step_conditional",
		SkipIf: func(s *ExecutionState) bool {
			v, _ := s.Get("skip_next")
			return v == true
		},
		Execute: func(ctx context.Context, s *ExecutionState) (any, error) {
			return "ran", nil
		},
	})

	summary, err := dag.Execute(context.Background(), state, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.Skipped != 1 {
		t.Fatalf("expected 1 skipped step, got %d", summary.Skipped)
	}
}
