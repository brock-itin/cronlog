package label

import (
	"fmt"
	"io"
	"strings"
)

// Writer wraps an io.Writer and prepends a fixed label prefix to every line
// written through it. This is useful for distinguishing stdout vs stderr or
// tagging output with a job name in structured logs.
type Writer struct {
	dst    io.Writer
	prefix string
	buf    strings.Builder
}

// New returns a Writer that prepends prefix to each line before forwarding to
// dst. The prefix is separated from the line content by a single space.
func New(dst io.Writer, prefix string) *Writer {
	return &Writer{
		dst:    dst,
		prefix: fmt.Sprintf("[%s] ", prefix),
	}
}

// Write implements io.Writer. It buffers data until newlines are found, then
// flushes each complete line with the configured prefix.
func (w *Writer) Write(p []byte) (int, error) {
	w.buf.Write(p)
	for {
		s := w.buf.String()
		idx := strings.IndexByte(s, '\n')
		if idx < 0 {
			break
		}
		line := s[:idx+1]
		w.buf.Reset()
		w.buf.WriteString(s[idx+1:])
		if _, err := fmt.Fprint(w.dst, w.prefix+line); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

// Close flushes any remaining buffered data that did not end with a newline.
func (w *Writer) Close() error {
	if w.buf.Len() == 0 {
		return nil
	}
	_, err := fmt.Fprintln(w.dst, w.prefix+w.buf.String())
	w.buf.Reset()
	return err
}
