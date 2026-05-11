package ratelimit_test

import (
	"errors"
	"testing"
	"time"

	"github.com/cronlog/internal/ratelimit"
)

// fakeRunner records how many times it was called.
type fakeRunner struct {
	calls  int
	exitCode int
	err    error
}

func (f *fakeRunner) Run() (int, error) {
	f.calls++
	return f.exitCode, f.err
}

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestNew_RejectsNilRunner(t *testing.T) {
	_, err := ratelimit.New(nil, 1, time.Minute, nil)
	if err == nil {
		t.Fatal("expected error for nil runner")
	}
}

func TestNew_RejectsZeroMax(t *testing.T) {
	_, err := ratelimit.New(&fakeRunner{}, 0, time.Minute, nil)
	if err == nil {
		t.Fatal("expected error for maxRuns=0")
	}
}

func TestNew_RejectsZeroWindow(t *testing.T) {
	_, err := ratelimit.New(&fakeRunner{}, 1, 0, nil)
	if err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestRun_AllowsUpToMax(t *testing.T) {
	base := time.Now()
	clock := fixedClock(base)
	fr := &fakeRunner{}

	lim, err := ratelimit.New(fr, 3, time.Minute, clock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < 3; i++ {
		if _, err := lim.Run(); err != nil {
			t.Fatalf("run %d: unexpected error: %v", i+1, err)
		}
	}
	if fr.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", fr.calls)
	}
}

func TestRun_BlocksWhenLimitReached(t *testing.T) {
	base := time.Now()
	clock := fixedClock(base)
	fr := &fakeRunner{}

	lim, _ := ratelimit.New(fr, 2, time.Minute, clock)
	lim.Run() //nolint
	lim.Run() //nolint

	_, err := lim.Run()
	if err == nil {
		t.Fatal("expected rate-limit error on third call")
	}
	if fr.calls != 2 {
		t.Fatalf("runner should not have been called a third time, got %d calls", fr.calls)
	}
}

func TestRun_ResetsAfterWindow(t *testing.T) {
	now := time.Now()
	current := now
	clock := func() time.Time { return current }

	fr := &fakeRunner{}
	lim, _ := ratelimit.New(fr, 1, time.Minute, clock)

	lim.Run() //nolint

	// Advance clock beyond the window.
	current = now.Add(2 * time.Minute)

	if _, err := lim.Run(); err != nil {
		t.Fatalf("expected run to succeed after window reset: %v", err)
	}
	if fr.calls != 2 {
		t.Fatalf("expected 2 calls, got %d", fr.calls)
	}
}

func TestRun_PropagatesRunnerError(t *testing.T) {
	sentinel := errors.New("boom")
	fr := &fakeRunner{exitCode: 1, err: sentinel}
	lim, _ := ratelimit.New(fr, 5, time.Minute, nil)

	code, err := lim.Run()
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
}
