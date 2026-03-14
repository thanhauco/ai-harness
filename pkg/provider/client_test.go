package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/thanhauco/ai-harness/pkg/harness"
)

func TestHTTPProvider_Generate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"id":"test-123","model":"gpt-test","choices":[{"message":{"content":"Hello from mock server"},"finish_reason":"stop"}],"usage":{"prompt_tokens":5,"completion_tokens":4,"total_tokens":9}}`)
	}))
	defer ts.Close()

	opts := DefaultClientOptions()
	opts.BaseURL = ts.URL
	p := NewHTTPProvider("test-http", opts)

	resp, err := p.Generate(context.Background(), harness.NewPrompt(harness.NewUserMessage("Hi")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Content != "Hello from mock server" {
		t.Fatalf("content mismatch: %s", resp.Content)
	}
	if resp.Usage.TotalTokens != 9 {
		t.Fatalf("expected 9 tokens, got %d", resp.Usage.TotalTokens)
	}
}

func TestHTTPProvider_Stream(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			return
		}
		fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hello \"}}]}\n\n")
		flusher.Flush()
		fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"World\"}}]}\n\n")
		flusher.Flush()
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer ts.Close()

	opts := DefaultClientOptions()
	opts.BaseURL = ts.URL
	p := NewHTTPProvider("test-http-stream", opts)

	var sb strings.Builder
	for chunk, err := range p.Stream(context.Background(), harness.NewPrompt()) {
		if err != nil {
			t.Fatalf("unexpected stream error: %v", err)
		}
		sb.WriteString(chunk.Delta)
	}

	if sb.String() != "Hello World" {
		t.Fatalf("expected 'Hello World', got %q", sb.String())
	}
}
