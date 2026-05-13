package stamp

import (
	"strings"
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

var epoch = time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

func TestWrite_PrependTimestampToLine(t *testing.T) {
	var out strings.Builder
	w := New(&out, time.RFC3339)
	w.clock = fixedClock(epoch)

	w.Write([]byte("hello world\n"))

	got := out.String()
	want := "2024-01-15T10:00:00Z hello world\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWrite_MultipleLines(t *testing.T) {
	var out strings.Builder
	w := New(&out, time.RFC3339)
	w.clock = fixedClock(epoch)

	w.Write([]byte("line one\nline two\n"))

	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	for _, l := range lines {
		if !strings.HasPrefix(l, "2024-01-15T10:00:00Z ") {
			t.Errorf("line missing timestamp prefix: %q", l)
		}
	}
}

func TestWrite_BuffersPartialLine(t *testing.T) {
	var out strings.Builder
	w := New(&out, time.RFC3339)
	w.clock = fixedClock(epoch)

	w.Write([]byte("partial"))
	if out.Len() != 0 {
		t.Errorf("expected no output for partial line, got %q", out.String())
	}

	w.Write([]byte(" line\n"))
	if !strings.Contains(out.String(), "partial line") {
		t.Errorf("expected flushed line to contain 'partial line', got %q", out.String())
	}
}

func TestClose_FlushesRemainingData(t *testing.T) {
	var out strings.Builder
	w := New(&out, time.RFC3339)
	w.clock = fixedClock(epoch)

	w.Write([]byte("no newline"))
	if err := w.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	if !strings.Contains(out.String(), "no newline") {
		t.Errorf("expected flushed content, got %q", out.String())
	}
}

func TestClose_NoOp_WhenBufferEmpty(t *testing.T) {
	var out strings.Builder
	w := New(&out, time.RFC3339)
	w.clock = fixedClock(epoch)

	w.Write([]byte("complete\n"))
	pre := out.String()

	if err := w.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	if out.String() != pre {
		t.Errorf("Close wrote extra data: %q", out.String()[len(pre):])
	}
}

func TestNew_DefaultFormat(t *testing.T) {
	var out strings.Builder
	w := New(&out, "")
	if w.format != time.RFC3339 {
		t.Errorf("expected RFC3339 default, got %q", w.format)
	}
}
