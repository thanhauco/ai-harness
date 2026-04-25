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

// TopologicalSort returns steps grouped into parallel execution tiers (Kahn's algorithm).
func (d *DAG) TopologicalSort() ([][]string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	inDegree := make(map[string]int)
	adj := make(map[string][]string)

	for id := range d.steps {
		inDegree[id] = 0
	}

	for id, step := range d.steps {
		for _, dep := range step.Dependencies {
			if _, exists := d.steps[dep]; !exists {
				return nil, fmt.Errorf("step %q depends on non-existent step %q", id, dep)
			}
			adj[dep] = append(adj[dep], id)
			inDegree[id]++
		}
	}

	var tiers [][]string
	currentTier := make([]string, 0)
	for id, deg := range inDegree {
		if deg == 0 {
			currentTier = append(currentTier, id)
		}
	}

	visitedCount := 0
	for len(currentTier) > 0 {
		tiers = append(tiers, currentTier)
		visitedCount += len(currentTier)
		nextTier := make([]string, 0)

		for _, node := range currentTier {
			for _, neighbor := range adj[node] {
				inDegree[neighbor]--
				if inDegree[neighbor] == 0 {
					nextTier = append(nextTier, neighbor)
				}
			}
		}
		currentTier = nextTier
	}

	if visitedCount != len(d.steps) {
		return nil, errors.New("cycle detected in pipeline DAG")
	}

	return tiers, nil
}
