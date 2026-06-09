package sandbox

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

// ProcessResult captures the output of an isolated subprocess run.
type ProcessResult struct {
	Stdout   string        `json:"stdout"`
	Stderr   string        `json:"stderr"`
	ExitCode int           `json:"exit_code"`
	Duration time.Duration `json:"duration"`
}

// SandboxPolicy enforces execution limits and environment isolation.
type SandboxPolicy struct {
	Timeout         time.Duration
	MaxOutputBytes  int
	AllowedCommands []string
	EnvWhitelist    []string
}

func DefaultSandboxPolicy() SandboxPolicy {
	return SandboxPolicy{
		Timeout:        5 * time.Second,
		MaxOutputBytes: 64 * 1024, // 64 KB
	}
}

// Runner executes subprocess commands under sandbox policy constraints.
type Runner struct {
	policy SandboxPolicy
}

func NewRunner(policy SandboxPolicy) *Runner {
	if policy.Timeout <= 0 {
		policy.Timeout = 5 * time.Second
	}
	if policy.MaxOutputBytes <= 0 {
		policy.MaxOutputBytes = 64 * 1024
	}
	return &Runner{policy: policy}
}

func (r *Runner) ExecuteCommand(ctx context.Context, cmdName string, args ...string) (*ProcessResult, error) {
	if len(r.policy.AllowedCommands) > 0 {
		allowed := false
		for _, ac := range r.policy.AllowedCommands {
			if ac == cmdName {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, fmt.Errorf("command %q is disallowed by sandbox policy", cmdName)
		}
	}

	execCtx, cancel := context.WithTimeout(ctx, r.policy.Timeout)
	defer cancel()

	start := time.Now()
	cmd := exec.CommandContext(execCtx, cmdName, args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	duration := time.Since(start)

	stdoutStr := stdoutBuf.String()
	stderrStr := stderrBuf.String()

	if len(stdoutStr) > r.policy.MaxOutputBytes {
		stdoutStr = stdoutStr[:r.policy.MaxOutputBytes] + "... [output truncated]"
	}
	if len(stderrStr) > r.policy.MaxOutputBytes {
		stderrStr = stderrStr[:r.policy.MaxOutputBytes] + "... [output truncated]"
	}

	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return &ProcessResult{
		Stdout:   stdoutStr,
		Stderr:   stderrStr,
		ExitCode: exitCode,
		Duration: duration,
	}, err
}
