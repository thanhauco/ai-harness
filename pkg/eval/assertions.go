package eval

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/thanhauco/ai-harness/pkg/harness"
)

// AssertionResult records whether a specific evaluation check succeeded.
type AssertionResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

// Assertion evaluates a harness response.
type Assertion interface {
	Name() string
	Check(resp *harness.Response) AssertionResult
}

type ExactMatchAssertion struct {
	Expected string
}

func (a *ExactMatchAssertion) Name() string {
	return "exact_match"
}

func (a *ExactMatchAssertion) Check(resp *harness.Response) AssertionResult {
	if resp.Content == a.Expected {
		return AssertionResult{Name: a.Name(), Passed: true, Message: "content matches exactly"}
	}
	return AssertionResult{
		Name:    a.Name(),
		Passed:  false,
		Message: fmt.Sprintf("expected %q, got %q", a.Expected, resp.Content),
	}
}

type ContainsAssertion struct {
	Substring string
}

func (a *ContainsAssertion) Name() string {
	return "contains"
}

func (a *ContainsAssertion) Check(resp *harness.Response) AssertionResult {
	if strings.Contains(resp.Content, a.Substring) {
		return AssertionResult{Name: a.Name(), Passed: true, Message: "contains required substring"}
	}
	return AssertionResult{
		Name:    a.Name(),
		Passed:  false,
		Message: fmt.Sprintf("missing required substring %q", a.Substring),
	}
}

type NotContainsAssertion struct {
	Forbidden string
}

func (a *NotContainsAssertion) Name() string {
	return "not_contains"
}

func (a *NotContainsAssertion) Check(resp *harness.Response) AssertionResult {
	if !strings.Contains(resp.Content, a.Forbidden) {
		return AssertionResult{Name: a.Name(), Passed: true, Message: "does not contain forbidden substring"}
	}
	return AssertionResult{
		Name:    a.Name(),
		Passed:  false,
		Message: fmt.Sprintf("found forbidden substring %q", a.Forbidden),
	}
}

type RegexAssertion struct {
	Pattern *regexp.Regexp
}

func NewRegexAssertion(pattern string) (*RegexAssertion, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &RegexAssertion{Pattern: re}, nil
}

func (a *RegexAssertion) Name() string {
	return "regex_match"
}

func (a *RegexAssertion) Check(resp *harness.Response) AssertionResult {
	if a.Pattern.MatchString(resp.Content) {
		return AssertionResult{Name: a.Name(), Passed: true, Message: "matches regex pattern"}
	}
	return AssertionResult{
		Name:    a.Name(),
		Passed:  false,
		Message: fmt.Sprintf("content does not match regex %q", a.Pattern.String()),
	}
}

type JSONValidAssertion struct{}

func (a *JSONValidAssertion) Name() string {
	return "json_valid"
}

func (a *JSONValidAssertion) Check(resp *harness.Response) AssertionResult {
	var js any
	if err := json.Unmarshal([]byte(resp.Content), &js); err == nil {
		return AssertionResult{Name: a.Name(), Passed: true, Message: "valid json"}
	}
	return AssertionResult{Name: a.Name(), Passed: false, Message: "invalid json format"}
}

type LatencyAssertion struct {
	MaxDurationMs int64
}

func (a *LatencyAssertion) Name() string {
	return "latency_budget"
}

func (a *LatencyAssertion) Check(resp *harness.Response) AssertionResult {
	if resp.Usage.DurationMs <= a.MaxDurationMs {
		return AssertionResult{Name: a.Name(), Passed: true, Message: "latency within budget"}
	}
	return AssertionResult{
		Name:    a.Name(),
		Passed:  false,
		Message: fmt.Sprintf("latency %dms exceeds budget %dms", resp.Usage.DurationMs, a.MaxDurationMs),
	}
}
