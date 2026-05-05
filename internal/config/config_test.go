package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/cronlog/internal/config"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "cronlog.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeConfig: %v", err)
	}
	return p
}

func TestLoad_Defaults(t *testing.T) {
	p := writeConfig(t, "job_name: backup\n")
	cfg, err := config.Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LogDir != config.DefaultLogDir {
		t.Errorf("LogDir = %q, want %q", cfg.LogDir, config.DefaultLogDir)
	}
	if cfg.MaxFiles != config.DefaultMaxFiles {
		t.Errorf("MaxFiles = %d, want %d", cfg.MaxFiles, config.DefaultMaxFiles)
	}
	if cfg.Timeout != config.DefaultTimeout {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, config.DefaultTimeout)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	p := writeConfig(t, "log_dir: /tmp/logs\nmax_files: 3\nmax_size_mb: 10\ntimeout: 5m\njob_name: sync\n")
	cfg, err := config.Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LogDir != "/tmp/logs" {
		t.Errorf("LogDir = %q", cfg.LogDir)
	}
	if cfg.MaxFiles != 3 {
		t.Errorf("MaxFiles = %d", cfg.MaxFiles)
	}
	if cfg.Timeout != 5*time.Minute {
		t.Errorf("Timeout = %v", cfg.Timeout)
	}
}

func TestLoad_InvalidMaxFiles(t *testing.T) {
	p := writeConfig(t, "max_files: 0\n")
	_, err := config.Load(p)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/path/cronlog.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	p := writeConfig(t, ": : invalid: yaml:::")
	_, err := config.Load(p)
	if err == nil {
		t.Fatal("expected parse error")
	}
}
