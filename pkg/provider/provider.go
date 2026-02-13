package provider

import (
	"context"
	"iter"

	"github.com/thanhauco/ai-harness/pkg/harness"
)

// StreamChunk represents an incremental delta during model generation.
type StreamChunk struct {
	Delta        string               `json:"delta"`
	FinishReason harness.FinishReason `json:"finish_reason,omitempty"`
	Usage        *harness.TokenUsage  `json:"usage,omitempty"`
}

// Provider abstracts language model backends with unary and streaming capabilities.
type Provider interface {
	Name() string
	Generate(ctx context.Context, prompt *harness.Prompt) (*harness.Response, error)
	Stream(ctx context.Context, prompt *harness.Prompt) iter.Seq2[StreamChunk, error]
}
