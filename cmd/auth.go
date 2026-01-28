package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"lab.plat.farm/menor/sol/internal/auth"
	"lab.plat.farm/menor/sol/internal/cli"
	clierrors "lab.plat.farm/menor/sol/internal/errors"
)

func init() {
	// Add auth commands directly to root so "sol auth:login" works
	// (not "sol auth login" - we use colon convention like Upsun CLI)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(authInfoCmd)

	// Add --force flag to login command
	loginCmd.Flags().BoolP("force", "f", false, "Force re-authentication even if already logged in")
}

var loginCmd = &cobra.Command{
	Use:   "auth:login",
	Short: "Log in to Upsun",
	Long: `Authenticate with Upsun using OAuth2.

This command opens your browser to complete authentication.
After successful login, your credentials are stored securely
in the system keychain.

For CI/automated environments, use the UPSUN_TOKEN environment
variable instead of interactive login.`,
	RunE: runLogin,
}

func runLogin(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cfg, err := cli.FromCommand(cmd)
	if err != nil {
		return err
	}

	force, _ := cmd.Flags().GetBool("force")

	// Create service with default (production) dependencies
	svc := auth.DefaultService()

	// Progress callback prints to stderr (only if not quiet)
	progress := func(msg string) {
		if !cfg.Quiet {
			fmt.Fprintln(os.Stderr, msg)
		}
	}

	result, err := svc.Login(ctx, auth.LoginOptions{
		Force:      force,
		OnProgress: progress,
	})
	if err != nil {
		// Already logged in is not an error - just inform the user
		if errors.Is(err, auth.ErrAlreadyLoggedIn) {
			progress("Already logged in. Use --force to re-authenticate.")
			return nil
		}
		return clierrors.NewAuthError("authentication failed").WithDetail("cause", err.Error())
	}

	return cfg.Formatter().Write(map[string]any{
		"status":     "authenticated",
		"expires_at": result.ExpiresAt,
	})
}

var logoutCmd = &cobra.Command{
	Use:   "auth:logout",
	Short: "Log out of Upsun",
	Long: `Remove stored authentication credentials.

This deletes your access and refresh tokens from the system keychain.`,
	RunE: runLogout,
}

func runLogout(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cfg, err := cli.FromCommand(cmd)
	if err != nil {
		return err
	}

	svc := auth.DefaultService()

	// Check if logged in first to give appropriate message
	status, err := svc.Status(ctx)
	if err != nil {
		return clierrors.NewInternalError(fmt.Sprintf("check status: %v", err))
	}

	if !status.Authenticated || status.Method == "environment_variable" {
		if !cfg.Quiet {
			fmt.Fprintln(os.Stderr, "Not currently logged in via keychain.")
		}
		return nil
	}

	if err := svc.Logout(ctx); err != nil {
		return clierrors.NewInternalError(fmt.Sprintf("logout: %v", err))
	}

	if !cfg.Quiet {
		fmt.Fprintln(os.Stderr, "Logged out successfully.")
	}

	return cfg.Formatter().Write(map[string]any{
		"status": "logged_out",
	})
}

var authInfoCmd = &cobra.Command{
	Use:   "auth:info",
	Short: "Show authentication status",
	Long:  `Display information about current authentication status.`,
	RunE:  runAuthInfo,
}

func runAuthInfo(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cfg, err := cli.FromCommand(cmd)
	if err != nil {
		return err
	}

	svc := auth.DefaultService()

	status, err := svc.Status(ctx)
	if err != nil {
		return clierrors.NewInternalError(fmt.Sprintf("get status: %v", err))
	}

	// Build response map
	info := map[string]any{
		"authenticated": status.Authenticated,
	}

	if status.Method != "" && status.Method != "none" {
		info["method"] = status.Method
		if status.Method == "environment_variable" {
			info["variable"] = auth.EnvTokenVar
		}
	}

	if status.ExpiresAt != "" {
		info["expires_at"] = status.ExpiresAt
	}

	if status.Expired {
		info["expired"] = true
	}

	if status.Hint != "" {
		info["hint"] = status.Hint
	}

	return cfg.Formatter().Write(info)
}
