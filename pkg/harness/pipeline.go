package harness

import (
	"context"
	"sync"
	"time"
)

// ExecutionState maintains shared thread-safe state during pipeline execution.
type ExecutionState struct {
	mu   sync.RWMutex
	data map[string]any
}

func NewExecutionState() *ExecutionState {
	return &ExecutionState{
		data: make(map[string]any),
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
