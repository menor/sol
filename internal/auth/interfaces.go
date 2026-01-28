package auth

import "context"

// TokenStore defines the interface for persisting OAuth tokens.
// Implementations include KeyringStore (OS keychain) and MemoryStore (testing).
type TokenStore interface {
	Save(token *StoredToken) error
	Load() (*StoredToken, error)
	Delete() error
	Exists() bool
}

// BrowserOpener defines the interface for opening URLs in a browser.
// This abstraction allows testing without launching actual browsers.
type BrowserOpener interface {
	Open(url string) error
}

// ProgressFunc is called to report progress during long-running operations.
// This allows the CLI to display status without hardcoding output in business logic.
type ProgressFunc func(message string)

// LoginOptions configures the login behavior.
type LoginOptions struct {
	// Force re-authentication even if already logged in
	Force bool
	// OnProgress is called with status messages during login
	OnProgress ProgressFunc
	// Timeout for waiting for OAuth callback (0 = default)
	Timeout context.Context
}

// AuthStatus represents the current authentication state.
type AuthStatus struct {
	Authenticated bool
	Method        string // "environment_variable", "keychain", or "none"
	Expired       bool
	ExpiresAt     string // RFC3339 format, empty if no expiry
	Hint          string // Helpful message for the user
}

// LoginResult contains the result of a successful login.
type LoginResult struct {
	ExpiresAt string // RFC3339 format
}
