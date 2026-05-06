// Package notify provides optional alerting when a cron job fails.
// It supports sending a summary message to a configured webhook URL
// (e.g. Slack, generic HTTP endpoint) on non-zero exit codes.
package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Notifier sends failure notifications to a webhook endpoint.
type Notifier struct {
	webhookURL string
	client     *http.Client
}

// Payload is the JSON body sent to the webhook.
type Payload struct {
	Job      string `json:"job"`
	ExitCode int    `json:"exit_code"`
	Message  string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// New creates a Notifier that posts to the given webhook URL.
// A zero-value webhookURL disables notifications silently.
func New(webhookURL string) *Notifier {
	return &Notifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Notify sends a failure notification for the given job name and exit code.
// It is a no-op when the webhook URL is empty or the exit code is zero.
func (n *Notifier) Notify(job string, exitCode int) error {
	if n.webhookURL == "" || exitCode == 0 {
		return nil
	}

	p := Payload{
		Job:       job,
		ExitCode:  exitCode,
		Message:   fmt.Sprintf("cron job %q exited with code %d", job, exitCode),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("notify: marshal payload: %w", err)
	}

	resp, err := n.client.Post(n.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notify: post webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("notify: webhook returned status %d", resp.StatusCode)
	}

	return nil
}
