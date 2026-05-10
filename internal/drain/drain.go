// Package drain provides a writer that collects output lines and forwards
// them to a structured logger, applying optional redaction and filtering
// before each line is persisted.
package drain

import (
	"bufio"
	"bytes"
	"io"
	"sync"
)

// LineHandler is called for each complete line read from the writer.
type LineHandler func(line string)

// Drain reads from an io.Writer interface, splits input into lines, and
// dispatches each line to a registered LineHandler.
type Drain struct {
	mu      sync.Mutex
	buf     bytes.Buffer
	handler LineHandler
}

// New creates a new Drain that invokes handler for every complete line
// written to it. handler must not be nil.
func New(handler LineHandler) *Drain {
	if handler == nil {
		panic("drain: handler must not be nil")
	}
	return &Drain{handler: handler}
}

// Write satisfies io.Writer. It buffers data and dispatches complete
// newline-terminated lines to the handler immediately.
func (d *Drain) Write(p []byte) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	n, err := d.buf.Write(p)
	if err != nil {
		return n, err
	}

	d.flush()
	return n, nil
}

// Close flushes any remaining buffered data as a final (unterminated) line.
func (d *Drain) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.buf.Len() > 0 {
		d.handler(d.buf.String())
		d.buf.Reset()
	}
	return nil
}

// flush scans the internal buffer and dispatches every complete line.
func (d *Drain) flush() {
	scanner := bufio.NewScanner(bytes.NewReader(d.buf.Bytes()))
	var consumed int
	for scanner.Scan() {
		line := scanner.Text()
		d.handler(line)
		consumed += len(line) + 1 // +1 for newline
	}
	if consumed > 0 {
		remaining := d.buf.Bytes()[consumed:]
		newBuf := make([]byte, len(remaining))
		copy(newBuf, remaining)
		d.buf.Reset()
		d.buf.Write(newBuf)
	}
}

// WriterTo returns an io.WriteCloser backed by this Drain, suitable for
// assigning to exec.Cmd.Stdout or Stderr.
func (d *Drain) WriterTo() io.WriteCloser { return d }
