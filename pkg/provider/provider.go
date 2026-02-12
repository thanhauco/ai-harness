package provider

import (
	"context"

	"github.com/thanhauco/ai-harness/pkg/harness"
)

// Provider abstracts language model backends.
type Provider interface {
	Name() string
	Generate(ctx context.Context, prompt *harness.Prompt) (*harness.Response, error)
}
