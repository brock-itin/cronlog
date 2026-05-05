//go:build integration

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/cronlog/internal/config"
)

// TestLoad_ExampleConfig ensures the shipped example config parses without error.
func TestLoad_ExampleConfig(t *testing.T) {
	examplePath := filepath.Join("..", "..", "config", "cronlog.example.yaml")
	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		t.Skipf("example config not found at %s", examplePath)
	}

	cfg, err := config.Load(examplePath)
	if err != nil {
		t.Fatalf("Load(%q): %v", examplePath, err)
	}

	if cfg.LogDir == "" {
		t.Error("expected non-empty LogDir")
	}
	if cfg.MaxFiles < 1 {
		t.Errorf("MaxFiles = %d, want >= 1", cfg.MaxFiles)
	}
	if cfg.MaxSizeMB < 1 {
		t.Errorf("MaxSizeMB = %d, want >= 1", cfg.MaxSizeMB)
	}
	if cfg.Timeout <= 0 {
		t.Errorf("Timeout = %v, want > 0", cfg.Timeout)
	}
}
