package auth

import (
	"context"
	"testing"
)

func TestEnvTokenSource(t *testing.T) {
	t.Setenv(EnvTokenVar, "test-api-token")

	ts, err := TokenSource(context.Background())
	if err != nil {
		t.Fatalf("TokenSource() error = %v", err)
	}

	token, err := ts.Token()
	if err != nil {
		t.Fatalf("Token() error = %v", err)
	}

	if token.AccessToken != "test-api-token" {
		t.Errorf("AccessToken = %q, want %q", token.AccessToken, "test-api-token")
	}
	if token.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want %q", token.TokenType, "Bearer")
	}
}

func TestTokenSourcePrecedence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping keyring test in short mode")
	}

	// Setup: save a token to keyring
	keyringToken := &StoredToken{
		AccessToken:  "keyring-token",
		RefreshToken: "refresh-token",
		TokenType:    "Bearer",
	}
	if err := SaveToken(keyringToken); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}
	defer DeleteToken()

	// Test: env var should take precedence over keyring
	t.Setenv(EnvTokenVar, "env-token")

	ts, err := TokenSource(context.Background())
	if err != nil {
		t.Fatalf("TokenSource() error = %v", err)
	}

	token, err := ts.Token()
	if err != nil {
		t.Fatalf("Token() error = %v", err)
	}

	// Should get env token, not keyring token
	if token.AccessToken != "env-token" {
		t.Errorf("AccessToken = %q, want %q (env var should take precedence)", token.AccessToken, "env-token")
	}
}

func TestIsAuthenticated(t *testing.T) {
	// Clean state
	DeleteToken()

	// Not authenticated
	if IsAuthenticated() {
		t.Error("IsAuthenticated() = true, want false (no credentials)")
	}

	// Authenticated via env var
	t.Setenv(EnvTokenVar, "test-token")
	if !IsAuthenticated() {
		t.Error("IsAuthenticated() = false, want true (env var set)")
	}
}

func TestAuthMethod(t *testing.T) {
	// Clean state
	DeleteToken()

	// No auth
	if got := AuthMethod(); got != "none" {
		t.Errorf("AuthMethod() = %q, want %q", got, "none")
	}

	// Env var auth
	t.Setenv(EnvTokenVar, "test-token")
	if got := AuthMethod(); got != "environment_variable" {
		t.Errorf("AuthMethod() = %q, want %q", got, "environment_variable")
	}
}

func TestAuthMethodKeychain(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping keyring test in short mode")
	}

	// Ensure env var is empty (not set to a value)
	// Note: t.Setenv("", "") sets to empty string, not unset - but our code
	// checks `!= ""` so empty string behaves the same as unset
	t.Setenv(EnvTokenVar, "")

	// Save token to keyring
	token := &StoredToken{
		AccessToken: "keyring-token",
		TokenType:   "Bearer",
	}
	if err := SaveToken(token); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}
	defer DeleteToken()

	if got := AuthMethod(); got != "keychain" {
		t.Errorf("AuthMethod() = %q, want %q", got, "keychain")
	}
}
