// Package drain implements a line-oriented io.WriteCloser that splits a
// continuous stream of bytes (typically the stdout or stderr of a subprocess)
// into individual lines and dispatches each line to a caller-supplied
// LineHandler.
//
// Typical usage:
//
//	d := drain.New(func(line string) {
//		logger.Info(line)
//	})
//	cmd.Stdout = d
//	cmd.Stderr = d
//	cmd.Run()
//	d.Close() // flush any trailing output not terminated by a newline
//
// The Drain is safe for concurrent use; multiple goroutines may write to it
// simultaneously without external synchronisation.
package drain
