package harness

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Step defines a node in the DAG.
type Step struct {
	ID           string
	Name         string
	Dependencies []string
	SkipIf       func(state *ExecutionState) bool
	Execute      func(ctx context.Context, state *ExecutionState) (any, error)
}

// DAG manages directed acyclic graph execution.
type DAG struct {
	mu         sync.RWMutex
	steps      map[string]Step
	beforeStep func(stepID string, state *ExecutionState)
	afterStep  func(stepID string, out any, err error, duration time.Duration)
}

func NewDAG() *DAG {
	return &DAG{
		steps: make(map[string]Step),
	}
}

func (d *DAG) AddStep(step Step) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if step.ID == "" {
		return errors.New("step ID cannot be empty")
	}
	if _, exists := d.steps[step.ID]; exists {
		return fmt.Errorf("duplicate step ID: %s", step.ID)
	}
	d.steps[step.ID] = step
	return nil
}
