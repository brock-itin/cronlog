package label_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/cronlog/internal/label"
)

// TestLabel_RoundTrip verifies that many lines written in chunks all arrive
// correctly prefixed in the destination writer.
func TestLabel_RoundTrip(t *testing.T) {
	var buf bytes.Buffer
	w := label.New(&buf, "cron")

	const lineCount = 20
	for i := 0; i < lineCount; i++ {
		fmt.Fprintf(w, "line %d\n", i)
	}
	w.Close()

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != lineCount {
		t.Fatalf("expected %d lines, got %d", lineCount, len(lines))
	}
	for i, l := range lines {
		expected := fmt.Sprintf("[cron] line %d", i)
		if l != expected {
			t.Errorf("line %d: expected %q, got %q", i, expected, l)
		}
	}
}

// TestLabel_InterleavedStreams simulates stdout and stderr being labelled
// separately and merged into a single destination.
func TestLabel_InterleavedStreams(t *testing.T) {
	var buf bytes.Buffer
	stdout := label.New(&buf, "stdout")
	stderr := label.New(&buf, "stderr")

	stdout.Write([]byte("result ok\n"))
	stderr.Write([]byte("warning: low disk\n"))
	stdout.Write([]byte("done\n"))

	got := buf.String()
	if !strings.Contains(got, "[stdout] result ok") {
		t.Errorf("missing stdout line in: %q", got)
	}
	if !strings.Contains(got, "[stderr] warning: low disk") {
		t.Errorf("missing stderr line in: %q", got)
	}
	if !strings.Contains(got, "[stdout] done") {
		t.Errorf("missing final stdout line in: %q", got)
	}
}
