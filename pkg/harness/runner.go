package harness

import (
	"context"
	"fmt"
	"time"

	"github.com/thanhauco/ai-harness/pkg/resilience"
	"github.com/thanhauco/ai-harness/pkg/storage"
	"github.com/thanhauco/ai-harness/pkg/telemetry"
)

// ProviderCaller abstracts model provider invocation without circular imports.
type ProviderCaller interface {
	Name() string
	Generate(ctx context.Context, prompt *Prompt) (*Response, error)
}

// RunnerConfig configures the top-level execution harness.
type RunnerConfig struct {
	Provider   ProviderCaller
	Policy     *resilience.Policy
	Cache      *storage.LRUCache
	Metrics    *telemetry.Metrics
	EnableLogs bool
}

// Runner executes prompts with caching, rate limiting, circuit breaker, retries, and telemetry.
type Runner struct {
	cfg RunnerConfig
}

func NewRunner(cfg RunnerConfig) *Runner {
	if cfg.Metrics == nil {
		cfg.Metrics = telemetry.NewMetrics()
	}
	return &Runner{cfg: cfg}
}

func (r *Runner) ExecutePrompt(ctx context.Context, prompt *Prompt) (*Response, error) {
	ctx, span := telemetry.StartSpan(ctx, "harness.ExecutePrompt")
	defer span.End()

	var cacheKey string
	if r.cfg.Cache != nil {
		cacheKey = storage.HashKey(prompt)
		if val, ok := r.cfg.Cache.Get(cacheKey); ok {
			if cached, ok := val.(*Response); ok {
				cachedCopy := *cached
				cachedCopy.Cached = true
				span.SetAttribute("cache.hit", "true")
				return &cachedCopy, nil
			}
		}
		span.SetAttribute("cache.hit", "false")
	}

	start := time.Now()
	var resp *Response
	var execErr error

	op := func(opCtx context.Context) error {
		if r.cfg.Provider == nil {
			return fmt.Errorf("no provider configured in harness")
		}
		var err error
		resp, err = r.cfg.Provider.Generate(opCtx, prompt)
		return err
	}

	if r.cfg.Policy != nil {
		execErr = r.cfg.Policy.Execute(ctx, op)
	} else {
		execErr = op(ctx)
	}

	duration := time.Since(start)
	promptTok := 0
	complTok := 0
	if resp != nil {
		promptTok = resp.Usage.PromptTokens
		complTok = resp.Usage.CompletionTokens
	}

	r.cfg.Metrics.RecordRequest(duration, promptTok, complTok, execErr != nil)

	if execErr != nil {
		span.SetAttribute("error", execErr.Error())
		return nil, execErr
	}

	if r.cfg.Cache != nil && resp != nil {
		r.cfg.Cache.Set(cacheKey, resp)
	}

	return resp, nil
}
