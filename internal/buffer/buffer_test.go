package buffer

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

func TestWrite_AccumulatesData(t *testing.T) {
	b := New(0)
	b.Write([]byte("hello "))
	b.Write([]byte("world"))

	if got := string(b.Bytes()); got != "hello world" {
		t.Fatalf("expected %q, got %q", "hello world", got)
	}
}

func TestWrite_EnforcesLimit(t *testing.T) {
	b := New(10)
	b.Write([]byte("12345"))
	b.Write([]byte("67890"))
	n, err := b.Write([]byte("overflow"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len("overflow") {
		t.Fatalf("expected n=%d, got %d", len("overflow"), n)
	}
	if b.Len() != 10 {
		t.Fatalf("expected len=10, got %d", b.Len())
	}
}

func TestWrite_PartialFill_UnderLimit(t *testing.T) {
	b := New(8)
	b.Write([]byte("hello")) // 5 bytes
	b.Write([]byte("world")) // would be 10, only 3 fit

	if got := string(b.Bytes()); got != "hellowor" {
		t.Fatalf("expected %q, got %q", "hellowor", got)
	}
}

func TestFlush_WritesToDestination(t *testing.T) {
	b := New(0)
	b.Write([]byte("flush me"))

	var dst bytes.Buffer
	if err := b.Flush(&dst); err != nil {
		t.Fatalf("flush error: %v", err)
	}
	if dst.String() != "flush me" {
		t.Fatalf("expected %q, got %q", "flush me", dst.String())
	}
	if b.Len() != 0 {
		t.Fatalf("expected buffer to be empty after flush")
	}
}

func TestFlush_EmptyBuffer_IsNoop(t *testing.T) {
	b := New(0)
	var dst bytes.Buffer
	if err := b.Flush(&dst); err != nil {
		t.Fatalf("unexpected error on empty flush: %v", err)
	}
	if dst.Len() != 0 {
		t.Fatal("expected nothing written to dst")
	}
}

func TestReset_ClearsBuffer(t *testing.T) {
	b := New(0)
	b.Write([]byte("data"))
	b.Reset()
	if b.Len() != 0 {
		t.Fatal("expected buffer to be empty after reset")
	}
}

func TestWrite_ConcurrentSafety(t *testing.T) {
	b := New(0)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Write([]byte("x"))
		}()
	}
	wg.Wait()
	if b.Len() != 50 {
		t.Fatalf("expected 50 bytes, got %d", b.Len())
	}
	_ = strings.Repeat("x", 50) // just to use the import
}
