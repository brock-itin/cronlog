package truncate_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cronlog/cronlog/internal/truncate"
)

func TestWrite_BelowLimit_PassesThrough(t *testing.T) {
	var buf bytes.Buffer
	tr := truncate.New(&buf, 100)

	payload := "hello world"
	n, err := tr.Write([]byte(payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(payload) {
		t.Fatalf("expected n=%d, got %d", len(payload), n)
	}
	if buf.String() != payload {
		t.Fatalf("expected %q, got %q", payload, buf.String())
	}
	if tr.Tripped() {
		t.Fatal("expected Tripped()=false")
	}
}

func TestWrite_ExceedsLimit_Truncates(t *testing.T) {
	var buf bytes.Buffer
	tr := truncate.New(&buf, 5)

	n, err := tr.Write([]byte("hello world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Caller sees full length even though only 5 bytes were forwarded.
	if n != len("hello world") {
		t.Fatalf("expected n=11, got %d", n)
	}
	if buf.String() != "hello" {
		t.Fatalf("expected %q, got %q", "hello", buf.String())
	}
	if !tr.Tripped() {
		t.Fatal("expected Tripped()=true")
	}
	if tr.Written() != 5 {
		t.Fatalf("expected Written()=5, got %d", tr.Written())
	}
}

func TestWrite_AfterLimit_DiscardsSilently(t *testing.T) {
	var buf bytes.Buffer
	tr := truncate.New(&buf, 3)

	tr.Write([]byte("abc")) // fills limit
	n, err := tr.Write([]byte("xyz"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Fatalf("expected n=3, got %d", n)
	}
	if buf.String() != "abc" {
		t.Fatalf("expected only %q in buf, got %q", "abc", buf.String())
	}
}

func TestClose_AppendsTruncationNotice(t *testing.T) {
	var buf bytes.Buffer
	tr := truncate.New(&buf, 5)
	tr.Write([]byte("hello world"))

	if err := tr.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
	if !strings.Contains(buf.String(), "truncated after 5 bytes") {
		t.Fatalf("expected truncation notice in output, got: %q", buf.String())
	}
}

func TestClose_NoNotice_WhenNotTripped(t *testing.T) {
	var buf bytes.Buffer
	tr := truncate.New(&buf, 100)
	tr.Write([]byte("small"))
	tr.Close()

	if strings.Contains(buf.String(), "truncated") {
		t.Fatalf("unexpected truncation notice: %q", buf.String())
	}
}

func TestWrite_ZeroLimit_NoTruncation(t *testing.T) {
	var buf bytes.Buffer
	tr := truncate.New(&buf, 0)

	payload := strings.Repeat("x", 10_000)
	tr.Write([]byte(payload))

	if tr.Tripped() {
		t.Fatal("zero limit should disable truncation")
	}
	if buf.Len() != 10_000 {
		t.Fatalf("expected 10000 bytes, got %d", buf.Len())
	}
}
