// Package prefix provides a writer that prepends a static string to every
// line of output before forwarding it to an underlying io.Writer.
//
// It is useful when merging stdout and stderr into a single log stream and
// you want to distinguish the source of each line.
package prefix

import (
	"bytes"
	"io"
)

// Writer wraps an io.Writer and prepends a fixed prefix to every line.
type Writer struct {
	dst    io.Writer
	prefix []byte
	buf    []byte
}

// New returns a Writer that prepends prefix to every line written to dst.
// A trailing newline is appended to each flushed line if not already present.
func New(dst io.Writer, prefix string) *Writer {
	return &Writer{
		dst:    dst,
		prefix: []byte(prefix),
	}
}

// Write buffers p and flushes complete lines to the underlying writer with
// the configured prefix prepended. Partial lines are held until the next
// Write or Close call.
func (w *Writer) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)

	for {
		idx := bytes.IndexByte(w.buf, '\n')
		if idx < 0 {
			break
		}
		line := w.buf[:idx+1]
		if err := w.writeLine(line); err != nil {
			return 0, err
		}
		w.buf = w.buf[idx+1:]
	}

	return len(p), nil
}

// Close flushes any remaining buffered data (without a trailing newline) to
// the underlying writer with the prefix prepended, then releases resources.
func (w *Writer) Close() error {
	if len(w.buf) == 0 {
		return nil
	}
	line := append(w.buf, '\n')
	w.buf = nil
	return w.writeLine(line)
}

func (w *Writer) writeLine(line []byte) error {
	out := make([]byte, 0, len(w.prefix)+len(line))
	out = append(out, w.prefix...)
	out = append(out, line...)
	_, err := w.dst.Write(out)
	return err
}
