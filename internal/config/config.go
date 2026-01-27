package config

import (
	"os"
	"path/filepath"
)

// Config holds the CLI configuration
type Config struct {
	// Default project and environment
	DefaultProject     string `yaml:"default_project"`
	DefaultEnvironment string `yaml:"default_environment"`

	// Output preferences
	Output OutputConfig `yaml:"output"`

	// Cache settings
	Cache CacheConfig `yaml:"cache"`

	// Network settings
	Network NetworkConfig `yaml:"network"`
}

type OutputConfig struct {
	Format string `yaml:"format"` // json, toon, text
	Color  string `yaml:"color"`  // auto, always, never
}

type CacheConfig struct {
	Enabled    bool `yaml:"enabled"`
	TTLSeconds int  `yaml:"ttl_seconds"`
}

type NetworkConfig struct {
	Timeout string `yaml:"timeout"`
	Retries int    `yaml:"retries"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Output: OutputConfig{
			Format: "json",
			Color:  "auto",
		},
		Cache: CacheConfig{
			Enabled:    true,
			TTLSeconds: 600, // 10 minutes
		},
		Network: NetworkConfig{
			Timeout: "30s",
			Retries: 3,
		},
	}
}

// ConfigDir returns the configuration directory path
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".sol"), nil
}

// ConfigPath returns the configuration file path
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// CacheDir returns the cache directory path
func CacheDir() (string, error) {
	// In sandbox mode, use /tmp
	if IsSourceOperation() {
		return "/tmp/.sol/cache", nil
	}

	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "cache"), nil
}

// IsSourceOperation detects if running in Upsun source operation
func IsSourceOperation() bool {
	return os.Getenv("PLATFORM_APPLICATION") != "" &&
		os.Getenv("PLATFORM_SOURCE_OPERATION") != ""
}

// Load loads configuration from file and environment
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// TODO: load from config file
	// TODO: override with environment variables

	return cfg, nil
}
