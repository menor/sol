package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	outputFormat string
	projectID    string
	environment  string
	quiet        bool
	noCache      bool
)

var rootCmd = &cobra.Command{
	Use:   "sol",
	Short: "Agent-first CLI for Upsun",
	Long: `Sol is a minimal, agent-first CLI for Upsun.

It optimizes for code agents first, humans second:
  - Structured JSON output by default
  - No interactive prompts
  - Predictable exit codes
  - Machine-readable errors

Example:
  sol project:list --output json
  sol ssh --project abc123 --environment main`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags available to all commands
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "json", "Output format: json, toon, text")
	rootCmd.PersistentFlags().StringVarP(&projectID, "project", "p", "", "Project ID")
	rootCmd.PersistentFlags().StringVarP(&environment, "environment", "e", "", "Environment name")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().BoolVar(&noCache, "no-cache", false, "Bypass cache for this request")

	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("sol version 0.1.0-dev")
		},
	})
}

// GetOutputFormat returns the output format flag value
func GetOutputFormat() string {
	return outputFormat
}

// GetProjectID returns project ID from flag, config, or environment
func GetProjectID() string {
	if projectID != "" {
		return projectID
	}
	// TODO: read from config file
	// TODO: read from PLATFORM_PROJECT env var
	return os.Getenv("UPSUN_PROJECT")
}

// GetEnvironment returns environment from flag, config, or environment
func GetEnvironment() string {
	if environment != "" {
		return environment
	}
	// TODO: read from config file
	// TODO: read from PLATFORM_BRANCH env var
	return os.Getenv("UPSUN_ENVIRONMENT")
}
