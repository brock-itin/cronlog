package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "audit.jsonl")
}

func TestRecord_WritesJSONLine(t *testing.T) {
	p := tempPath(t)
	a, err := New(p)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer a.Close()

	if err := a.Record("backup", "start", 0, nil); err != nil {
		t.Fatalf("Record: %v", err)
	}

	f, _ := os.Open(p)
	defer f.Close()
	var entry Entry
	if err := json.NewDecoder(f).Decode(&entry); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if entry.Job != "backup" {
		t.Errorf("job = %q, want backup", entry.Job)
	}
	if entry.Event != "start" {
		t.Errorf("event = %q, want start", entry.Event)
	}
}

func TestRecord_IncludesExitCode(t *testing.T) {
	p := tempPath(t)
	a, _ := New(p)
	defer a.Close()

	a.Record("cleanup", "done", 2, nil)

	f, _ := os.Open(p)
	defer f.Close()
	var entry Entry
	json.NewDecoder(f).Decode(&entry)
	if entry.ExitCode != 2 {
		t.Errorf("exit_code = %d, want 2", entry.ExitCode)
	}
}

func TestRecord_IncludesMeta(t *testing.T) {
	p := tempPath(t)
	a, _ := New(p)
	defer a.Close()

	a.Record("sync", "finish", 0, map[string]string{"host": "srv1"})

	f, _ := os.Open(p)
	defer f.Close()
	var entry Entry
	json.NewDecoder(f).Decode(&entry)
	if entry.Meta["host"] != "srv1" {
		t.Errorf("meta host = %q, want srv1", entry.Meta["host"])
	}
}

func TestRecord_MultipleEntries(t *testing.T) {
	p := tempPath(t)
	a, _ := New(p)
	defer a.Close()

	for i := 0; i < 3; i++ {
		a.Record("job", "tick", 0, nil)
	}
	a.Close()

	f, _ := os.Open(p)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		count++
	}
	if count != 3 {
		t.Errorf("lines = %d, want 3", count)
	}
}

func TestRecord_TimestampIsUTC(t *testing.T) {
	p := tempPath(t)
	a, _ := New(p)
	a.now = func() time.Time { return time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC) }
	defer a.Close()

	a.Record("job", "run", 0, nil)

	f, _ := os.Open(p)
	defer f.Close()
	var entry Entry
	json.NewDecoder(f).Decode(&entry)
	if entry.Timestamp.Location() != time.UTC {
		t.Errorf("timestamp not UTC")
	}
}
