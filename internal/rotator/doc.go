// Package rotator provides a size-based, rotating log file writer for cronlog.
//
// A Rotator opens a new timestamped log file in a configured directory and
// automatically rotates it once the file reaches a maximum size. Old log
// files beyond a configurable retention count are pruned automatically.
//
// Usage:
//
//	r, err := rotator.New(rotator.Config{
//		Dir:       "/var/log/cronlog",
//		BaseName:  "backup-job",
//		MaxSizeMB: 10,
//		MaxFiles:  5,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer r.Close()
//
//	// Use as an io.Writer
//	fmt.Fprintln(r, "cron job started")
//
// Log files are named using the pattern: <BaseName>-<timestamp>.log
// where timestamp is formatted as 20060102T150405Z (UTC).
//
// The Rotator is safe for concurrent use.
package rotator
