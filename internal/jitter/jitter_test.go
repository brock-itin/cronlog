package jitter

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeRunner records calls and returns preconfigured results.
type fakeRunner struct {
	called bool
	code   int
	err    error
}

func (f *fakeRunner) Run(_ context.Context, _ string, _ []string) (int, error) {
	f.called = true
	return f.code, f.err
}

func noSleep(_ context.Context, _ time.Duration) error { return nil }

func TestNew_RejectsNilRunner(t *testing.T) {
	_, err := New(nil, time.Second)
	if err == nil {
		t.Fatal("expected error for nil runner")
	}
}

func TestNew_RejectsNonPositiveMax(t *testing.T) {
	r := &fakeRunner{}
	for _, d := range []time.Duration{0, -time.Millisecond} {
		_, err := New(r, d)
		if err == nil {
			t.Fatalf("expected error for max=%s", d)
		}
	}
}

func TestRun_DelegatesAfterDelay(t *testing.T) {
	r := &fakeRunner{code: 0}
	j, err := New(r, time.Second)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	j.sleep = noSleep

	code, err := j.Run(context.Background(), "echo", []string{"hi"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !r.called {
		t.Fatal("inner runner was not called")
	}
}

func TestRun_PropagatesRunnerError(t *testing.T) {
	want := errors.New("boom")
	r := &fakeRunner{code: 1, err: want}
	j, _ := New(r, time.Second)
	j.sleep = noSleep

	_, err := j.Run(context.Background(), "fail", nil)
	if !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
}

func TestRun_CancelledContext_AbortsBeforeRunner(t *testing.T) {
	r := &fakeRunner{}
	j, _ := New(r, time.Second)
	// Replace sleep with one that honours cancellation.
	j.sleep = contextSleep

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	_, err := j.Run(ctx, "echo", nil)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	if r.called {
		t.Fatal("runner should not have been called after context cancellation")
	}
}
