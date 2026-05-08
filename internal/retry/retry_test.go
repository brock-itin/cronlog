package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func noSleep(_ time.Duration) {}

func newTestRunner(maxAttempts int, delay time.Duration) *Runner {
	r := New(Policy{MaxAttempts: maxAttempts, Delay: delay, Backoff: 1.0})
	r.sleep = noSleep
	return r
}

func TestRun_SucceedsOnFirstAttempt(t *testing.T) {
	r := newTestRunner(3, 0)
	res := r.Run(context.Background(), func() error { return nil })
	if res.Attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", res.Attempts)
	}
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
}

func TestRun_RetriesUntilSuccess(t *testing.T) {
	calls := 0
	r := newTestRunner(5, 0)
	res := r.Run(context.Background(), func() error {
		calls++
		if calls < 3 {
			return errors.New("not yet")
		}
		return nil
	})
	if res.Attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", res.Attempts)
	}
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
}

func TestRun_ExhaustsMaxAttempts(t *testing.T) {
	sentinel := errors.New("always fails")
	r := newTestRunner(3, 0)
	res := r.Run(context.Background(), func() error { return sentinel })
	if res.Attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", res.Attempts)
	}
	if !errors.Is(res.Err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", res.Err)
	}
}

func TestRun_RespectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0
	r := newTestRunner(5, 0)
	res := r.Run(ctx, func() error {
		calls++
		return errors.New("fail")
	})
	if calls != 0 {
		t.Fatalf("expected 0 calls after cancelled context, got %d", calls)
	}
	if res.Err == nil {
		t.Fatal("expected error due to cancelled context")
	}
}

func TestRun_BackoffIncreasesDelay(t *testing.T) {
	var delays []time.Duration
	r := New(Policy{MaxAttempts: 4, Delay: 10 * time.Millisecond, Backoff: 2.0})
	r.sleep = func(d time.Duration) { delays = append(delays, d) }

	r.Run(context.Background(), func() error { return errors.New("fail") })

	// 3 sleeps for 4 attempts
	if len(delays) != 3 {
		t.Fatalf("expected 3 sleep calls, got %d", len(delays))
	}
	if delays[1] != 2*delays[0] {
		t.Fatalf("expected delay to double: %v -> %v", delays[0], delays[1])
	}
}
