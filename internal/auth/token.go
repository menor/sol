package auth

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

// EnvTokenVar is the environment variable checked for API tokens.
// When set, it takes precedence over the keyring token.
// This is the preferred authentication method for CI/automated environments.
// The value is a Console API token, not an access token: it is exchanged at
// the auth server for a short-lived access token on first use.
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
		return &envTokenSource{
			ctx:      ctx,
			apiToken: envToken,
			tokenURL: TokenURL,
		}, nil
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

// exchangeRefreshBuffer is how long before expiry the cached access token is
// re-exchanged, so in-flight requests don't ride a token that dies mid-call.
const exchangeRefreshBuffer = 60 * time.Second

// envTokenSource exchanges the UPSUN_TOKEN API token for a short-lived
// access token and caches it in memory, re-exchanging near expiry.
// The exchanged token is never persisted: the env var path serves CI,
// where keyrings are absent.
//
// CONTEXT LIMITATION: This struct stores a context.Context, which is generally
// an anti-pattern in Go. However, this is a constraint of the oauth2.TokenSource
// interface, which doesn't accept context in its Token() method. If the stored
// context is cancelled before Token() is called, exchange operations will fail.
type envTokenSource struct {
	ctx      context.Context
	apiToken string
	tokenURL string // TokenURL in production; overridden in tests

	mu     sync.Mutex
	cached *oauth2.Token
}

func (s *envTokenSource) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cached != nil && time.Until(s.cached.Expiry) > exchangeRefreshBuffer {
		return s.cached, nil
	}

	token, err := ExchangeAPIToken(s.ctx, s.tokenURL, ClientID, s.apiToken)
	if err != nil {
		return nil, fmt.Errorf("exchange %s: %w", EnvTokenVar, err)
	}

	s.cached = token
	return token, nil
}

// keyringTokenSource provides tokens from the keyring with automatic refresh.
//
// CONTEXT LIMITATION: This struct stores a context.Context, which is generally
// an anti-pattern in Go. However, this is a constraint of the oauth2.TokenSource
// interface, which doesn't accept context in its Token() method. If the stored
// context is cancelled before Token() is called, refresh operations will fail.
// For long-lived token sources, consider creating new ones with fresh contexts.
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
		// Warn but don't fail - we have a valid token
		warnFunc("warning: couldn't save refreshed token to keyring: %v\n", err)
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

// getEnv is a wrapper around os.Getenv.
// It's a variable so it can be overridden in tests.
var getEnv = os.Getenv

// warnFunc is called to emit warning messages.
// It's a variable so it can be overridden in tests.
var warnFunc = func(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
}
