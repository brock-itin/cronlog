package stagger_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cronlog/cronlog/internal/stagger"
)

// fakeRunner records whether it was called and returns a preset exit code.
type fakeRunner struct {
	called  bool
	exitCode int
	err     error
}

func (f *fakeRunner) Run(_ context.Context, _ string, _ []string, _ []string) (int, error) {
	f.called = true
	return f.exitCode, f.err
}

// noSleep is a sleep function that returns immediately, making tests fast.
func noSleep(_ context.Context, _ time.Duration) error { return nil }

func TestNew_RejectsNilRunner(t *testing.T) {
	_, err := stagger.New(nil, time.Second, noSleep)
	if err == nil {
		t.Fatal("expected error for nil runner")
	}
}

func TestNew_RejectsNonPositiveMax(t *testing.T) {
	r := &fakeRunner{}
	for _, d := range []time.Duration{0, -time.Second} {
		_, err := stagger.New(r, d, noSleep)
		if err == nil {
			t.Fatalf("expected error for max=%s", d)
		}
	}
}

func TestRun_DelegatesAfterDelay(t *testing.T) {
	r := &fakeRunner{exitCode: 0}
	s, err := stagger.New(r, time.Minute, noSleep)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	code, err := s.Run(context.Background(), "echo", []string{"hi"}, nil)
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !r.called {
		t.Fatal("underlying runner was not called")
	}
}

func TestRun_PropagatesRunnerError(t *testing.T) {
	want := errors.New("boom")
	r := &fakeRunner{exitCode: 1, err: want}
	s, err := stagger.New(r, time.Minute, noSleep)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, got := s.Run(context.Background(), "bad", nil, nil)
	if !errors.Is(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestRun_CancelledContext_AbortsBeforeRunner(t *testing.T) {
	r := &fakeRunner{}
	// Use a real sleep so cancellation can be observed.
	s, err := stagger.New(r, time.Hour, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err = s.Run(ctx, "echo", nil, nil)
	if err == nil {
		t.Fatal("expected error due to cancelled context")
	}
	if r.called {
		t.Fatal("runner should not have been called when context is cancelled")
	}
}
