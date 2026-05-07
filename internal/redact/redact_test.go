package redact_test

import (
	"testing"

	"github.com/yourorg/cronlog/internal/redact"
)

func TestApply_ReplacesLiteral(t *testing.T) {
	r := redact.New([]string{"s3cr3t", "mypassword"}, nil)

	got := r.Apply("connecting with password=s3cr3t to host")
	want := "connecting with password=[REDACTED] to host"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestApply_ReplacesMultipleLiterals(t *testing.T) {
	r := redact.New([]string{"tok1", "tok2"}, nil)

	got := r.Apply("tok1 and tok2 present")
	want := "[REDACTED] and [REDACTED] present"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestApply_ReplacesRegex(t *testing.T) {
	r := redact.New(nil, []string{`(?i)password=\S+`})

	got := r.Apply("login Password=hunter2 ok")
	want := "login [REDACTED] ok"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestApply_InvalidRegexSkipped(t *testing.T) {
	// Should not panic; invalid pattern is ignored.
	r := redact.New(nil, []string{`[invalid`})
	got := r.Apply("some output")
	if got != "some output" {
		t.Errorf("unexpected modification: %q", got)
	}
}

func TestApply_EmptyLiteralSkipped(t *testing.T) {
	r := redact.New([]string{""}, nil)
	got := r.Apply("hello world")
	if got != "hello world" {
		t.Errorf("unexpected modification: %q", got)
	}
}

func TestLines_RedactsEachLine(t *testing.T) {
	r := redact.New([]string{"secret"}, nil)
	input := []string{"ok line", "bad secret here", "another secret"}
	got := r.Lines(input)

	expected := []string{"ok line", "bad [REDACTED] here", "another [REDACTED]"}
	for i, want := range expected {
		if got[i] != want {
			t.Errorf("line %d: got %q, want %q", i, got[i], want)
		}
	}
}

func TestLines_EmptySlice(t *testing.T) {
	r := redact.New([]string{"x"}, nil)
	got := r.Lines([]string{})
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}
