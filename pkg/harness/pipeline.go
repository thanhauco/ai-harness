package harness

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type StepStatus string

const (
	StepPending   StepStatus = "pending"
	StepRunning   StepStatus = "running"
	StepCompleted StepStatus = "completed"
	StepFailed    StepStatus = "failed"
	StepSkipped   StepStatus = "skipped"
)

// StepRecord captures execution telemetry for a single pipeline step.
type StepRecord struct {
	StepID     string        `json:"step_id"`
	Status     StepStatus    `json:"status"`
	Output     any           `json:"output,omitempty"`
	Error      string        `json:"error,omitempty"`
	Duration   time.Duration `json:"duration"`
	StartedAt  time.Time     `json:"started_at"`
	FinishedAt time.Time     `json:"finished_at"`
}

// ExecutionState maintains shared thread-safe state during pipeline execution.
type ExecutionState struct {
	mu      sync.RWMutex
	data    map[string]any
	records map[string]*StepRecord
}

func NewExecutionState() *ExecutionState {
	return &ExecutionState{
		data:    make(map[string]any),
		records: make(map[string]*StepRecord),
	}
}

func (s *ExecutionState) Set(key string, val any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = val
}

func (s *ExecutionState) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

func (s *ExecutionState) GetString(key string) (string, bool) {
	v, ok := s.Get(key)
	if !ok {
		return "", false
	}
	str, ok := v.(string)
	return str, ok
}

func (s *ExecutionState) RecordStep(record *StepRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[record.StepID] = record
}

func (s *ExecutionState) GetRecord(stepID string) (*StepRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, ok := s.records[stepID]
	return rec, ok
}

// SequentialStep is an actionable function block.
type SequentialStep struct {
	ID      string
	Execute func(ctx context.Context, state *ExecutionState) (any, error)
}

func RunSequential(ctx context.Context, state *ExecutionState, steps []SequentialStep) error {
	for _, s := range steps {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := &StepRecord{
			StepID:    s.ID,
			Status:    StepRunning,
			StartedAt: time.Now(),
		}
		out, err := s.Execute(ctx, state)
		rec.FinishedAt = time.Now()
		rec.Duration = rec.FinishedAt.Sub(rec.StartedAt)
		if err != nil {
			rec.Status = StepFailed
			rec.Error = err.Error()
			state.RecordStep(rec)
			return fmt.Errorf("step %s failed: %w", s.ID, err)
		}
		rec.Status = StepCompleted
		rec.Output = out
		state.RecordStep(rec)
		state.Set("output:"+s.ID, out)
	}
	return nil
}
