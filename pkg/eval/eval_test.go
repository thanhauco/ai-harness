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

func TestRubric_Evaluate(t *testing.T) {
	rubric := NewRubric(75.0,
		Criterion{
			Name:   "has_solution",
			Weight: 2.0,
			Evaluator: func(resp string) (float64, string) {
				return 1.0, "solution provided"
			},
		},
		Criterion{
			Name:   "conciseness",
			Weight: 1.0,
			Evaluator: func(resp string) (float64, string) {
				return 0.5, "moderately concise"
			},
		},
	)

	score := rubric.Evaluate("some response")
	// (2.0*1.0 + 1.0*0.5) / 3.0 = 2.5 / 3.0 = 83.33%
	if score.OverallScore < 83.0 || score.OverallScore > 84.0 {
		t.Fatalf("expected ~83.33%%, got %f", score.OverallScore)
	}
	if !score.Passed {
		t.Fatal("expected rubric pass")
	}
}
