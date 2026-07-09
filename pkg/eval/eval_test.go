package eval

import (
	"context"
	"strings"
	"testing"

	"github.com/thanhauco/ai-harness/pkg/harness"
	"github.com/thanhauco/ai-harness/pkg/provider"
)

func TestAssertions(t *testing.T) {
	resp := &harness.Response{
		Content: `{"status":"ok"}`,
		Usage: harness.TokenUsage{
			DurationMs:  120,
			TotalTokens: 45,
		},
	}

	contains := &ContainsAssertion{Substring: "status"}
	res1 := contains.Check(resp)
	if !res1.Passed {
		t.Fatalf("expected contains to pass")
	}

	notContains := &NotContainsAssertion{Forbidden: "error"}
	res2 := notContains.Check(resp)
	if !res2.Passed {
		t.Fatalf("expected not_contains to pass")
	}

	jsonValid := &JSONValidAssertion{}
	res3 := jsonValid.Check(resp)
	if !res3.Passed {
		t.Fatalf("expected json_valid to pass")
	}

	latency := &LatencyAssertion{MaxDurationMs: 200}
	res4 := latency.Check(resp)
	if !res4.Passed {
		t.Fatalf("expected latency within 200ms to pass")
	}
}
