// Package env provides utilities for injecting and redacting environment
// variables passed to cron job subprocesses.
package env

import (
	"fmt"
	"os"
	"strings"
)

// Resolver merges extra environment variables with the current process
// environment and optionally masks secrets before they reach the runner.
type Resolver struct {
	base    []string
	extras  map[string]string
	masked  []string
}

// New returns a Resolver seeded with the current process environment.
func New(extras map[string]string, masked []string) *Resolver {
	return &Resolver{
		base:   os.Environ(),
		extras:  extras,
		masked: masked,
	}
}

// Resolve returns the final environment slice suitable for exec.Cmd.Env.
// Extra key=value pairs override any inherited values with the same key.
// Keys listed in masked are set to an empty string.
func (r *Resolver) Resolve() []string {
	env := make(map[string]string, len(r.base)+len(r.extras))

	for _, entry := range r.base {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}

	for k, v := range r.extras {
		env[k] = v
	}

	maskedSet := make(map[string]struct{}, len(r.masked))
	for _, k := range r.masked {
		maskedSet[k] = struct{}{}
	}

	out := make([]string, 0, len(env))
	for k, v := range env {
		if _, ok := maskedSet[k]; ok {
			v = ""
		}
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	return out
}
