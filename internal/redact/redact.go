// Package redact provides utilities for scrubbing sensitive values
// (e.g. passwords, tokens, API keys) from cron job output before it
// is written to log files or forwarded to notification endpoints.
package redact

import (
	"regexp"
	"strings"
)

const placeholder = "[REDACTED]"

// Redactor scrubs sensitive patterns from text.
type Redactor struct {
	literals []string
	patterns []*regexp.Regexp
}

// New returns a Redactor configured with the supplied literal strings and
// regular-expression patterns. Invalid regex patterns are silently skipped.
func New(literals []string, patterns []string) *Redactor {
	r := &Redactor{literals: literals}
	for _, p := range patterns {
		if re, err := regexp.Compile(p); err == nil {
			r.patterns = append(r.patterns, re)
		}
	}
	return r
}

// Apply returns a copy of text with all sensitive values replaced by
// [REDACTED].
func (r *Redactor) Apply(text string) string {
	for _, lit := range r.literals {
		if lit == "" {
			continue
		}
		text = strings.ReplaceAll(text, lit, placeholder)
	}
	for _, re := range r.patterns {
		text = re.ReplaceAllString(text, placeholder)
	}
	return text
}

// Lines applies Apply to every element of lines and returns the result.
func (r *Redactor) Lines(lines []string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = r.Apply(l)
	}
	return out
}
