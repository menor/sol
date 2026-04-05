package cmd

import (
	"errors"

	"github.com/menor/sol/auth"
	clierrors "github.com/menor/sol/internal/errors"
)

// AuthLoginCmd logs in to Upsun.
type AuthLoginCmd struct {
	Force bool `help:"Force re-authentication even if already logged in" short:"f"`
}

// Run executes the auth:login command.
func (c *AuthLoginCmd) Run(ctx *Context) error {
	svc := auth.DefaultService()

	progress := func(msg string) {
		ctx.Log(msg)
	}

	result, err := svc.Login(ctx, auth.LoginOptions{
		Force:      c.Force,
		OnProgress: progress,
	})
	if err != nil {
		if errors.Is(err, auth.ErrAlreadyLoggedIn) {
			progress("Already logged in. Use --force to re-authenticate.")
			return nil
		}
		return clierrors.NewAuthError("authentication failed").WithDetail("cause", err.Error())
	}

	return ctx.Output(map[string]any{
		"status":     "authenticated",
		"expires_at": result.ExpiresAt,
	})
}

// AuthLogoutCmd logs out of Upsun.
type AuthLogoutCmd struct{}

// Run executes the auth:logout command.
func (c *AuthLogoutCmd) Run(ctx *Context) error {
	svc := auth.DefaultService()

	status, err := svc.Status(ctx)
	if err != nil {
		return clierrors.NewInternalError("check status: " + err.Error())
	}

	if !status.Authenticated || status.Method == "environment_variable" {
		ctx.Log("Not currently logged in via keychain.")
		return nil
	}

	if err := svc.Logout(ctx); err != nil {
		return clierrors.NewInternalError("logout: " + err.Error())
	}

	ctx.Log("Logged out successfully.")

	return ctx.Output(map[string]any{
		"status": "logged_out",
	})
}

// AuthInfoCmd shows authentication status.
type AuthInfoCmd struct{}

// Run executes the auth:info command.
func (c *AuthInfoCmd) Run(ctx *Context) error {
	svc := auth.DefaultService()

	status, err := svc.Status(ctx)
	if err != nil {
		return clierrors.NewInternalError("get status: " + err.Error())
	}

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

	return ctx.Output(info)
}
