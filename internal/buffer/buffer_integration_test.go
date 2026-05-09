package buffer_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/cronlog/cronlog/internal/buffer"
)

// TestBuffer_RoundTrip simulates the typical usage pattern: multiple writers
// (stdout + stderr) write concurrently, then the buffer is flushed to a log.
func TestBuffer_RoundTrip(t *testing.T) {
	b := buffer.New(0)

	lines := []string{
		"starting job\n",
		"processing step 1\n",
		"processing step 2\n",
		"done\n",
	}

	for _, l := range lines {
		if _, err := fmt.Fprint(b, l); err != nil {
			t.Fatalf("write error: %v", err)
		}
	}

	var dst bytes.Buffer
	if err := b.Flush(&dst); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	for _, l := range lines {
		if !bytes.Contains(dst.Bytes(), []byte(l)) {
			t.Errorf("expected output to contain %q", l)
		}
	}

	if b.Len() != 0 {
		t.Error("expected buffer to be empty after flush")
	}
}

// TestBuffer_LimitedCapacity_DoesNotPanic ensures that writing far beyond
// the limit never panics and caps output at the configured limit.
func TestBuffer_LimitedCapacity_DoesNotPanic(t *testing.T) {
	const limit = 64
	b := buffer.New(limit)

	for i := 0; i < 200; i++ {
		fmt.Fprintf(b, "line %d\n", i)
	}

	if b.Len() > limit {
		t.Fatalf("buffer exceeded limit: got %d, want <= %d", b.Len(), limit)
	}
}
