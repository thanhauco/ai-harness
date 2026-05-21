package harness

import (
	"context"
)

// PipelineBuilder provides a fluent API to assemble DAG workflows.
type PipelineBuilder struct {
	dag *DAG
	err error
}

func NewPipelineBuilder() *PipelineBuilder {
	return &PipelineBuilder{
		dag: NewDAG(),
	}
}

func (b *PipelineBuilder) Step(id, name string, action func(ctx context.Context, state *ExecutionState) (any, error), deps ...string) *PipelineBuilder {
	if b.err != nil {
		return b
	}
	err := b.dag.AddStep(Step{
		ID:           id,
		Name:         name,
		Dependencies: deps,
		Execute:      action,
	})
	if err != nil {
		b.err = err
	}
	return b
}

func (b *PipelineBuilder) Build() (*DAG, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.dag, nil
}
