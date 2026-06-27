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
