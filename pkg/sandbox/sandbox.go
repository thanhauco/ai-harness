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
