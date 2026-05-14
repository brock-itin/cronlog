// Package cooldown provides a Runner decorator that enforces a minimum
// quiet period between consecutive job executions.
//
// Unlike [throttle], which gates runs based on the last *successful*
// execution, cooldown applies regardless of outcome. This prevents
// rapid re-execution when a cron schedule fires more frequently than
// the job's safe repetition rate — for example, a job that modifies
// shared state and must not run twice within 10 minutes.
//
// Usage:
//
//	cp, err := checkpoint.New("/var/lib/cronlog/myjob.cooldown.json")
//	if err != nil { ... }
//
//	cd, err := cooldown.New(baseRunner, cp, 10*time.Minute)
//	if err != nil { ... }
//
//	code, err := cd.Run("myjob", args)
//
// If the job ran less than 10 minutes ago, Run returns exit code 0
// and a non-nil error describing the remaining cooldown duration.
// The checkpoint is updated after every execution, including failures.
package cooldown
