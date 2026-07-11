package eval

// Criterion defines a scored grading dimension.
type Criterion struct {
	Name        string                               `json:"name"`
	Weight      float64                              `json:"weight"`
	Description string                               `json:"description"`
	Evaluator   func(response string) (float64, string) // returns score 0.0-1.0 and feedback
}

// RubricScore aggregates weighted criteria grades.
type RubricScore struct {
	OverallScore float64            `json:"overall_score"` // 0.0 - 100.0
	Passed       bool               `json:"passed"`
	Feedback     map[string]string  `json:"feedback"`
}

// Rubric holds a collection of weighted criteria.
type Rubric struct {
	Criteria      []Criterion `json:"criteria"`
	PassThreshold float64     `json:"pass_threshold"` // default 70.0
}

func NewRubric(passThreshold float64, criteria ...Criterion) *Rubric {
	if passThreshold <= 0 {
		passThreshold = 70.0
	}
	return &Rubric{
		Criteria:      criteria,
		PassThreshold: passThreshold,
	}
}
