// Package jitter provides a Runner decorator that introduces a uniformly
// distributed random delay before each job execution.
//
// When dozens of cron jobs are scheduled at the same minute, they can create
// thundering-herd pressure on downstream systems (databases, APIs, shared
// filesystems). Jitter spreads those bursts by sleeping a random duration in
// the range [0, max) before delegating to the wrapped Runner.
//
// Usage:
//
//	base := runner.New(os.Stdout, os.Stderr)
//	j, err := jitter.New(base, 30*time.Second)
//	if err != nil {
//		log.Fatal(err)
//	}
//	exitCode, err := j.Run(ctx, "backup.sh", args)
//
// The delay is skipped immediately if the context is already cancelled,
// allowing clean shutdown without hanging.
package jitter
