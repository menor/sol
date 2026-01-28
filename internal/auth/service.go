package auth

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// DefaultAuthTimeout is the default time to wait for OAuth callback.
const DefaultAuthTimeout = 5 * time.Minute

// Sentinel errors for authentication operations.
var (
	// ErrAlreadyLoggedIn is returned when login is attempted but user is already authenticated.
	ErrAlreadyLoggedIn = errors.New("already logged in")
)

// Service orchestrates authentication operations.
// It uses dependency injection for storage and browser, enabling testing.
type Service struct {
	store   TokenStore
	browser BrowserOpener
}

// NewService creates a new authentication service.
func NewService(store TokenStore, browser BrowserOpener) *Service {
	return &Service{
		store:   store,
		browser: browser,
	}
}

// DefaultService returns a Service configured with production dependencies.
func DefaultService() *Service {
	return NewService(&KeyringStore{}, &SystemBrowser{})
}

// Login performs OAuth2 + PKCE authentication flow.
// It opens a browser for the user to authenticate and waits for the callback.
func (s *Service) Login(ctx context.Context, opts LoginOptions) (*LoginResult, error) {
	progress := opts.OnProgress
	if progress == nil {
		progress = func(string) {} // No-op if not provided
	}

	// Check if already logged in
	if !opts.Force && s.store.Exists() {
		token, err := s.store.Load()
		if err == nil && token != nil && !token.IsExpired() {
			return nil, ErrAlreadyLoggedIn
		}
	}

	progress("Starting authentication...")

	// Start callback server
	server, redirectURL, resultChan, err := StartCallbackServer(ctx)
	if err != nil {
		return nil, fmt.Errorf("start callback server: %w", err)
	}
	defer server.Shutdown(ctx)

	// Generate PKCE parameters
	pkce, err := GeneratePKCE()
	if err != nil {
		return nil, fmt.Errorf("generate PKCE: %w", err)
	}

	// Generate state for CSRF protection
	state, err := GenerateState()
	if err != nil {
		return nil, fmt.Errorf("generate state: %w", err)
	}

	// Build OAuth config and authorization URL
	oauthCfg := OAuthConfig(redirectURL)
	authURL := AuthorizationURL(oauthCfg, pkce, state)

	// Open browser
	progress("Opening browser for authentication...")
	if err := s.browser.Open(authURL); err != nil {
		// Not fatal - user can manually open URL
		progress(fmt.Sprintf("Couldn't open browser: %v", err))
	}
	progress(fmt.Sprintf("If browser doesn't open, visit: %s", authURL))

	// Wait for callback
	progress("Waiting for authentication...")

	// Create context with timeout
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = DefaultAuthTimeout
	}
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var result CallbackResult
	select {
	case result = <-resultChan:
		// Got callback
	case <-waitCtx.Done():
		if ctx.Err() != nil {
			return nil, fmt.Errorf("authentication cancelled")
		}
		return nil, fmt.Errorf("authentication timed out")
	}

	// Check for errors from OAuth provider
	if result.Error != "" {
		return nil, fmt.Errorf("OAuth error: %s", result.Error)
	}

	// Validate state (CSRF protection)
	if result.State != state {
		return nil, fmt.Errorf("state mismatch - possible CSRF attack")
	}

	// Exchange code for tokens
	progress("Exchanging authorization code for tokens...")

	token, err := ExchangeCode(ctx, oauthCfg, result.Code, pkce)
	if err != nil {
		return nil, fmt.Errorf("token exchange: %w", err)
	}

	// Save tokens
	stored := TokenToStored(token)
	if err := s.store.Save(stored); err != nil {
		return nil, fmt.Errorf("save token: %w", err)
	}

	progress("Authentication successful!")

	return &LoginResult{
		ExpiresAt: token.Expiry.Format(time.RFC3339),
	}, nil
}

// Logout removes stored authentication credentials.
func (s *Service) Logout(ctx context.Context) error {
	if !s.store.Exists() {
		return nil // Already logged out
	}
	return s.store.Delete()
}

// Status returns the current authentication status.
func (s *Service) Status(ctx context.Context) (*AuthStatus, error) {
	status := &AuthStatus{}

	// Check for env var first (CI path)
	if envToken := getEnv(EnvTokenVar); envToken != "" {
		status.Authenticated = true
		status.Method = "environment_variable"
		return status, nil
	}

	// Check keyring
	token, err := s.store.Load()
	if err != nil {
		return nil, fmt.Errorf("load token: %w", err)
	}

	if token == nil {
		status.Authenticated = false
		status.Method = "none"
		status.Hint = "Run 'sol auth:login' to authenticate"
		return status, nil
	}

	status.Authenticated = true
	status.Method = "keychain"
	status.Expired = token.IsExpired()

	if !token.Expiry.IsZero() {
		status.ExpiresAt = token.Expiry.Format(time.RFC3339)
	}

	if token.IsExpired() {
		status.Hint = "Token expired. Run 'sol auth:login' to re-authenticate"
	}

	return status, nil
}
