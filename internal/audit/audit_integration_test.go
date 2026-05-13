package audit_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/cronlog/cronlog/internal/audit"
)

func TestAudit_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")

	a, err := audit.New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	events := []struct {
		event string
		code  int
	}{
		{"start", 0},
		{"done", 0},
	}
	for _, e := range events {
		if err := a.Record("nightly-backup", e.event, e.code, map[string]string{"env": "prod"}); err != nil {
			t.Fatalf("Record %s: %v", e.event, err)
		}
	}
	if err := a.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var entries []audit.Entry
	for scanner.Scan() {
		var e audit.Entry
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		entries = append(entries, e)
	}

	if len(entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(entries))
	}
	if entries[0].Event != "start" || entries[1].Event != "done" {
		t.Errorf("unexpected events: %v", entries)
	}
	if entries[0].Meta["env"] != "prod" {
		t.Errorf("meta env = %q, want prod", entries[0].Meta["env"])
	}
}
