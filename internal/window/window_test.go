package window_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/cronlog/internal/history"
	"github.com/example/cronlog/internal/window"
)

// --- fakes ---

type fakeRunner struct {
	calls int
	code  int
	err   error
}

func (f *fakeRunner) Run(_ context.Context, _ []string) (int, error) {
	f.calls++
	return f.code, f.err
}

type fakeRecorder struct {
	entries []history.Entry
	err     error
}

func (f *fakeRecorder) Add(_ history.Entry) error        { return nil }
func (f *fakeRecorder) Entries() ([]history.Entry, error) { return f.entries, f.err }

// --- helpers ---

func entry(exitCode int, ago time.Duration) history.Entry {
	return history.Entry{ExitCode: exitCode, StartedAt: time.Now().Add(-ago)}
}

// --- tests ---

func TestNew_RejectsNilRunner(t *testing.T) {
	_, err := window.New(nil, &fakeRecorder{}, time.Hour)
	if err == nil {
		t.Fatal("expected error for nil runner")
	}
}

func TestNew_RejectsNilRecorder(t *testing.T) {
	_, err := window.New(&fakeRunner{}, nil, time.Hour)
	if err == nil {
		t.Fatal("expected error for nil recorder")
	}
}

func TestNew_RejectsNonPositiveWindow(t *testing.T) {
	_, err := window.New(&fakeRunner{}, &fakeRecorder{}, 0)
	if err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestRun_NoHistory_ExecutesJob(t *testing.T) {
	r := &fakeRunner{code: 0}
	w, _ := window.New(r, &fakeRecorder{}, time.Hour)

	code, err := w.Run(context.Background(), nil)
	if err != nil || code != 0 {
		t.Fatalf("unexpected result: code=%d err=%v", code, err)
	}
	if r.calls != 1 {
		t.Fatalf("expected 1 call, got %d", r.calls)
	}
}

func TestRun_RecentSuccess_Suppressed(t *testing.T) {
	r := &fakeRunner{code: 0}
	rec := &fakeRecorder{entries: []history.Entry{entry(0, 5*time.Minute)}}
	w, _ := window.New(r, rec, time.Hour)

	code, err := w.Run(context.Background(), nil)
	if err != nil || code != 0 {
		t.Fatalf("unexpected result: code=%d err=%v", code, err)
	}
	if r.calls != 0 {
		t.Fatalf("expected job to be suppressed, got %d calls", r.calls)
	}
}

func TestRun_RecentFailure_ExecutesJob(t *testing.T) {
	r := &fakeRunner{code: 1}
	rec := &fakeRecorder{entries: []history.Entry{entry(1, 5*time.Minute)}}
	w, _ := window.New(r, rec, time.Hour)

	w.Run(context.Background(), nil) //nolint:errcheck
	if r.calls != 1 {
		t.Fatalf("expected 1 call after recent failure, got %d", r.calls)
	}
}

func TestRun_OldSuccess_ExecutesJob(t *testing.T) {
	r := &fakeRunner{code: 0}
	rec := &fakeRecorder{entries: []history.Entry{entry(0, 2*time.Hour)}}
	w, _ := window.New(r, rec, time.Hour)

	w.Run(context.Background(), nil) //nolint:errcheck
	if r.calls != 1 {
		t.Fatalf("expected 1 call for expired window entry, got %d", r.calls)
	}
}

func TestRun_RecorderError_ReturnsError(t *testing.T) {
	rec := &fakeRecorder{err: errors.New("disk error")}
	w, _ := window.New(&fakeRunner{}, rec, time.Hour)

	_, err := w.Run(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error when recorder fails")
	}
}
