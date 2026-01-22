package harness

import (
	"errors"
	"fmt"
	"time"
)

// Role defines the participant role in a conversation.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleSystem, RoleUser, RoleAssistant, RoleTool:
		return true
	default:
		return false
	}
}

// Message represents a single conversational turn.
type Message struct {
	Role       Role           `json:"role"`
	Content    string         `json:"content"`
	Name       string         `json:"name,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// TokenUsage records execution resource consumption and timing.
type TokenUsage struct {
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	DurationMs       int64   `json:"duration_ms"`
	CostUSD          float64 `json:"cost_usd"`
}

// Prompt encapsulates parameters for model execution.
type Prompt struct {
	Messages    []Message      `json:"messages"`
	Model       string         `json:"model,omitempty"`
	Temperature float32        `json:"temperature,omitempty"`
	MaxTokens   int            `json:"max_tokens,omitempty"`
	StopWords   []string       `json:"stop_words,omitempty"`
	Options     map[string]any `json:"options,omitempty"`
}

// HarnessError is a structured error containing error classification and retry metadata.
type HarnessError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
	Err       error  `json:"-"`
}

func (e *HarnessError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *HarnessError) Unwrap() error {
	return e.Err
}

// Predefined harness sentinel errors.
var (
	ErrContextCanceled  = errors.New("operation canceled by context")
	ErrRateLimited      = &HarnessError{Code: "RATE_LIMITED", Message: "rate limit exceeded", Retryable: true}
	ErrCircuitOpen      = &HarnessError{Code: "CIRCUIT_OPEN", Message: "circuit breaker is open", Retryable: true}
	ErrTimeout          = &HarnessError{Code: "TIMEOUT", Message: "request timed out", Retryable: true}
	ErrProviderFailed   = &HarnessError{Code: "PROVIDER_FAILED", Message: "upstream model provider failed", Retryable: true}
	ErrValidationFailed = &HarnessError{Code: "VALIDATION_FAILED", Message: "input or schema validation failed", Retryable: false}
)

// Helper constructors
func NewSystemMessage(content string) Message {
	return Message{Role: RoleSystem, Content: content}
}

func NewUserMessage(content string) Message {
	return Message{Role: RoleUser, Content: content}
}

func NewAssistantMessage(content string) Message {
	return Message{Role: RoleAssistant, Content: content}
}

func NewToolMessage(toolCallID, content string) Message {
	return Message{Role: RoleTool, ToolCallID: toolCallID, Content: content}
}

func NewPrompt(messages ...Message) *Prompt {
	return &Prompt{
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   2048,
	}
}
