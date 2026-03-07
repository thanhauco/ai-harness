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
