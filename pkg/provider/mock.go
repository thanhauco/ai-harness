package provider

import (
	"context"
	"iter"
	"strings"
	"sync/atomic"
	"time"

	"github.com/thanhauco/ai-harness/pkg/harness"
)

// MockProvider is an in-memory test implementation of Provider.
type MockProvider struct {
	name          string
	fixedResponse string
	latency       time.Duration
	chunkDelay    time.Duration
	promptTokens  int
	complTokens   int
	callCount     atomic.Int64
}

func NewMockProvider(name, fixedResponse string) *MockProvider {
	if name == "" {
		name = "mock-provider"
	}
	return &MockProvider{
		name:          name,
		fixedResponse: fixedResponse,
		promptTokens:  12,
		complTokens:   len(fixedResponse) / 4,
	}
}

func (m *MockProvider) SetTokens(prompt, compl int) {
	m.promptTokens = prompt
	m.complTokens = compl
}

func (m *MockProvider) SetLatency(d time.Duration) {
	m.latency = d
}

func (m *MockProvider) SetChunkDelay(d time.Duration) {
	m.chunkDelay = d
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) CallCount() int64 {
	return m.callCount.Load()
}

func (m *MockProvider) Generate(ctx context.Context, prompt *harness.Prompt) (*harness.Response, error) {
	m.callCount.Add(1)

	if m.latency > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(m.latency):
		}
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return &harness.Response{
		ID:      "mock-resp-1",
		Model:   m.name,
		Content: m.fixedResponse,
		Usage: harness.TokenUsage{
			PromptTokens:     m.promptTokens,
			CompletionTokens: m.complTokens,
			TotalTokens:      m.promptTokens + m.complTokens,
			DurationMs:       m.latency.Milliseconds(),
		},
		FinishReason: harness.FinishStop,
		CreatedAt:    time.Now(),
	}, nil
}

func (m *MockProvider) Stream(ctx context.Context, prompt *harness.Prompt) iter.Seq2[StreamChunk, error] {
	return func(yield func(StreamChunk, error) bool) {
		m.callCount.Add(1)
		words := strings.Split(m.fixedResponse, " ")

		for i, word := range words {
			if m.chunkDelay > 0 {
				select {
				case <-ctx.Done():
					yield(StreamChunk{}, ctx.Err())
					return
				case <-time.After(m.chunkDelay):
				}
			}

			if ctx.Err() != nil {
				yield(StreamChunk{}, ctx.Err())
				return
			}

			delta := word
			if i < len(words)-1 {
				delta += " "
			}

			chunk := StreamChunk{
				Delta: delta,
			}
			if i == len(words)-1 {
				chunk.FinishReason = harness.FinishStop
				chunk.Usage = &harness.TokenUsage{
					PromptTokens:     m.promptTokens,
					CompletionTokens: len(words),
					TotalTokens:      m.promptTokens + len(words),
				}
			}

			if !yield(chunk, nil) {
				return
			}
		}
	}
}
