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

func (r *Rubric) Evaluate(response string) RubricScore {
	var totalWeight float64
	var weightedSum float64
	feedback := make(map[string]string)

	for _, c := range r.Criteria {
		w := c.Weight
		if w <= 0 {
			w = 1.0
		}
		totalWeight += w

		score, fb := c.Evaluator(response)
		if score < 0 {
			score = 0
		}
		if score > 1.0 {
			score = 1.0
		}

		weightedSum += score * w
		feedback[c.Name] = fb
	}

	overall := 0.0
	if totalWeight > 0 {
		overall = (weightedSum / totalWeight) * 100.0
	}

	return RubricScore{
		OverallScore: overall,
		Passed:       overall >= r.PassThreshold,
		Feedback:     feedback,
	}
}
