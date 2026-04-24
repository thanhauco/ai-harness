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
