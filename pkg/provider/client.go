package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"strings"
	"time"

	"github.com/thanhauco/ai-harness/pkg/harness"
)

type ClientOptions struct {
	BaseURL    string
	APIKey     string
	Timeout    time.Duration
	MaxIdle    int
	HTTPClient *http.Client
}

func DefaultClientOptions() ClientOptions {
	return ClientOptions{
		Timeout: 30 * time.Second,
		MaxIdle: 100,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

type openAIChatRequest struct {
	Model       string            `json:"model"`
	Messages    []harness.Message `json:"messages"`
	Temperature float32           `json:"temperature,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Stream      bool              `json:"stream"`
}

type openAIChatResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type openAIStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

type HTTPProvider struct {
	name string
	opts ClientOptions
}

func NewHTTPProvider(name string, opts ClientOptions) *HTTPProvider {
	if opts.HTTPClient == nil {
		opts = DefaultClientOptions()
	}
	return &HTTPProvider{
		name: name,
		opts: opts,
	}
}

func (h *HTTPProvider) Name() string {
	return h.name
}

func (h *HTTPProvider) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if h.opts.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.opts.APIKey)
	}
}

func (h *HTTPProvider) Generate(ctx context.Context, prompt *harness.Prompt) (*harness.Response, error) {
	reqBody := openAIChatRequest{
		Model:       prompt.Model,
		Messages:    prompt.Messages,
		Temperature: prompt.Temperature,
		MaxTokens:   prompt.MaxTokens,
		Stream:      false,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.opts.BaseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	h.setHeaders(req)

	start := time.Now()
	resp, err := h.opts.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upstream error (%d): %s", resp.StatusCode, string(body))
	}

	var chatResp openAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	content := ""
	finishReason := harness.FinishStop
	if len(chatResp.Choices) > 0 {
		content = chatResp.Choices[0].Message.Content
		finishReason = harness.FinishReason(chatResp.Choices[0].FinishReason)
	}

	return &harness.Response{
		ID:      chatResp.ID,
		Model:   chatResp.Model,
		Content: content,
		Usage: harness.TokenUsage{
			PromptTokens:     chatResp.Usage.PromptTokens,
			CompletionTokens: chatResp.Usage.CompletionTokens,
			TotalTokens:      chatResp.Usage.TotalTokens,
			DurationMs:       time.Since(start).Milliseconds(),
		},
		FinishReason: finishReason,
		CreatedAt:    time.Now(),
	}, nil
}

func (h *HTTPProvider) Stream(ctx context.Context, prompt *harness.Prompt) iter.Seq2[StreamChunk, error] {
	return func(yield func(StreamChunk, error) bool) {
		reqBody := openAIChatRequest{
			Model:       prompt.Model,
			Messages:    prompt.Messages,
			Temperature: prompt.Temperature,
			MaxTokens:   prompt.MaxTokens,
			Stream:      true,
		}

		payload, err := json.Marshal(reqBody)
		if err != nil {
			yield(StreamChunk{}, fmt.Errorf("marshal stream request: %w", err))
			return
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.opts.BaseURL+"/chat/completions", bytes.NewReader(payload))
		if err != nil {
			yield(StreamChunk{}, fmt.Errorf("create stream request: %w", err))
			return
		}

		h.setHeaders(req)

		resp, err := h.opts.HTTPClient.Do(req)
		if err != nil {
			yield(StreamChunk{}, fmt.Errorf("execute stream request: %w", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			yield(StreamChunk{}, fmt.Errorf("upstream stream error (%d): %s", resp.StatusCode, string(body)))
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			if ctx.Err() != nil {
				yield(StreamChunk{}, ctx.Err())
				return
			}

			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}

			if line == "data: [DONE]" {
				yield(StreamChunk{FinishReason: harness.FinishStop}, nil)
				return
			}

			if strings.HasPrefix(line, "data: ") {
				dataStr := strings.TrimPrefix(line, "data: ")
				var chunk openAIStreamChunk
				if err := json.Unmarshal([]byte(dataStr), &chunk); err == nil {
					if len(chunk.Choices) > 0 {
						delta := chunk.Choices[0].Delta.Content
						fReason := harness.FinishReason(chunk.Choices[0].FinishReason)
						if !yield(StreamChunk{Delta: delta, FinishReason: fReason}, nil) {
							return
						}
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			yield(StreamChunk{}, err)
		}
	}
}

// AnthropicMessagesPayload formats requests according to the Anthropic Messages spec.
type AnthropicMessagesPayload struct {
	Model     string            `json:"model"`
	Messages  []harness.Message `json:"messages"`
	MaxTokens int               `json:"max_tokens"`
}

func isPrematureStreamTermination(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "unexpected EOF") || strings.Contains(err.Error(), "connection closed")
}

// MaxBufferSize defines 64KB scan buffer for large SSE token payloads.
const MaxStreamBufferSize = 64 * 1024
