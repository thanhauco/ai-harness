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

// PipelineSummary aggregates results of a DAG execution.
type PipelineSummary struct {
	TotalSteps int                     `json:"total_steps"`
	Completed  int                     `json:"completed"`
	Failed     int                     `json:"failed"`
	Skipped    int                     `json:"skipped"`
	Duration   time.Duration           `json:"duration"`
	Records    map[string]*StepRecord  `json:"records"`
}

func (d *DAG) Execute(ctx context.Context, state *ExecutionState, maxConcurrency int) (*PipelineSummary, error) {
	tiers, err := d.TopologicalSort()
	if err != nil {
		return nil, err
	}

	if maxConcurrency <= 0 {
		maxConcurrency = 4
	}

	start := time.Now()
	summary := &PipelineSummary{
		TotalSteps: len(d.steps),
		Records:    make(map[string]*StepRecord),
	}

	for _, tier := range tiers {
		if err := ctx.Err(); err != nil {
			summary.Duration = time.Since(start)
			return summary, err
		}

		sem := make(chan struct{}, maxConcurrency)
		var wg sync.WaitGroup
		errChan := make(chan error, len(tier))

		for _, stepID := range tier {
			d.mu.RLock()
			step := d.steps[stepID]
			d.mu.RUnlock()

			if step.SkipIf != nil && step.SkipIf(state) {
				rec := &StepRecord{
					StepID:     stepID,
					Status:     StepSkipped,
					StartedAt:  time.Now(),
					FinishedAt: time.Now(),
				}
				state.RecordStep(rec)
				summary.Skipped++
				summary.Records[stepID] = rec
				continue
			}

			wg.Add(1)
			sem <- struct{}{}

			go func(s Step) {
				defer wg.Done()
				defer func() { <-sem }()

				rec := &StepRecord{
					StepID:    s.ID,
					Status:    StepRunning,
					StartedAt: time.Now(),
				}

				if d.beforeStep != nil {
					d.beforeStep(s.ID, state)
				}

				out, execErr := s.Execute(ctx, state)
				rec.FinishedAt = time.Now()
				rec.Duration = rec.FinishedAt.Sub(rec.StartedAt)

				if d.afterStep != nil {
					d.afterStep(s.ID, out, execErr, rec.Duration)
				}

				if execErr != nil {
					rec.Status = StepFailed
					rec.Error = execErr.Error()
					state.RecordStep(rec)
					errChan <- fmt.Errorf("step %s failed: %w", s.ID, execErr)
				} else {
					rec.Status = StepCompleted
					rec.Output = out
					state.RecordStep(rec)
					state.Set("output:"+s.ID, out)
				}
			}(step)
		}

		wg.Wait()
		close(errChan)

		if len(errChan) > 0 {
			var firstErr error
			for e := range errChan {
				if firstErr == nil {
					firstErr = e
				}
				summary.Failed++
			}
			summary.Duration = time.Since(start)
			return summary, firstErr
		}

		for _, stepID := range tier {
			if rec, ok := state.GetRecord(stepID); ok {
				summary.Records[stepID] = rec
				if rec.Status == StepCompleted {
					summary.Completed++
				}
			}
		}
	}

	summary.Duration = time.Since(start)
	return summary, nil
}

// SetHooks configures pre- and post-step execution hooks.
func (d *DAG) SetHooks(
	before func(stepID string, state *ExecutionState),
	after func(stepID string, out any, err error, duration time.Duration),
) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.beforeStep = before
	d.afterStep = after
}

// StepCount returns the total number of registered steps.
func (d *DAG) StepCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.steps)
}

// HasStep returns true if stepID is registered.
func (d *DAG) HasStep(stepID string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, exists := d.steps[stepID]
	return exists
}
