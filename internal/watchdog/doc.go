// Package watchdog wraps a Runner with overdue-detection logic.
//
// A Watchdog records a checkpoint after each successful job execution and
// compares the elapsed time against a configured interval on the next run.
// When the gap exceeds the interval the supplied Notifier is called so that
// operators are alerted to missed or silently-skipped cron jobs.
//
// Typical usage:
//
//	cp, _ := checkpoint.New("/var/lib/cronlog/watchdog.json")
//	wd, _ := watchdog.New(runner, cp, notifier, 25*time.Hour)
//	exitCode, err := wd.Run("daily-backup", "--target", "/mnt/backups")
//
// The Watchdog itself satisfies the runner.Runner interface so it can be
// composed freely with other middleware in a pipeline.
package watchdog
