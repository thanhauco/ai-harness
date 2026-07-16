package eval

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/thanhauco/ai-harness/pkg/harness"
	"github.com/thanhauco/ai-harness/pkg/provider"
)

// TestCase represents an individual evaluation benchmark item.
type TestCase struct {
	ID         string
	Prompt     *harness.Prompt
	Assertions []Assertion
	Rubric     *Rubric
}

// CaseResult captures execution results for a single test case.
type CaseResult struct {
	CaseID     string            `json:"case_id"`
	Passed     bool              `json:"passed"`
	DurationMs int64             `json:"duration_ms"`
	Tokens     int               `json:"tokens"`
	Assertions []AssertionResult `json:"assertions"`
	Rubric     *RubricScore      `json:"rubric,omitempty"`
}
