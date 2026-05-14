package prefix_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cronlog/internal/prefix"
)

func TestWrite_PrependsToEachLine(t *testing.T) {
	var buf bytes.Buffer
	w := prefix.New(&buf, "[out] ")

	_, err := w.Write([]byte("hello\nworld\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "[out] hello\n") {
		t.Errorf("expected prefix on first line, got: %q", got)
	}
	if !strings.Contains(got, "[out] world\n") {
		t.Errorf("expected prefix on second line, got: %q", got)
	}
}

func TestWrite_BuffersPartialLine(t *testing.T) {
	var buf bytes.Buffer
	w := prefix.New(&buf, ">> ")

	w.Write([]byte("partial"))
	if buf.Len() != 0 {
		t.Errorf("expected nothing flushed yet, got %q", buf.String())
	}

	w.Write([]byte(" line\n"))
	got := buf.String()
	if got != ">> partial line\n" {
		t.Errorf("unexpected output: %q", got)
	}
}

func TestClose_FlushesRemainingData(t *testing.T) {
	var buf bytes.Buffer
	w := prefix.New(&buf, "[x] ")

	w.Write([]byte("no newline"))
	if err := w.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	got := buf.String()
	if got != "[x] no newline\n" {
		t.Errorf("unexpected output: %q", got)
	}
}

func TestClose_NoOp_WhenBufferEmpty(t *testing.T) {
	var buf bytes.Buffer
	w := prefix.New(&buf, "[x] ")

	if err := w.Close(); err != nil {
		t.Fatalf("unexpected error on empty close: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestWrite_MultipleWritesSingleLine(t *testing.T) {
	var buf bytes.Buffer
	w := prefix.New(&buf, "-- ")

	for _, chunk := range []string{"a", "b", "c", "\n"} {
		w.Write([]byte(chunk))
	}

	got := buf.String()
	if got != "-- abc\n" {
		t.Errorf("unexpected output: %q", got)
	}
}

func TestWrite_EmptyPrefix(t *testing.T) {
	var buf bytes.Buffer
	w := prefix.New(&buf, "")

	w.Write([]byte("line\n"))
	got := buf.String()
	if got != "line\n" {
		t.Errorf("unexpected output: %q", got)
	}
}
