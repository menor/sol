package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/menor/sol/internal/output"
)

// Config holds all CLI configuration extracted from flags, env vars, and config file.
// This struct makes dependencies explicit and simplifies testing.
type Config struct {
	Output      string
	ProjectID   string
	Environment string
	Quiet       bool
	NoCache     bool
}

// FromCommand extracts configuration from a cobra command.
// Precedence: flag > environment variable > config file.
func FromCommand(cmd *cobra.Command) (*Config, error) {
	cfg := &Config{}

	// Output format (flag only, no env fallback)
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return nil, err
	}
	cfg.Output = output

	// Project ID: flag > env > config
	projectID, err := cmd.Flags().GetString("project")
	if err != nil {
		return nil, err
	}
	cfg.ProjectID = firstNonEmpty(
		projectID,
		os.Getenv("UPSUN_PROJECT"),
		// TODO: read from config file
	)

	// Environment: flag > env > config
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return nil, err
	}
	cfg.Environment = firstNonEmpty(
		environment,
		os.Getenv("UPSUN_ENVIRONMENT"),
		// TODO: read from config file
	)

	// Quiet mode (flag only)
	quiet, err := cmd.Flags().GetBool("quiet")
	if err != nil {
		return nil, err
	}
	cfg.Quiet = quiet

	// No cache (flag only)
	noCache, err := cmd.Flags().GetBool("no-cache")
	if err != nil {
		return nil, err
	}
	cfg.NoCache = noCache

	return cfg, nil
}

// Formatter returns an output formatter configured for this Config.
// Use this instead of output.New() directly to ensure --output flag is respected.
func (c *Config) Formatter() output.Formatter {
	return output.New(c.Output)
}

// firstNonEmpty returns the first non-empty string from the arguments.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
