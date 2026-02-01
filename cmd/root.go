package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// debug is set by the --debug flag and enables verbose API logging.
var debug bool

// debugLog prints debug messages to stderr when --debug is enabled.
func debugLog(format string, args ...any) {
	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}

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
	// Global flags available to all commands.
	// Subcommands extract these via cli.FromCommand(cmd).
	rootCmd.PersistentFlags().StringP("output", "o", "json", "Output format: json, toon, text")
	rootCmd.PersistentFlags().StringP("project", "p", "", "Project ID")
	rootCmd.PersistentFlags().StringP("environment", "e", "", "Environment name")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().Bool("no-cache", false, "Bypass cache for this request")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Show API request/response details")

	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("sol version 0.1.0-dev")
		},
	})
}
