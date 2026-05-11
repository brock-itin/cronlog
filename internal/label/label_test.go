package label_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cronlog/internal/label"
)

func TestWrite_PrependsPrefixToEachLine(t *testing.T) {
	var buf bytes.Buffer
	w := label.New(&buf, "job")

	_, err := w.Write([]byte("hello\nworld\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "[job] hello\n") {
		t.Errorf("expected [job] hello line, got: %q", got)
	}
	if !strings.Contains(got, "[job] world\n") {
		t.Errorf("expected [job] world line, got: %q", got)
	}
}

func TestWrite_BuffersPartialLines(t *testing.T) {
	var buf bytes.Buffer
	w := label.New(&buf, "x")

	w.Write([]byte("par"))
	if buf.Len() != 0 {
		t.Errorf("expected no output before newline, got %q", buf.String())
	}

	w.Write([]byte("tial\n"))
	if !strings.Contains(buf.String(), "[x] partial\n") {
		t.Errorf("expected prefixed line after newline, got %q", buf.String())
	}
}

func TestClose_FlushesRemainingData(t *testing.T) {
	var buf bytes.Buffer
	w := label.New(&buf, "fin")

	w.Write([]byte("no newline"))
	if err := w.Close(); err != nil {
		t.Fatalf("unexpected error on Close: %v", err)
	}

	if !strings.Contains(buf.String(), "[fin] no newline") {
		t.Errorf("expected flushed content with prefix, got %q", buf.String())
	}
}

func TestClose_NoOp_WhenBufferEmpty(t *testing.T) {
	var buf bytes.Buffer
	w := label.New(&buf, "empty")

	if err := w.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty close, got %q", buf.String())
	}
}

func TestWrite_MultipleWritesSingleLine(t *testing.T) {
	var buf bytes.Buffer
	w := label.New(&buf, "multi")

	w.Write([]byte("one "))
	w.Write([]byte("two "))
	w.Write([]byte("three\n"))

	expected := "[multi] one two three\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}
