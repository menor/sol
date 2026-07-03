package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

// exchangeServer returns an httptest server that answers the api_token grant,
// plus a counter of how many exchange requests it served.
func exchangeServer(t *testing.T, expiresIn int64) (*httptest.Server, *atomic.Int64) {
	t.Helper()
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"access_token":"exchanged-%d","token_type":"Bearer","expires_in":%d}`, n, expiresIn)
	}))
	t.Cleanup(server.Close)
	return server, &requests
}

func TestEnvTokenSourceExchanges(t *testing.T) {
	server, requests := exchangeServer(t, 900)

	ts := &envTokenSource{
		ctx:      context.Background(),
		apiToken: "test-api-token",
		tokenURL: server.URL,
	}

	token, err := ts.Token()
	if err != nil {
		t.Fatalf("Token() error = %v", err)
	}

	if token.AccessToken != "exchanged-1" {
		t.Errorf("AccessToken = %q, want %q (the exchanged token, not the API token)", token.AccessToken, "exchanged-1")
	}
	if token.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want %q", token.TokenType, "Bearer")
	}
	if got := requests.Load(); got != 1 {
		t.Errorf("exchange requests = %d, want 1", got)
	}
}

func TestEnvTokenSourceCachesToken(t *testing.T) {
	server, requests := exchangeServer(t, 900)

	ts := &envTokenSource{
		ctx:      context.Background(),
		apiToken: "test-api-token",
		tokenURL: server.URL,
	}

	first, err := ts.Token()
	if err != nil {
		t.Fatalf("first Token() error = %v", err)
	}
	second, err := ts.Token()
	if err != nil {
		t.Fatalf("second Token() error = %v", err)
	}

	if first.AccessToken != second.AccessToken {
		t.Errorf("second call returned %q, want cached %q", second.AccessToken, first.AccessToken)
	}
	if got := requests.Load(); got != 1 {
		t.Errorf("exchange requests = %d, want 1 (second call must hit the cache)", got)
	}
}

func TestEnvTokenSourceReExchangesAfterExpiry(t *testing.T) {
	// expires_in 30s is inside the 60s early-refresh buffer, so every call
	// treats the cached token as stale and re-exchanges.
	server, requests := exchangeServer(t, 30)

	ts := &envTokenSource{
		ctx:      context.Background(),
		apiToken: "test-api-token",
		tokenURL: server.URL,
	}

	if _, err := ts.Token(); err != nil {
		t.Fatalf("first Token() error = %v", err)
	}
	token, err := ts.Token()
	if err != nil {
		t.Fatalf("second Token() error = %v", err)
	}

	if token.AccessToken != "exchanged-2" {
		t.Errorf("AccessToken = %q, want %q (re-exchange, not cache)", token.AccessToken, "exchanged-2")
	}
	if got := requests.Load(); got != 2 {
		t.Errorf("exchange requests = %d, want 2", got)
	}
}

func TestEnvTokenSourceEarlyRefreshBuffer(t *testing.T) {
	server, requests := exchangeServer(t, 900)

	ts := &envTokenSource{
		ctx:      context.Background(),
		apiToken: "test-api-token",
		tokenURL: server.URL,
		// Valid for oauth2.Token.Valid(), but inside the 60s buffer.
		cached: &oauth2.Token{
			AccessToken: "stale-token",
			TokenType:   "Bearer",
			Expiry:      time.Now().Add(30 * time.Second),
		},
	}

	token, err := ts.Token()
	if err != nil {
		t.Fatalf("Token() error = %v", err)
	}

	if token.AccessToken == "stale-token" {
		t.Error("Token() returned the near-expiry token, want early re-exchange")
	}
	if got := requests.Load(); got != 1 {
		t.Errorf("exchange requests = %d, want 1", got)
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

	// Should get the env source, not the keyring source. Don't call Token():
	// it would perform a live exchange against the real auth server.
	envTS, ok := ts.(*envTokenSource)
	if !ok {
		t.Fatalf("TokenSource() = %T, want *envTokenSource (env var should take precedence)", ts)
	}
	if envTS.apiToken != "env-token" {
		t.Errorf("apiToken = %q, want %q", envTS.apiToken, "env-token")
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
