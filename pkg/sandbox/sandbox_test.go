package sandbox

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestToolDefinition_Validation(t *testing.T) {
	tool := ToolDefinition{
		Name: "query_database",
		Parameters: map[string]ParameterSchema{
			"table": {Type: "string", Required: true},
			"limit": {Type: "number", Required: false},
		},
	}

	// Missing required table
	err := tool.ValidateArguments(map[string]any{"limit": 10})
	if err == nil || !strings.Contains(err.Error(), "missing required parameter") {
		t.Fatalf("expected missing required parameter error, got %v", err)
	}

	// Valid arguments
	errValid := tool.ValidateArguments(map[string]any{"table": "users", "limit": 25})
	if errValid != nil {
		t.Fatalf("unexpected error for valid arguments: %v", errValid)
	}
}

func TestToolDefinition_EnumAndTypeValidation(t *testing.T) {
	tool := ToolDefinition{
		Name: "set_mode",
		Parameters: map[string]ParameterSchema{
			"mode": {Type: "string", Required: true, Enum: []string{"read", "write"}},
		},
	}

	// Invalid enum
	err := tool.ValidateArguments(map[string]any{"mode": "admin"})
	if err == nil || !strings.Contains(err.Error(), "not in allowed enum") {
		t.Fatalf("expected enum validation failure, got %v", err)
	}

	// Type mismatch (number instead of string)
	errType := tool.ValidateArguments(map[string]any{"mode": 123})
	if errType == nil || !strings.Contains(errType.Error(), "must be a string") {
		t.Fatalf("expected type mismatch error, got %v", errType)
	}
}

type dummyTool struct{}

func (d *dummyTool) Definition() ToolDefinition {
	return ToolDefinition{Name: "dummy"}
}

func (d *dummyTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	return "dummy executed", nil
}

func TestRegistry_RegisterAndExecute(t *testing.T) {
	reg := NewRegistry()
	err := reg.Register(&dummyTool{})
	if err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	// Duplicate
	errDup := reg.Register(&dummyTool{})
	if errDup == nil {
		t.Fatal("expected duplicate error")
	}

	out, errExec := reg.Execute(context.Background(), "dummy", nil)
	if errExec != nil || out != "dummy executed" {
		t.Fatalf("unexpected execution result: %v, %v", out, errExec)
	}
}

func TestRunner_Timeout(t *testing.T) {
	policy := SandboxPolicy{
		Timeout: 20 * time.Millisecond,
	}
	runner := NewRunner(policy)

	// 'sleep 1' exceeds 20ms timeout
	_, err := runner.ExecuteCommand(context.Background(), "sleep", "1")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestRunner_DisallowedCommand(t *testing.T) {
	policy := SandboxPolicy{
		Timeout:         time.Second,
		AllowedCommands: []string{"echo"},
	}
	runner := NewRunner(policy)

	_, err := runner.ExecuteCommand(context.Background(), "ls")
	if err == nil || !strings.Contains(err.Error(), "disallowed") {
		t.Fatalf("expected disallowed command error, got %v", err)
	}
}
