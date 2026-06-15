package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// CalculatorTool safely evaluates simple arithmetic expressions.
type CalculatorTool struct{}

func (c *CalculatorTool) Definition() ToolDefinition {
	return ToolDefinition{
		Name:        "calculator",
		Description: "Evaluates basic arithmetic expressions (e.g. '10 + 5 * 2')",
		Parameters: map[string]ParameterSchema{
			"expression": {
				Type:        "string",
				Description: "The arithmetic expression to evaluate",
				Required:    true,
			},
		},
	}
}

func (c *CalculatorTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	expr, ok := args["expression"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid expression argument")
	}

	parts := strings.Fields(expr)
	if len(parts) == 3 {
		left, err1 := strconv.ParseFloat(parts[0], 64)
		right, err2 := strconv.ParseFloat(parts[2], 64)
		if err1 == nil && err2 == nil {
			switch parts[1] {
			case "+":
				return left + right, nil
			case "-":
				return left - right, nil
			case "*":
				return left * right, nil
			case "/":
				if right == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				return left / right, nil
			}
		}
	}

	return nil, fmt.Errorf("unsupported expression format: %q", expr)
}
