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
