package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/thanhauco/ai-harness/pkg/eval"
	"github.com/thanhauco/ai-harness/pkg/harness"
	"github.com/thanhauco/ai-harness/pkg/provider"
	"github.com/thanhauco/ai-harness/pkg/resilience"
	"github.com/thanhauco/ai-harness/pkg/storage"
)

const (
	Version   = "v1.0.0"
	BuildDate = "2026-09-04"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		handleRun(os.Args[2:])
	case "eval":
		handleEval(os.Args[2:])
	case "bench":
		handleBench(os.Args[2:])
	case "version":
		handleVersion()
	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("AI Harness (aih) - High-throughput AI execution and evaluation runtime in Go")
	fmt.Println("\nUsage:")
	fmt.Println("  aih <command> [arguments]")
	fmt.Println("\nAvailable Commands:")
	fmt.Println("  run       Execute a prompt through resilient execution pipeline")
	fmt.Println("  eval      Run automated benchmark suite with assertion rules")
	fmt.Println("  bench     Stress-test execution harness throughput and latency")
	fmt.Println("  version   Display version and Go runtime build metadata")
}

func handleVersion() {
	fmt.Printf("aih version %s (%s) %s/%s\n", Version, BuildDate, runtime.GOOS, runtime.GOARCH)
}

func handleRun(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	model := fs.String("model", "mock-model", "Model name or provider identifier")
	promptText := fs.String("prompt", "Analyze recent advances in distributed systems", "Prompt content")
	_ = fs.Parse(args)

	fmt.Printf("Running prompt on model [%s]...\n", *model)
	p := provider.NewMockProvider(*model, "Simulated production AI harness response generated at high throughput.")

	runner := harness.NewRunner(harness.RunnerConfig{
		Provider: p,
		Policy: &resilience.Policy{
			Limiter: resilience.NewTokenBucket(100, 50),
			Breaker: resilience.NewCircuitBreaker(resilience.CircuitBreakerConfig{FailureThreshold: 5}),
			Retry:   resilience.DefaultRetryConfig(),
		},
		Cache: storage.NewLRUCache(500, time.Hour),
	})

	resp, err := runner.ExecutePrompt(context.Background(), harness.NewPrompt(harness.NewUserMessage(*promptText)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Execution failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n--- Response ---")
	fmt.Println(resp.Content)
	fmt.Printf("\nMetrics: %d tokens | %dms latency | cached: %v\n",
		resp.Usage.TotalTokens, resp.Usage.DurationMs, resp.Cached)
}

func handleEval(args []string) {
	fmt.Println("Starting Evaluation Suite...")
	mockP := provider.NewMockProvider("eval-model", "Result: status=healthy, checks=passed, error=none")

	suite := eval.NewSuite(
		eval.TestCase{
			ID:     "health_check",
			Prompt: harness.NewPrompt(harness.NewUserMessage("Perform system health check")),
			Assertions: []eval.Assertion{
				&eval.ContainsAssertion{Substring: "status=healthy"},
				&eval.NotContainsAssertion{Forbidden: "error=fatal"},
			},
		},
		eval.TestCase{
			ID:     "sentiment_probe",
			Prompt: harness.NewPrompt(harness.NewUserMessage("Probe system readiness")),
			Assertions: []eval.Assertion{
				&eval.ContainsAssertion{Substring: "checks=passed"},
			},
		},
	)

	report, err := suite.Run(context.Background(), mockP, 2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Eval failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n" + report.MarkdownTable())
}

func handleBench(args []string) {
	fs := flag.NewFlagSet("bench", flag.ExitOnError)
	requests := fs.Int("n", 1000, "Total requests")
	concurrency := fs.Int("c", 16, "Concurrency")
	_ = fs.Parse(args)

	fmt.Printf("Benchmarking harness with %d requests (%d concurrency)...\n", *requests, *concurrency)
	p := provider.NewMockProvider("bench-model", "Bench output payload.")
	runner := harness.NewRunner(harness.RunnerConfig{
		Provider: p,
		Cache:    storage.NewLRUCache(1000, time.Hour),
	})

	start := time.Now()
	sem := make(chan struct{}, *concurrency)
	for i := 0; i < *requests; i++ {
		sem <- struct{}{}
		go func() {
			defer func() { <-sem }()
			_, _ = runner.ExecutePrompt(context.Background(), harness.NewPrompt(harness.NewUserMessage("bench")))
		}()
	}

	for i := 0; i < *concurrency; i++ {
		sem <- struct{}{}
	}
	elapsed := time.Since(start)

	rps := float64(*requests) / elapsed.Seconds()
	fmt.Printf("Completed %d requests in %v (%.2f req/sec)\n", *requests, elapsed, rps)
}
