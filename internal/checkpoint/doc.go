// Package checkpoint provides persistent job-run bookmarking for cronlog.
//
// After each successful execution cronlog can record a checkpoint so that
// operators or alerting rules can detect jobs that have not run within their
// expected cadence.
//
// # Usage
//
//	cp := checkpoint.New("/var/lib/cronlog/backup.checkpoint.json")
//
//	// Record a successful run.
//	_ = cp.Save(checkpoint.Entry{Job: "backup", LastOK: time.Now().UTC()})
//
//	// Check whether the job is overdue by more than 25 hours.
//	overdue, err := cp.Overdue("backup", 25*time.Hour)
//
// Checkpoint files are plain JSON and safe to inspect or reset manually.
package checkpoint
