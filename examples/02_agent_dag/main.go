package main

import (
	"context"
	"fmt"

	"github.com/thanhauco/ai-harness/pkg/harness"
)

func main() {
	dag, err := harness.NewPipelineBuilder().
		Step("fetch_docs", "Fetch Context", func(ctx context.Context, s *harness.ExecutionState) (any, error) {
			return "Context: High-performance Go microservices in 2026", nil
		}).
		Step("analyze", "Analyze Context", func(ctx context.Context, s *harness.ExecutionState) (any, error) {
			docs, _ := s.Get("output:fetch_docs")
			return fmt.Sprintf("Analysis of [%v]: verified robust.", docs), nil
		}, "fetch_docs").
		Build()

	if err != nil {
		panic(err)
	}

	state := harness.NewExecutionState()
	summary, err := dag.Execute(context.Background(), state, 2)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Pipeline finished: %d completed, %d failed, duration: %v\n",
		summary.Completed, summary.Failed, summary.Duration)
}
