// Package buffer provides a thread-safe, optionally size-limited in-memory
// byte buffer for accumulating command output before it is written to a
// structured log entry.
//
// # Overview
//
// Buffer wraps a [bytes.Buffer] with a mutex so it can be safely written
// from the stdout and stderr goroutines of a running job simultaneously.
// An optional byte limit prevents unbounded memory growth for chatty jobs;
// bytes that exceed the limit are silently discarded (consistent with the
// behaviour of [internal/truncate]).
//
// # Usage
//
//	b := buffer.New(1 << 20) // 1 MiB cap
//	cmd.Stdout = b
//	cmd.Stderr = b
//	cmd.Run()
//	b.Flush(logWriter)
package buffer
