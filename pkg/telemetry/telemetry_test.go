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

func TestMetricsAndTracing(t *testing.T) {
	metrics := NewMetrics()
	metrics.RecordRequest(50*time.Millisecond, 10, 20, false)
	metrics.RecordRequest(100*time.Millisecond, 15, 25, true)

	snap := metrics.Snapshot()
	if snap.Requests != 2 || snap.Errors != 1 {
		t.Fatalf("metrics mismatch: %+v", snap)
	}

	// Tracing
	ctx, root := StartSpan(context.Background(), "root")
	_, child := StartSpan(ctx, "child")
	child.SetAttribute("model", "gpt-4")
	child.End()
	root.End()

	if child.TraceID != root.TraceID {
		t.Fatalf("child trace ID %s != root trace ID %s", child.TraceID, root.TraceID)
	}
	if child.ParentID != root.SpanID {
		t.Fatalf("child parent ID %s != root span ID %s", child.ParentID, root.SpanID)
	}
}
