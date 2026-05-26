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

func (t *ToolDefinition) ValidateArguments(args map[string]any) error {
	for name, param := range t.Parameters {
		val, exists := args[name]
		if param.Required && (!exists || val == nil) {
			return fmt.Errorf("missing required parameter: %s", name)
		}
		if !exists || val == nil {
			continue
		}

		switch param.Type {
		case "string":
			s, ok := val.(string)
			if !ok {
				return fmt.Errorf("parameter %s must be a string, got %T", name, val)
			}
			if len(param.Enum) > 0 {
				found := false
				for _, e := range param.Enum {
					if s == e {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("parameter %s value %q not in allowed enum: %v", name, s, param.Enum)
				}
			}
		case "number":
			switch val.(type) {
			case float64, float32, int, int64, int32:
			default:
				return fmt.Errorf("parameter %s must be a number, got %T", name, val)
			}
		case "boolean":
			if _, ok := val.(bool); !ok {
				return fmt.Errorf("parameter %s must be a boolean, got %T", name, val)
			}
		}
	}
	return nil
}
