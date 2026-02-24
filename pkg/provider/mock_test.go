package provider

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/thanhauco/ai-harness/pkg/harness"
)

func TestMockProvider_Stream(t *testing.T) {
	content := "The quick brown fox jumps over the lazy dog"
	p := NewMockProvider("test-mock", content)

	var sb strings.Builder
	var lastChunk StreamChunk

	for chunk, err := range p.Stream(context.Background(), harness.NewPrompt()) {
		if err != nil {
			t.Fatalf("unexpected stream error: %v", err)
		}
		sb.WriteString(chunk.Delta)
		lastChunk = chunk
	}

	if sb.String() != content {
		t.Fatalf("stream content mismatch: expected %q, got %q", content, sb.String())
	}

	if lastChunk.FinishReason != harness.FinishStop {
		t.Fatalf("expected finish reason stop, got %s", lastChunk.FinishReason)
	}
}

func TestMockProvider_StreamCancellation(t *testing.T) {
	p := NewMockProvider("test-mock", "Word1 Word2 Word3 Word4 Word5")
	p.SetChunkDelay(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	var gotError bool
	for _, err := range p.Stream(ctx, harness.NewPrompt()) {
		if err != nil {
			gotError = true
			break
		}
	}

	if !gotError {
		t.Fatal("expected cancellation error in stream iterator")
	}
}
