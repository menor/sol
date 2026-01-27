package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   string
	}{
		{"first wins", []string{"a", "b", "c"}, "a"},
		{"skip empty", []string{"", "b", "c"}, "b"},
		{"skip multiple empty", []string{"", "", "c"}, "c"},
		{"all empty", []string{"", "", ""}, ""},
		{"no values", []string{}, ""},
		{"single value", []string{"only"}, "only"},
		{"single empty", []string{""}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firstNonEmpty(tt.values...); got != tt.want {
				t.Errorf("firstNonEmpty() = %q, want %q", got, tt.want)
			}
		})
	}
}

// newTestCommand creates a cobra.Command with the same flags as rootCmd.
// This allows testing FromCommand without importing the cmd package.
func newTestCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "json", "")
	cmd.Flags().StringP("project", "p", "", "")
	cmd.Flags().StringP("environment", "e", "", "")
	cmd.Flags().BoolP("quiet", "q", false, "")
	cmd.Flags().Bool("no-cache", false, "")
	return cmd
}

func TestFromCommand(t *testing.T) {
	cmd := newTestCommand()
	cmd.Flags().Set("output", "text")
	cmd.Flags().Set("project", "my-project")
	cmd.Flags().Set("environment", "staging")
	cmd.Flags().Set("quiet", "true")
	cmd.Flags().Set("no-cache", "true")

	cfg, err := FromCommand(cmd)
	if err != nil {
		t.Fatalf("FromCommand() error = %v", err)
	}

	if cfg.Output != "text" {
		t.Errorf("Output = %q, want %q", cfg.Output, "text")
	}
	if cfg.ProjectID != "my-project" {
		t.Errorf("ProjectID = %q, want %q", cfg.ProjectID, "my-project")
	}
	if cfg.Environment != "staging" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "staging")
	}
	if !cfg.Quiet {
		t.Error("Quiet = false, want true")
	}
	if !cfg.NoCache {
		t.Error("NoCache = false, want true")
	}
}

func TestFromCommandDefaults(t *testing.T) {
	cmd := newTestCommand()

	cfg, err := FromCommand(cmd)
	if err != nil {
		t.Fatalf("FromCommand() error = %v", err)
	}

	if cfg.Output != "json" {
		t.Errorf("Output = %q, want %q (default)", cfg.Output, "json")
	}
	if cfg.Quiet {
		t.Error("Quiet = true, want false (default)")
	}
	if cfg.NoCache {
		t.Error("NoCache = true, want false (default)")
	}
}

func TestFromCommandEnvFallback(t *testing.T) {
	cmd := newTestCommand()

	// Set env vars
	t.Setenv("UPSUN_PROJECT", "env-project")
	t.Setenv("UPSUN_ENVIRONMENT", "env-staging")

	cfg, err := FromCommand(cmd)
	if err != nil {
		t.Fatalf("FromCommand() error = %v", err)
	}

	if cfg.ProjectID != "env-project" {
		t.Errorf("ProjectID = %q, want %q (from env)", cfg.ProjectID, "env-project")
	}
	if cfg.Environment != "env-staging" {
		t.Errorf("Environment = %q, want %q (from env)", cfg.Environment, "env-staging")
	}
}

func TestFromCommandFlagOverridesEnv(t *testing.T) {
	cmd := newTestCommand()
	cmd.Flags().Set("project", "flag-project")

	// Set env var that should be overridden
	t.Setenv("UPSUN_PROJECT", "env-project")

	cfg, err := FromCommand(cmd)
	if err != nil {
		t.Fatalf("FromCommand() error = %v", err)
	}

	// Flag should win over env
	if cfg.ProjectID != "flag-project" {
		t.Errorf("ProjectID = %q, want %q (flag should override env)", cfg.ProjectID, "flag-project")
	}
}

func TestFromCommandMissingFlag(t *testing.T) {
	// Command without required flags
	cmd := &cobra.Command{}

	_, err := FromCommand(cmd)
	if err == nil {
		t.Error("FromCommand() should error when flags are missing")
	}
}
