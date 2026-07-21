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

// SuiteReport contains aggregated benchmark outcomes.
type SuiteReport struct {
	TotalCases int          `json:"total_cases"`
	Passed     int          `json:"passed"`
	Failed     int          `json:"failed"`
	PassRate   float64      `json:"pass_rate"`
	TotalTime  time.Duration`json:"total_time"`
	P50Ms      int64        `json:"p50_ms"`
	P90Ms      int64        `json:"p90_ms"`
	P99Ms      int64        `json:"p99_ms"`
	TotalTokens int         `json:"total_tokens"`
	Results    []CaseResult `json:"results"`
}

type Suite struct {
	Cases []TestCase
}

func NewSuite(cases ...TestCase) *Suite {
	return &Suite{Cases: cases}
}

func (s *Suite) Run(ctx context.Context, p provider.Provider, concurrency int) (*SuiteReport, error) {
	if concurrency <= 0 {
		concurrency = 4
	}

	start := time.Now()
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	results := make([]CaseResult, 0, len(s.Cases))

	for _, tc := range s.Cases {
		wg.Add(1)
		sem <- struct{}{}

		go func(c TestCase) {
			defer wg.Done()
			defer func() { <-sem }()

			caseStart := time.Now()
			resp, err := p.Generate(ctx, c.Prompt)
			duration := time.Since(caseStart)

			cRes := CaseResult{
				CaseID:     c.ID,
				DurationMs: duration.Milliseconds(),
				Passed:     true,
			}

			if err != nil {
				cRes.Passed = false
				cRes.Assertions = append(cRes.Assertions, AssertionResult{
					Name:    "generation_error",
					Passed:  false,
					Message: err.Error(),
				})
			} else {
				cRes.Tokens = resp.Usage.TotalTokens

				// Run assertions
				for _, a := range c.Assertions {
					res := a.Check(resp)
					cRes.Assertions = append(cRes.Assertions, res)
					if !res.Passed {
						cRes.Passed = false
					}
				}

				// Run rubric if present
				if c.Rubric != nil {
					rScore := c.Rubric.Evaluate(resp.Content)
					cRes.Rubric = &rScore
					if !rScore.Passed {
						cRes.Passed = false
					}
				}
			}

			mu.Lock()
			results = append(results, cRes)
			mu.Unlock()
		}(tc)
	}

	wg.Wait()

	report := &SuiteReport{
		TotalCases: len(results),
		Results:    results,
		TotalTime:  time.Since(start),
	}

	latencies := make([]int64, len(results))
	for i, r := range results {
		if r.Passed {
			report.Passed++
		} else {
			report.Failed++
		}
		report.TotalTokens += r.Tokens
		latencies[i] = r.DurationMs
	}

	if report.TotalCases > 0 {
		report.PassRate = (float64(report.Passed) / float64(report.TotalCases)) * 100.0
	}

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	if len(latencies) > 0 {
		report.P50Ms = latencies[len(latencies)*50/100]
		report.P90Ms = latencies[len(latencies)*90/100]
		report.P99Ms = latencies[len(latencies)*99/100]
	}

	return report, nil
}

// Percentiles computes p50, p90, and p99 from a slice of latencies.
func ComputePercentiles(latencies []int64) (p50, p90, p99 int64) {
	if len(latencies) == 0 {
		return 0, 0, 0
	}
	sorted := make([]int64, len(latencies))
	copy(sorted, latencies)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	return sorted[len(sorted)*50/100], sorted[len(sorted)*90/100], sorted[len(sorted)*99/100]
}

// MarkdownTable renders the SuiteReport as a GitHub-flavored Markdown table.
func (r *SuiteReport) MarkdownTable() string {
	var sb strings.Builder
	sb.WriteString("| Case ID | Status | Duration | Tokens | Details |\n")
	sb.WriteString("|---------|--------|----------|--------|---------|\n")

	for _, c := range r.Results {
		status := "PASS"
		if !c.Passed {
			status = "FAIL"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %dms | %d | %d assertions |\n",
			c.CaseID, status, c.DurationMs, c.Tokens, len(c.Assertions)))
	}

	sb.WriteString(fmt.Sprintf("\n**Summary:** %d/%d Passed (%.1f%%) | P50: %dms | P99: %dms | Total Tokens: %d\n",
		r.Passed, r.TotalCases, r.PassRate, r.P50Ms, r.P99Ms, r.TotalTokens))

	return sb.String()
}
