// Package history provides persistent run-history tracking for cron jobs.
//
// Each invocation of cronlog can record an Entry describing the job name,
// start time, duration, exit code, and the path to the associated log file.
// Entries are stored as a JSON array on disk and automatically pruned to a
// configurable maximum size so the history file does not grow unbounded.
//
// Typical usage:
//
//	h, err := history.New("/var/lib/cronlog/backup.history.json", 50)
//	if err != nil { /* handle */ }
//
//	err = h.Add(history.Entry{
//		Job:       "backup",
//		StartedAt: time.Now(),
//		Duration:  duration,
//		ExitCode:  exitCode,
//		LogFile:   "/var/log/cronlog/backup-2024-01-01.log",
//	})
package history
