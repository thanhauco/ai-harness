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

func TestSuite_Run(t *testing.T) {
	mockP := provider.NewMockProvider("bench-mock", "expected output answer")
	cases := []TestCase{
		{
			ID:     "case_1",
			Prompt: harness.NewPrompt(harness.NewUserMessage("test")),
			Assertions: []Assertion{
				&ContainsAssertion{Substring: "expected output"},
			},
		},
	}

	suite := NewSuite(cases...)
	report, err := suite.Run(context.Background(), mockP, 2)
	if err != nil {
		t.Fatalf("unexpected benchmark suite error: %v", err)
	}

	if report.TotalCases != 1 || report.Passed != 1 {
		t.Fatalf("report mismatch: %+v", report)
	}

	md := report.MarkdownTable()
	if !strings.Contains(md, "case_1") {
		t.Fatalf("markdown table missing case_1")
	}
}

func TestDetectRegression(t *testing.T) {
	regressed, diff := DetectRegression(95.0, 85.0, 2.0)
	if !regressed || diff != -10.0 {
		t.Fatalf("expected regression detected: %v, diff=%f", regressed, diff)
	}

	noRegression, _ := DetectRegression(90.0, 89.5, 1.0)
	if noRegression {
		t.Fatal("did not expect regression within tolerance")
	}
}
