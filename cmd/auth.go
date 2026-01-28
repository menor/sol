package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"lab.plat.farm/menor/sol/internal/auth"
	"lab.plat.farm/menor/sol/internal/cli"
	"lab.plat.farm/menor/sol/internal/errors"
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
	force, _ := cmd.Flags().GetBool("force")

	// Check if already logged in
	if !force && auth.HasToken() {
		token, err := auth.LoadToken()
		if err == nil && token != nil && !token.IsExpired() {
			fmt.Fprintln(os.Stderr, "Already logged in. Use --force to re-authenticate.")
			return nil
		}
	}

	fmt.Fprintln(os.Stderr, "Starting authentication...")

	// Start callback server
	server, redirectURL, resultChan, err := auth.StartCallbackServer(ctx)
	if err != nil {
		return errors.NewInternalError(fmt.Sprintf("start callback server: %v", err))
	}
	defer server.Shutdown(ctx)

	// Generate PKCE parameters
	pkce, err := auth.GeneratePKCE()
	if err != nil {
		return errors.NewInternalError(fmt.Sprintf("generate PKCE: %v", err))
	}

	// Generate state for CSRF protection
	state, err := auth.GenerateState()
	if err != nil {
		return errors.NewInternalError(fmt.Sprintf("generate state: %v", err))
	}

	// Build OAuth config and authorization URL
	oauthCfg := auth.OAuthConfig(redirectURL)
	authURL := auth.AuthorizationURL(oauthCfg, pkce, state)

	// Open browser
	// Note: We intentionally print the full URL as a fallback when browser
	// doesn't open. The URL contains a state parameter (for CSRF protection)
	// but no secrets - it's safe to display.
	fmt.Fprintln(os.Stderr, "Opening browser for authentication...")
	fmt.Fprintln(os.Stderr, "If the browser doesn't open, visit this URL:")
	fmt.Fprintln(os.Stderr, authURL)
	fmt.Fprintln(os.Stderr)

	if err := auth.OpenBrowser(authURL); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: couldn't open browser: %v\n", err)
	}

	// Wait for callback (with timeout)
	fmt.Fprintln(os.Stderr, "Waiting for authentication...")

	var result auth.CallbackResult
	select {
	case result = <-resultChan:
		// Got callback
	case <-time.After(5 * time.Minute):
		return errors.NewAuthError("authentication timed out after 5 minutes")
	case <-ctx.Done():
		return errors.NewAuthError("authentication cancelled")
	}

	// Check for errors from OAuth provider
	if result.Error != "" {
		return errors.NewAuthError(fmt.Sprintf("authentication failed: %s", result.Error))
	}

	// Validate state (CSRF protection)
	if result.State != state {
		return errors.NewAuthError("state mismatch - possible CSRF attack")
	}

	// Exchange code for tokens
	fmt.Fprintln(os.Stderr, "Exchanging authorization code for tokens...")

	token, err := auth.ExchangeCode(ctx, oauthCfg, result.Code, pkce)
	if err != nil {
		return errors.NewAuthError(fmt.Sprintf("token exchange failed: %v", err))
	}

	// Save tokens to keyring
	stored := auth.TokenToStored(token)
	if err := auth.SaveToken(stored); err != nil {
		return errors.NewInternalError(fmt.Sprintf("save token: %v", err))
	}

	fmt.Fprintln(os.Stderr, "Authentication successful!")

	// Output result using Config pattern (respects --output flag)
	cfg, err := cli.FromCommand(cmd)
	if err != nil {
		return err
	}
	return cfg.Formatter().Write(map[string]any{
		"status":     "authenticated",
		"expires_at": token.Expiry.Format(time.RFC3339),
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
	if !auth.HasToken() {
		fmt.Fprintln(os.Stderr, "Not currently logged in.")
		return nil
	}

	if err := auth.DeleteToken(); err != nil {
		return errors.NewInternalError(fmt.Sprintf("delete token: %v", err))
	}

	fmt.Fprintln(os.Stderr, "Logged out successfully.")

	cfg, err := cli.FromCommand(cmd)
	if err != nil {
		return err
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
	cfg, err := cli.FromCommand(cmd)
	if err != nil {
		return err
	}

	// Check for env var first (CI path)
	if envToken := os.Getenv("UPSUN_TOKEN"); envToken != "" {
		return cfg.Formatter().Write(map[string]any{
			"authenticated": true,
			"method":        "environment_variable",
			"variable":      "UPSUN_TOKEN",
		})
	}

	// Check keyring
	token, err := auth.LoadToken()
	if err != nil {
		return errors.NewInternalError(fmt.Sprintf("load token: %v", err))
	}

	if token == nil {
		return cfg.Formatter().Write(map[string]any{
			"authenticated": false,
			"hint":          "Run 'sol auth:login' to authenticate",
		})
	}

	info := map[string]any{
		"authenticated": true,
		"method":        "keychain",
		"expired":       token.IsExpired(),
	}

	if !token.Expiry.IsZero() {
		info["expires_at"] = token.Expiry.Format(time.RFC3339)
	}

	if token.IsExpired() {
		info["hint"] = "Token expired. Run 'sol auth:login' to re-authenticate"
	}

	return cfg.Formatter().Write(info)
}
