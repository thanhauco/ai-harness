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
