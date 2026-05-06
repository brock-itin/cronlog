// Package filter provides output filtering for cron job logs,
// allowing suppression of noisy or irrelevant lines before they
// are written to the structured log.
package filter

import (
	"regexp"
	"strings"
)

// Filter holds compiled patterns used to suppress log lines.
type Filter struct {
	patterns []*regexp.Regexp
	exact    []string
}

// New creates a Filter from a list of regex patterns and exact strings.
// Patterns that fail to compile are silently skipped.
func New(patterns []string, exact []string) *Filter {
	f := &Filter{
		exact: exact,
	}
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			continue
		}
		f.patterns = append(f.patterns, re)
	}
	return f
}

// Suppress returns true if the given line should be suppressed (not logged).
func (f *Filter) Suppress(line string) bool {
	for _, e := range f.exact {
		if strings.TrimRight(line, "\n") == e {
			return true
		}
	}
	for _, re := range f.patterns {
		if re.MatchString(line) {
			return true
		}
	}
	return false
}

// Apply filters a slice of lines, returning only those that should be kept.
func (f *Filter) Apply(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		if !f.Suppress(l) {
			out = append(out, l)
		}
	}
	return out
}
