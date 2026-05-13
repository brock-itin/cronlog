// Package audit provides append-only structured audit logging for cron job
// lifecycle events.
//
// Each call to Record appends a single JSON line to the configured audit file,
// capturing the job name, event type (e.g. "start", "done", "error"), exit
// code, and optional key/value metadata.
//
// Usage:
//
//	auditor, err := audit.New("/var/log/cronlog/audit.jsonl")
//	if err != nil { ... }
//	defer auditor.Close()
//
//	auditor.Record("backup", "start", 0, nil)
//	// ... run job ...
//	auditor.Record("backup", "done", exitCode, map[string]string{"duration": "4s"})
//
// The audit log is distinct from the rotated job output log; it is intended
// as a lightweight, tamper-evident trail suitable for compliance or debugging.
package audit
