package harness

import (
	"errors"
	"testing"
)

func TestHarnessError_Formatting(t *testing.T) {
	inner := errors.New("connection reset by peer")
	hErr := &HarnessError{
		Code:      "UPSTREAM_ERROR",
		Message:   "network drop",
		Retryable: true,
		Err:       inner,
	}

	expected := "[UPSTREAM_ERROR] network drop: connection reset by peer"
	if hErr.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, hErr.Error())
	}

	if !errors.Is(hErr, inner) {
		t.Fatalf("expected errors.Is to match unwrapped error")
	}
}

func TestTokenUsage_Calculation(t *testing.T) {
	u := TokenUsage{
		PromptTokens:     150,
		CompletionTokens: 50,
		TotalTokens:      200,
		DurationMs:       420,
	}

	if u.TotalTokens != u.PromptTokens+u.CompletionTokens {
		t.Fatalf("token mismatch: %d != %d", u.TotalTokens, u.PromptTokens+u.CompletionTokens)
	}
}
