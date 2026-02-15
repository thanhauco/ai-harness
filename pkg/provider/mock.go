package provider

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/thanhauco/ai-harness/pkg/harness"
)

// MockProvider is an in-memory test implementation of Provider.
type MockProvider struct {
	name          string
	fixedResponse string
	callCount     atomic.Int64
}

func NewMockProvider(name, fixedResponse string) *MockProvider {
	if name == "" {
		name = "mock-provider"
	}
	return &MockProvider{
		name:          name,
		fixedResponse: fixedResponse,
	}
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) CallCount() int64 {
	return m.callCount.Load()
}

func (m *MockProvider) Generate(ctx context.Context, prompt *harness.Prompt) (*harness.Response, error) {
	m.callCount.Add(1)
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return &harness.Response{
		ID:      "mock-resp-1",
		Model:   m.name,
		Content: m.fixedResponse,
		Usage: harness.TokenUsage{
			PromptTokens:     10,
			CompletionTokens: len(m.fixedResponse) / 4,
			TotalTokens:      10 + len(m.fixedResponse)/4,
			DurationMs:       5,
		},
		FinishReason: harness.FinishStop,
		CreatedAt:    time.Now(),
	}, nil
}
