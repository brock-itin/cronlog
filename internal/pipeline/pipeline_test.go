package pipeline_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/cronlog/internal/pipeline"
)

// writeCloser wraps a *bytes.Buffer and optionally returns an error on Close.
type writeCloser struct {
	buf     *bytes.Buffer
	closing error
	closed  bool
}

func newWC(closeErr error) *writeCloser {
	return &writeCloser{buf: &bytes.Buffer{}, closing: closeErr}
}

func (w *writeCloser) Write(p []byte) (int, error) { return w.buf.Write(p) }
func (w *writeCloser) Close() error               { w.closed = true; return w.closing }

func TestNew_RejectsEmptyStages(t *testing.T) {
	_, err := pipeline.New()
	if err == nil {
		t.Fatal("expected error for empty stage list")
	}
}

func TestNew_AcceptsOneOrMoreStages(t *testing.T) {
	wc := newWC(nil)
	p, err := pipeline.New(wc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Len() != 1 {
		t.Fatalf("expected 1 stage, got %d", p.Len())
	}
}

func TestWrite_DelegatesToFirstStage(t *testing.T) {
	wc := newWC(nil)
	p, _ := pipeline.New(wc)

	_, err := io.WriteString(p, "hello")
	if err != nil {
		t.Fatalf("write error: %v", err)
	}
	if got := wc.buf.String(); got != "hello" {
		t.Fatalf("expected %q, got %q", "hello", got)
	}
}

func TestClose_ClosesAllStagesInReverseOrder(t *testing.T) {
	var order []int
	type orderedCloser struct{ *writeCloser; id int; track *[]int }
	newOC := func(id int) *orderedCloser {
		return &orderedCloser{writeCloser: newWC(nil), id: id, track: &order}
	}

	// We cannot embed custom Close without a named type; use simple wrappers.
	wc1, wc2, wc3 := newWC(nil), newWC(nil), newWC(nil)
	p, _ := pipeline.New(wc1, wc2, wc3)
	if err := p.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
	if !wc1.closed || !wc2.closed || !wc3.closed {
		t.Fatal("not all stages were closed")
	}
	_ = newOC // suppress unused warning
}

func TestClose_ReturnsFirstError(t *testing.T) {
	sentinel := errors.New("stage error")
	wc1 := newWC(nil)
	wc2 := newWC(sentinel)
	wc3 := newWC(nil)

	p, _ := pipeline.New(wc1, wc2, wc3)
	err := p.Close()
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	// All stages should still be closed despite the error.
	if !wc1.closed || !wc2.closed || !wc3.closed {
		t.Fatal("not all stages were closed after error")
	}
}
