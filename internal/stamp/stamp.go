// Package stamp injects a timestamp prefix into each line of output,
// allowing log consumers to correlate output lines with wall-clock time.
package stamp

import (
	"bytes"
	"fmt"
	"io"
	"time"
)

// Writer wraps an io.Writer and prepends a formatted timestamp to every
// newline-delimited line it receives.
type Writer struct {
	dst    io.Writer
	format string
	clock  func() time.Time
	buf    []byte
}

// New returns a Writer that writes to dst, prefixing each line with a
// timestamp formatted according to format. If format is empty the RFC3339
// layout is used.
func New(dst io.Writer, format string) *Writer {
	if format == "" {
		format = time.RFC3339
	}
	return &Writer{
		dst:    dst,
		format: format,
		clock:  time.Now,
	}
}

// Write satisfies io.Writer. Each complete line in p is prefixed with the
// current timestamp before being forwarded to the underlying writer.
// Partial lines (no trailing newline) are buffered until completed.
func (w *Writer) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)

	for {
		idx := bytes.IndexByte(w.buf, '\n')
		if idx < 0 {
			break
		}
		line := w.buf[:idx+1]
		prefixed := fmt.Sprintf("%s %s", w.clock().Format(w.format), line)
		if _, err := io.WriteString(w.dst, prefixed); err != nil {
			return 0, err
		}
		w.buf = w.buf[idx+1:]
	}

	return len(p), nil
}

// Close flushes any buffered data that was not terminated with a newline.
func (w *Writer) Close() error {
	if len(w.buf) == 0 {
		return nil
	}
	prefixed := fmt.Sprintf("%s %s\n", w.clock().Format(w.format), w.buf)
	w.buf = nil
	_, err := io.WriteString(w.dst, prefixed)
	return err
}
