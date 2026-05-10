// Package heartbeat sends HTTP pings to an external uptime-monitoring service
// after a cron job completes.
//
// Usage:
//
//	p := heartbeat.New(cfg.HeartbeatURL, 10*time.Second)
//	if err := p.Ping(exitCode); err != nil {
//		log.Printf("heartbeat failed: %v", err)
//	}
//
// When the job exits with code 0 the base URL is pinged directly.
// When the job exits with a non-zero code "/fail" is appended to the URL,
// following the Healthchecks.io convention so the monitoring service can
// distinguish a successful check-in from a failure signal.
//
// If HeartbeatURL is empty all calls are silent no-ops, making the feature
// fully opt-in with no configuration required.
package heartbeat
