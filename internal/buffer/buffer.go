package buffer

import (
	"bytes"
	"io"
	"sync"
)

// Buffer is a thread-safe in-memory buffer that accumulates output lines
// and can flush them to a writer. It is used to collect command output
// before writing to the structured log.
type Buffer struct {
	mu   sync.Mutex
	buf  bytes.Buffer
	limit int
}

// New creates a new Buffer. If limit is zero, no byte limit is enforced.
func New(limit int) *Buffer {
	return &Buffer{limit: limit}
}

// Write appends p to the internal buffer. If a byte limit is set and
// the buffer already holds that many bytes, Write is a no-op.
func (b *Buffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.limit > 0 && b.buf.Len() >= b.limit {
		return len(p), nil // silently discard; caller sees no error
	}

	if b.limit > 0 {
		remaining := b.limit - b.buf.Len()
		if len(p) > remaining {
			p = p[:remaining]
		}
	}

	return b.buf.Write(p)
}

// Len returns the number of bytes currently held in the buffer.
func (b *Buffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Len()
}

// Flush writes all buffered bytes to w and resets the buffer.
func (b *Buffer) Flush(w io.Writer) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.buf.Len() == 0 {
		return nil
	}
	_, err := w.Write(b.buf.Bytes())
	b.buf.Reset()
	return err
}

// Bytes returns a copy of the buffered data without resetting.
func (b *Buffer) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]byte, b.buf.Len())
	copy(out, b.buf.Bytes())
	return out
}

// Reset discards all buffered data.
func (b *Buffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf.Reset()
}
