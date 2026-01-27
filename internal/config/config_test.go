package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsSourceOperation(t *testing.T) {
	tests := []struct {
		name      string
		sourceDir string
		want      bool
	}{
		{"set", "/app", true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Always set env var to ensure known state, even if empty
			t.Setenv("PLATFORM_SOURCE_DIR", tt.sourceDir)

			if got := IsSourceOperation(); got != tt.want {
				t.Errorf("IsSourceOperation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Output.Format != "json" {
		t.Errorf("Output.Format = %q, want %q", cfg.Output.Format, "json")
	}
	if cfg.Output.Color != "auto" {
		t.Errorf("Output.Color = %q, want %q", cfg.Output.Color, "auto")
	}
	if !cfg.Cache.Enabled {
		t.Error("Cache.Enabled = false, want true")
	}
	if cfg.Cache.TTLSeconds != 600 {
		t.Errorf("Cache.TTLSeconds = %d, want %d", cfg.Cache.TTLSeconds, 600)
	}
	if cfg.Network.Timeout != "30s" {
		t.Errorf("Network.Timeout = %q, want %q", cfg.Network.Timeout, "30s")
	}
	if cfg.Network.Retries != 3 {
		t.Errorf("Network.Retries = %d, want %d", cfg.Network.Retries, 3)
	}
}

func TestConfigDir(t *testing.T) {
	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error = %v", err)
	}

	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".sol")

	if dir != want {
		t.Errorf("ConfigDir() = %q, want %q", dir, want)
	}
}

func TestConfigPath(t *testing.T) {
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() error = %v", err)
	}

	if !strings.HasSuffix(path, ".sol/config.yaml") {
		t.Errorf("ConfigPath() = %q, want suffix %q", path, ".sol/config.yaml")
	}
}

func TestCacheDir(t *testing.T) {
	// Normal case (not in source operation)
	dir, err := CacheDir()
	if err != nil {
		t.Fatalf("CacheDir() error = %v", err)
	}

	if !strings.HasSuffix(dir, ".sol/cache") {
		t.Errorf("CacheDir() = %q, want suffix %q", dir, ".sol/cache")
	}
}

func TestCacheDirInSourceOperation(t *testing.T) {
	t.Setenv("PLATFORM_SOURCE_DIR", "/app")

	dir, err := CacheDir()
	if err != nil {
		t.Fatalf("CacheDir() error = %v", err)
	}

	if dir != "/tmp/.sol/cache" {
		t.Errorf("CacheDir() = %q, want %q", dir, "/tmp/.sol/cache")
	}
}

func TestLoad(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Should return defaults since no config file exists
	if cfg.Output.Format != "json" {
		t.Errorf("Load() returned non-default config")
	}
}
