package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html"
	"net"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// OAuth2 configuration for Upsun
const (
	AuthURL  = "https://auth.upsun.com/oauth2/authorize"
	TokenURL = "https://auth.upsun.com/oauth2/token"

	// ClientID is the OAuth2 client identifier registered with Upsun's auth server.
	// We use "upsun-cli" which is already registered by Upsun/Platform.sh.
	// To use a different client ID, you would need to register it with Upsun's
	// OAuth server (auth.upsun.com) - this requires Upsun admin access.
	// For now, using the same client ID as the official CLI is fine since Sol
	// is just another CLI for the same platform.
	ClientID = "upsun-cli"

	RedirectURI = "http://127.0.0.1" // Port is dynamically assigned
)

// OAuthConfig returns the OAuth2 configuration for Upsun.
// The redirect URL includes the dynamically assigned port.
func OAuthConfig(redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:    ClientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:  AuthURL,
			TokenURL: TokenURL,
		},
		RedirectURL: redirectURL,
		Scopes:      []string{}, // Upsun uses default scopes
	}
}

// PKCEParams holds the PKCE code verifier and challenge.
// PKCE prevents authorization code interception attacks.
type PKCEParams struct {
	Verifier  string // Random secret (43-128 chars)
	Challenge string // SHA256 hash of verifier, base64url encoded
	Method    string // Always "S256"
}

// GeneratePKCE creates a new PKCE code verifier and challenge.
// The verifier is a random 32-byte string, base64url encoded (43 chars).
// The challenge is SHA256(verifier), base64url encoded.
func GeneratePKCE() (*PKCEParams, error) {
	// Generate 32 random bytes for the verifier
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return nil, fmt.Errorf("generate verifier: %w", err)
	}

	// Base64url encode without padding (per RFC 7636)
	verifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	// SHA256 hash the verifier for the challenge
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return &PKCEParams{
		Verifier:  verifier,
		Challenge: challenge,
		Method:    "S256",
	}, nil
}

// AuthorizationURL builds the URL to open in the user's browser.
// It includes the PKCE challenge and a state parameter for CSRF protection.
func AuthorizationURL(cfg *oauth2.Config, pkce *PKCEParams, state string) string {
	return cfg.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("code_challenge", pkce.Challenge),
		oauth2.SetAuthURLParam("code_challenge_method", pkce.Method),
	)
}

// CallbackResult is returned after the OAuth callback is received.
type CallbackResult struct {
	Code  string // Authorization code to exchange for tokens
	State string // State parameter for CSRF validation
	Error string // Error from OAuth provider (if any)
}

// StartCallbackServer starts a local HTTP server to receive the OAuth callback.
// It returns the server, the URL to use as redirect_uri, and a channel that
// receives the callback result.
//
// The server listens on a random available port on 127.0.0.1.
// The caller must call server.Shutdown() after receiving the result.
func StartCallbackServer(ctx context.Context) (*http.Server, string, <-chan CallbackResult, error) {
	// Find an available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, "", nil, fmt.Errorf("listen: %w", err)
	}

	// Get the assigned port
	//
	// REDIRECT URI FORMAT:
	// We use "http://127.0.0.1:PORT" without a path because that's what's
	// pre-registered for the "upsun-cli" OAuth client on auth.upsun.com.
	//
	// Ideally we'd use "http://127.0.0.1:PORT/callback" for clarity (explicit
	// callback path), but the OAuth server rejects URIs that don't match its
	// allowlist exactly. If Sol ever gets its own registered client ID, we
	// should switch to the /callback path convention.
	addr := listener.Addr().(*net.TCPAddr)
	redirectURL := fmt.Sprintf("http://127.0.0.1:%d", addr.Port)

	resultChan := make(chan CallbackResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		result := CallbackResult{
			Code:  query.Get("code"),
			State: query.Get("state"),
			Error: query.Get("error"),
		}

		// Send a nice response to the browser
		if result.Error != "" {
			w.WriteHeader(http.StatusBadRequest)
			// Escape error to prevent HTML injection (low risk since local callback)
			fmt.Fprintf(w, "<html><body><h1>Authentication Failed</h1><p>%s</p><p>You can close this window.</p></body></html>", html.EscapeString(result.Error))
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "<html><body><h1>Authentication Successful</h1><p>You can close this window and return to the terminal.</p></body></html>")
		}

		// Send result to channel (non-blocking)
		// Only send if we have an actual OAuth response (code or error)
		// This prevents favicon.ico or other browser requests from racing
		if result.Code != "" || result.Error != "" {
			select {
			case resultChan <- result:
			default:
			}
		}
	})

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Start serving in background
	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			// Log error but don't crash - the main goroutine will timeout
		}
	}()

	return server, redirectURL, resultChan, nil
}

// ExchangeCode exchanges an authorization code for tokens.
// The PKCE verifier must match the challenge sent during authorization.
func ExchangeCode(ctx context.Context, cfg *oauth2.Config, code string, pkce *PKCEParams) (*oauth2.Token, error) {
	return cfg.Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("code_verifier", pkce.Verifier),
	)
}

// GenerateState creates a random state parameter for CSRF protection.
// The state is sent to the OAuth provider and must match on callback.
func GenerateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// TokenToStored converts an oauth2.Token to our StoredToken format.
func TokenToStored(token *oauth2.Token) *StoredToken {
	return &StoredToken{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	}
}

// StoredToToken converts our StoredToken to an oauth2.Token.
func StoredToToken(stored *StoredToken) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  stored.AccessToken,
		RefreshToken: stored.RefreshToken,
		TokenType:    stored.TokenType,
		Expiry:       stored.Expiry,
	}
}

// OpenBrowser opens a URL in the user's default browser.
// This is a placeholder - the actual implementation depends on OS.
func OpenBrowser(url string) error {
	// We'll implement this in a separate file with build tags for each OS
	return openBrowser(url)
}

// RefreshToken uses the refresh token to get a new access token.
func RefreshToken(ctx context.Context, cfg *oauth2.Config, refreshToken string) (*oauth2.Token, error) {
	// Create a token source that will refresh using the refresh token
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	// TokenSource will automatically refresh when Token() is called
	// because the access token is empty/expired
	ts := cfg.TokenSource(ctx, token)
	return ts.Token()
}

