package sandbox

import (
	"context"
	"fmt"
)

// ParameterSchema defines constraints for a single tool argument.
type ParameterSchema struct {
	Type        string   `json:"type"` // "string", "number", "boolean", "array", "object"
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	Enum        []string `json:"enum,omitempty"`
}

// ToolDefinition describes metadata and expected arguments for a tool.
type ToolDefinition struct {
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Parameters  map[string]ParameterSchema `json:"parameters"`
}
