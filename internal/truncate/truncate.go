// Package truncate provides output truncation for cron job runs,
// capping captured output at a configurable maximum byte size to
// prevent runaway jobs from filling disk space.
package truncate

import (
	"fmt"
	"io"
)

// Truncator wraps a writer and limits the total number of bytes written.
// Once the limit is reached, further writes are silently discarded and
// a truncation notice is appended when Close is called.
type Truncator struct {
	dst     io.Writer
	limit   int64
	written int64
	tripped bool
}

// New returns a Truncator that forwards writes to dst until maxBytes
// have been written. A maxBytes value of zero disables truncation.
func New(dst io.Writer, maxBytes int64) *Truncator {
	return &Truncator{
		dst:   dst,
		limit: maxBytes,
	}
}

// Write implements io.Writer. Bytes beyond the configured limit are dropped.
func (t *Truncator) Write(p []byte) (int, error) {
	if t.limit <= 0 {
		return t.dst.Write(p)
	}

	remaining := t.limit - t.written
	if remaining <= 0 {
		t.tripped = true
		// Report success to the caller so the source keeps running.
		return len(p), nil
	}

	allowed := int64(len(p))
	if allowed > remaining {
		allowed = remaining
		t.tripped = true
	}

	n, err := t.dst.Write(p[:allowed])
	t.written += int64(n)
	// Return the full length so callers do not treat a partial write as an error.
	return len(p), err
}

// Close writes a truncation notice to the underlying writer when the limit
// was exceeded. It does not close dst itself.
func (t *Truncator) Close() error {
	if !t.tripped {
		return nil
	}
	notice := fmt.Sprintf("\n[cronlog] output truncated after %d bytes\n", t.limit)
	_, err := fmt.Fprint(t.dst, notice)
	return err
}

// Tripped reports whether the byte limit was reached during writing.
func (t *Truncator) Tripped() bool {
	return t.tripped
}

// Written returns the number of bytes forwarded to the underlying writer.
func (t *Truncator) Written() int64 {
	return t.written
}
