package main

import (
	"context"
	"fmt"

	"github.com/thanhauco/ai-harness/pkg/eval"
	"github.com/thanhauco/ai-harness/pkg/harness"
	"github.com/thanhauco/ai-harness/pkg/provider"
)

func main() {
	p := provider.NewMockProvider("mock-eval", "SQL Query: SELECT id, name FROM users WHERE active = 1;")

	suite := eval.NewSuite(
		eval.TestCase{
			ID:     "sql_generation",
			Prompt: harness.NewPrompt(harness.NewUserMessage("Generate SQL to fetch active users")),
			Assertions: []eval.Assertion{
				&eval.ContainsAssertion{Substring: "SELECT"},
				&eval.ContainsAssertion{Substring: "FROM users"},
				&eval.NotContainsAssertion{Forbidden: "DROP TABLE"},
			},
		},
	)

	report, err := suite.Run(context.Background(), p, 2)
	if err != nil {
		panic(err)
	}

	fmt.Println(report.MarkdownTable())
}
