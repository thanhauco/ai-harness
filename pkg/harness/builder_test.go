package harness

import (
	"context"
	"testing"
)

func TestPipelineBuilder_FluentAPI(t *testing.T) {
	dag, err := NewPipelineBuilder().
		Step("fetch", "Fetch Data", func(ctx context.Context, s *ExecutionState) (any, error) {
			return "data", nil
		}).
		Step("process", "Process Data", func(ctx context.Context, s *ExecutionState) (any, error) {
			return "processed", nil
		}, "fetch").
		Build()

	if err != nil {
		t.Fatalf("unexpected builder error: %v", err)
	}

	if dag.StepCount() != 2 {
		t.Fatalf("expected 2 steps, got %d", dag.StepCount())
	}
}
