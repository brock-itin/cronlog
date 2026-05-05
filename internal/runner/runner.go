// Package runner executes shell commands and captures their output,
// writing structured log entries via the rotator.
package runner

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/user/cronlog/internal/rotator"
)

// Result holds the outcome of a command execution.
type Result struct {
	Command  string
	Args     []string
	ExitCode int
	Duration time.Duration
	Started  time.Time
}

// Runner wraps a rotator and executes commands, logging their output.
type Runner struct {
	rot *rotator.Rotator
}

// New creates a Runner that writes output to the given rotator.
func New(rot *rotator.Rotator) *Runner {
	return &Runner{rot: rot}
}

// Run executes the named command with the provided arguments.
// Combined stdout and stderr are written to the rotator as they arrive.
// It returns a Result describing the execution outcome.
func (r *Runner) Run(command string, args ...string) (Result, error) {
	start := time.Now()
	cmd := exec.Command(command, args...)
	cmd.Stdout = r.rot
	cmd.Stderr = r.rot

	header := fmt.Sprintf("[cronlog] start command=%q args=%v time=%s\n",
		command, args, start.Format(time.RFC3339))
	if _, err := r.rot.Write([]byte(header)); err != nil {
		return Result{}, fmt.Errorf("runner: write header: %w", err)
	}

	runErr := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return Result{}, fmt.Errorf("runner: exec: %w", runErr)
		}
	}

	footer := fmt.Sprintf("[cronlog] done  command=%q exit_code=%d duration=%s\n",
		command, exitCode, duration.Round(time.Millisecond))
	if _, err := r.rot.Write([]byte(footer)); err != nil {
		return Result{}, fmt.Errorf("runner: write footer: %w", err)
	}

	return Result{
		Command:  command,
		Args:     args,
		ExitCode: exitCode,
		Duration: duration,
		Started:  start,
	}, nil
}
