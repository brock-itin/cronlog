// Package heartbeat provides a mechanism for sending periodic HTTP pings
// to an external monitoring service (e.g. Healthchecks.io, Better Uptime)
// to signal that a cron job is alive and completed successfully.
package heartbeat

import (
	"fmt"
	"net/http"
	"time"
)

// Pinger sends heartbeat signals to a monitoring URL.
type Pinger struct {
	url    string
	client *http.Client
}

// New creates a Pinger that will POST to the given URL.
// If url is empty, Ping is a no-op. timeout controls the HTTP request deadline.
func New(url string, timeout time.Duration) *Pinger {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &Pinger{
		url: url,
		client: &http.Client{Timeout: timeout},
	}
}

// Ping sends a GET request to the heartbeat URL.
// It appends "/fail" to the URL when exitCode is non-zero, following the
// Healthchecks.io convention. Returns nil when url is empty.
func (p *Pinger) Ping(exitCode int) error {
	if p.url == "" {
		return nil
	}

	target := p.url
	if exitCode != 0 {
		target = target + "/fail"
	}

	resp, err := p.client.Get(target)
	if err != nil {
		return fmt.Errorf("heartbeat: GET %s: %w", target, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("heartbeat: unexpected status %d from %s", resp.StatusCode, target)
	}

	return nil
}
