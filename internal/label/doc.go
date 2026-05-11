// Package label provides an io.Writer wrapper that prepends a configurable
// label prefix to every line of output passing through it.
//
// It is useful for annotating captured stdout and stderr streams with a job
// name or stream identifier before they are forwarded to a log rotator or
// structured logger, making it easy to distinguish the origin of each line
// in aggregated output.
//
// Usage:
//
//	out := label.New(rotatorWriter, "my-job")
//	cmd.Stdout = out
//	cmd.Stderr = label.New(rotatorWriter, "my-job:stderr")
//	defer out.Close()
package label
