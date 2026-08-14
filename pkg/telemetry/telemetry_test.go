package telemetry

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestMaskString(t *testing.T) {
	input := "Connecting with key sk-1234567890abcdef12345678 and email dev@example.com"
	masked := MaskString(input)

	if strings.Contains(masked, "sk-1234567890abcdef12345678") {
		t.Fatalf("expected key to be redacted, got: %s", masked)
	}
	if strings.Contains(masked, "dev@example.com") {
		t.Fatalf("expected email to be redacted, got: %s", masked)
	}
	if !strings.Contains(masked, "[REDACTED]") {
		t.Fatalf("expected [REDACTED] tag in output: %s", masked)
	}
}
