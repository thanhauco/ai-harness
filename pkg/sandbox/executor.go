package sandbox

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Registry manages registered tools.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

func (r *Registry) Register(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	def := tool.Definition()
	if def.Name == "" {
		return errors.New("tool name cannot be empty")
	}
	if _, exists := r.tools[def.Name]; exists {
		return fmt.Errorf("tool already registered: %s", def.Name)
	}
	r.tools[def.Name] = tool
	return nil
}

func (r *Registry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, ok := r.tools[name]
	return tool, ok
}

func (r *Registry) Execute(ctx context.Context, name string, args map[string]any) (any, error) {
	r.mu.RLock()
	tool, ok := r.tools[name]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	def := tool.Definition()
	if err := def.ValidateArguments(args); err != nil {
		return nil, fmt.Errorf("invalid arguments for tool %s: %w", name, err)
	}

	return tool.Execute(ctx, args)
}
