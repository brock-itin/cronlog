package filter_test

import (
	"testing"

	"github.com/yourorg/cronlog/internal/filter"
)

func TestSuppress_ExactMatch(t *testing.T) {
	f := filter.New(nil, []string{"no output", "nothing to do"})

	if !f.Suppress("no output") {
		t.Error("expected exact match to be suppressed")
	}
	if f.Suppress("some output") {
		t.Error("expected non-matching line to be kept")
	}
}

func TestSuppress_RegexMatch(t *testing.T) {
	f := filter.New([]string{`^DEBUG:`, `\bskipping\b`}, nil)

	if !f.Suppress("DEBUG: verbose message") {
		t.Error("expected regex match to be suppressed")
	}
	if !f.Suppress("INFO: skipping stale entry") {
		t.Error("expected word-boundary regex to match")
	}
	if f.Suppress("INFO: processing entry") {
		t.Error("expected non-matching line to be kept")
	}
}

func TestSuppress_InvalidRegexSkipped(t *testing.T) {
	// Invalid pattern should not panic; valid one still works.
	f := filter.New([]string{`[invalid`, `^WARN:`}, nil)

	if !f.Suppress("WARN: low disk") {
		t.Error("expected valid pattern to still suppress")
	}
}

func TestApply_FiltersLines(t *testing.T) {
	f := filter.New([]string{`^DEBUG:`}, []string{"nothing to do"})

	input := []string{
		"INFO: job started",
		"DEBUG: internal state",
		"nothing to do",
		"INFO: job finished",
	}

	got := f.Apply(input)

	if len(got) != 2 {
		t.Fatalf("expected 2 lines after filter, got %d", len(got))
	}
	if got[0] != "INFO: job started" || got[1] != "INFO: job finished" {
		t.Errorf("unexpected filtered output: %v", got)
	}
}

func TestApply_EmptyFilter_KeepsAll(t *testing.T) {
	f := filter.New(nil, nil)

	input := []string{"line one", "line two"}
	got := f.Apply(input)

	if len(got) != len(input) {
		t.Errorf("expected all lines kept, got %d", len(got))
	}
}
