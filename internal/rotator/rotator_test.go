package rotator_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/example/cronlog/internal/rotator"
)

func TestNew_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "logs", "sub")
	r, err := rotator.New(rotator.Config{
		Dir:      subDir,
		BaseName: "test",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer r.Close()
	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		t.Error("expected directory to be created")
	}
}

func TestWrite_CreatesLogFile(t *testing.T) {
	dir := t.TempDir()
	r, err := rotator.New(rotator.Config{
		Dir:      dir,
		BaseName: "myjob",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer r.Close()

	_, err = r.Write([]byte("hello log\n"))
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	matches, _ := filepath.Glob(filepath.Join(dir, "myjob-*.log"))
	if len(matches) == 0 {
		t.Error("expected at least one log file")
	}
}

func TestRotation_PrunesOldFiles(t *testing.T) {
	dir := t.TempDir()
	cfg := rotator.Config{
		Dir:       dir,
		BaseName:  "job",
		MaxSizeMB: 1,
		MaxFiles:  2,
	}

	// Pre-create fake old log files
	for _, name := range []string{"job-20230101T000000Z.log", "job-20230102T000000Z.log"} {
		f, _ := os.Create(filepath.Join(dir, name))
		f.Close()
	}

	r, err := rotator.New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer r.Close()

	// Simulate size exceeding limit by writing a large payload
	big := strings.Repeat("x", int(cfg.MaxSizeMB*1024*1024)+1)
	_, _ = r.Write([]byte(big))

	matches, _ := filepath.Glob(filepath.Join(dir, "job-*.log"))
	if len(matches) > cfg.MaxFiles {
		t.Errorf("expected at most %d files, got %d", cfg.MaxFiles, len(matches))
	}
}
