package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"lab.plat.farm/menor/sol/internal/api"
	"lab.plat.farm/menor/sol/internal/cli"
	"lab.plat.farm/menor/sol/internal/errors"
)

func init() {
	rootCmd.AddCommand(sshCmd)

	// --app is ssh-specific, not a global flag
	sshCmd.Flags().StringP("app", "A", "", "Application name (for multi-app projects)")
}

var sshCmd = &cobra.Command{
	Use:   "ssh [-- command]",
	Short: "SSH into an environment",
	Long: `Open an SSH connection to an environment.

Without arguments, opens an interactive shell. With a command after --,
executes that command on the remote environment.

Examples:
  sol ssh                           # Interactive shell
  sol ssh -p myproject -e main      # SSH to specific project/environment
  sol ssh -- ls -la                 # Run command on remote`,
	RunE:               runSSH,
	DisableFlagParsing: false,
}

func runSSH(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cfg, err := cli.FromCommand(cmd)
	if err != nil {
		return err
	}

	projectID := cfg.ProjectID
	if projectID == "" {
		projectID = detectProjectID()
		if projectID == "" {
			return errors.NewValidationError("no project specified").
				WithHint("Use --project or run from within a project directory")
		}
	}

	envID := cfg.Environment
	if envID == "" {
		envID = detectEnvironmentID()
		if envID == "" {
			return errors.NewValidationError("no environment specified").
				WithHint("Use --environment or run from within an environment")
		}
	}

	appName, _ := cmd.Flags().GetString("app")

	// Create API client
	client, err := api.New(ctx)
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	// Get environment to find SSH URL
	env, err := client.GetEnvironment(ctx, projectID, envID)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.StatusCode == 404 {
				return errors.NewNotFoundError("environment", envID)
			}
			return errors.NewAPIError(apiErr.Message, apiErr.StatusCode)
		}
		return errors.NewInternalError(fmt.Sprintf("get environment: %v", err))
	}

	// Get SSH URL from _links
	sshURL, ok := env.Links.GetHREF("ssh")
	if !ok {
		return errors.NewValidationError("environment has no SSH access").
			WithHint("The environment may be inactive or SSH access may be disabled")
	}

	// Modify SSH URL for specific app if requested
	if appName != "" {
		sshURL = modifySSHURLForApp(sshURL, appName)
	}

	// Parse and execute SSH
	sshArgs := parseSSHURL(sshURL)
	if len(args) > 0 {
		sshArgs = append(sshArgs, args...)
	}

	if !cfg.Quiet {
		fmt.Fprintf(os.Stderr, "Connecting to %s...\n", envID)
	}

	return execSSH(sshArgs)
}

// parseSSHURL parses an SSH URL into arguments for the ssh command.
// SSH URLs are in format: ssh://user@host:port or user@host
func parseSSHURL(sshURL string) []string {
	// Remove ssh:// prefix if present
	sshURL = strings.TrimPrefix(sshURL, "ssh://")

	var args []string

	// Check for port in URL (user@host:port)
	if idx := strings.LastIndex(sshURL, ":"); idx > strings.LastIndex(sshURL, "@") {
		host := sshURL[:idx]
		port := sshURL[idx+1:]
		args = append(args, "-p", port, host)
	} else {
		args = append(args, sshURL)
	}

	return args
}

// modifySSHURLForApp modifies the SSH URL to target a specific app.
// For multi-app projects, the app name is appended: user--app@host
func modifySSHURLForApp(sshURL, appName string) string {
	// Remove ssh:// prefix for manipulation
	sshURL = strings.TrimPrefix(sshURL, "ssh://")

	// Find the @ separator
	atIdx := strings.Index(sshURL, "@")
	if atIdx == -1 {
		return "ssh://" + sshURL
	}

	user := sshURL[:atIdx]
	host := sshURL[atIdx+1:]

	// Add app suffix to user
	return fmt.Sprintf("ssh://%s--%s@%s", user, appName, host)
}

// execSSH executes the ssh command with the given arguments.
// It replaces the current process with ssh.
func execSSH(args []string) error {
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return errors.NewInternalError("ssh command not found").
			WithHint("Ensure OpenSSH is installed and in your PATH")
	}

	// Execute ssh, replacing current process
	cmd := exec.Command(sshPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
