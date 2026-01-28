package auth

import (
	"context"
	"fmt"
	"os"
	"sync"

	"golang.org/x/oauth2"
)

// EnvTokenVar is the environment variable checked for API tokens.
// When set, it takes precedence over the keyring token.
// This is the preferred authentication method for CI/automated environments.
const EnvTokenVar = "UPSUN_TOKEN"

// TokenSource returns an oauth2.TokenSource that provides access tokens.
// It checks sources in this order:
//  1. UPSUN_TOKEN environment variable (for CI/automated use)
//  2. Keyring-stored OAuth token (for interactive use)
//
// If using the keyring token and it's expired, the refresh token is used
// to obtain a new access token automatically.
func TokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	// Check environment variable first (CI path)
	if envToken := os.Getenv(EnvTokenVar); envToken != "" {
		return &envTokenSource{token: envToken}, nil
	}

	// Fall back to keyring (interactive path)
	stored, err := LoadToken()
	if err != nil {
		return nil, fmt.Errorf("load token from keyring: %w", err)
	}
	if stored == nil {
		return nil, fmt.Errorf("not authenticated: run 'sol auth:login' or set %s", EnvTokenVar)
	}

	return &keyringTokenSource{
		ctx:    ctx,
		stored: stored,
	}, nil
}

// envTokenSource provides tokens from the UPSUN_TOKEN environment variable.
// These tokens don't expire (they're API tokens, not OAuth tokens).
type envTokenSource struct {
	token string
}

func (s *envTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: s.token,
		TokenType:   "Bearer",
		// No expiry - API tokens don't expire
	}, nil
}

// keyringTokenSource provides tokens from the keyring with automatic refresh.
type keyringTokenSource struct {
	ctx    context.Context
	stored *StoredToken
	mu     sync.Mutex
}

func (s *keyringTokenSource) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If token is still valid, return it
	if !s.stored.IsExpired() {
		return StoredToToken(s.stored), nil
	}

	// Token expired - try to refresh
	if s.stored.RefreshToken == "" {
		return nil, fmt.Errorf("token expired and no refresh token available: run 'sol auth:login'")
	}

	// Refresh the token
	cfg := OAuthConfig("") // Redirect URL not needed for refresh
	newToken, err := RefreshToken(s.ctx, cfg, s.stored.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh token: %w (run 'sol auth:login' to re-authenticate)", err)
	}

	// Update stored token, preserving refresh token if not returned
	oldRefreshToken := s.stored.RefreshToken
	s.stored = TokenToStored(newToken)
	if s.stored.RefreshToken == "" {
		s.stored.RefreshToken = oldRefreshToken
	}

	// Persist to keyring
	if err := SaveToken(s.stored); err != nil {
		// Log but don't fail - we have a valid token
		fmt.Fprintf(os.Stderr, "warning: couldn't save refreshed token to keyring: %v\n", err)
	}

	return newToken, nil
}

// GetToken is a convenience function that returns the current access token string.
// It handles token refresh automatically if needed.
func GetToken(ctx context.Context) (string, error) {
	ts, err := TokenSource(ctx)
	if err != nil {
		return "", err
	}

	token, err := ts.Token()
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

// IsAuthenticated returns true if authentication credentials are available.
// It checks both the environment variable and the keyring.
func IsAuthenticated() bool {
	if os.Getenv(EnvTokenVar) != "" {
		return true
	}
	return HasToken()
}

// AuthMethod returns the current authentication method.
// Returns "environment_variable", "keychain", or "none".
func AuthMethod() string {
	if os.Getenv(EnvTokenVar) != "" {
		return "environment_variable"
	}
	if HasToken() {
		return "keychain"
	}
	return "none"
}
