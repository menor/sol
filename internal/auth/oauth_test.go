package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestGeneratePKCE(t *testing.T) {
	pkce, err := GeneratePKCE()
	if err != nil {
		t.Fatalf("GeneratePKCE() error = %v", err)
	}

	// Verifier should be 43 characters (32 bytes base64url encoded without padding)
	if len(pkce.Verifier) != 43 {
		t.Errorf("Verifier length = %d, want 43", len(pkce.Verifier))
	}

	// Challenge should be 43 characters (32 bytes SHA256 hash, base64url encoded)
	if len(pkce.Challenge) != 43 {
		t.Errorf("Challenge length = %d, want 43", len(pkce.Challenge))
	}

	// Method should be S256
	if pkce.Method != "S256" {
		t.Errorf("Method = %q, want %q", pkce.Method, "S256")
	}

	// Verify challenge is SHA256(verifier)
	hash := sha256.Sum256([]byte(pkce.Verifier))
	expectedChallenge := base64.RawURLEncoding.EncodeToString(hash[:])
	if pkce.Challenge != expectedChallenge {
		t.Errorf("Challenge doesn't match SHA256(verifier)")
	}
}

func TestGeneratePKCEUniqueness(t *testing.T) {
	// Generate multiple PKCE params and verify they're unique
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		pkce, err := GeneratePKCE()
		if err != nil {
			t.Fatalf("GeneratePKCE() error = %v", err)
		}
		if seen[pkce.Verifier] {
			t.Errorf("Duplicate verifier generated")
		}
		seen[pkce.Verifier] = true
	}
}

func TestGenerateState(t *testing.T) {
	state, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState() error = %v", err)
	}

	// State should be 22 characters (16 bytes base64url encoded without padding)
	if len(state) != 22 {
		t.Errorf("State length = %d, want 22", len(state))
	}

	// Should be valid base64url
	if strings.ContainsAny(state, "+/=") {
		t.Error("State contains non-base64url characters")
	}
}

func TestGenerateStateUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		state, err := GenerateState()
		if err != nil {
			t.Fatalf("GenerateState() error = %v", err)
		}
		if seen[state] {
			t.Errorf("Duplicate state generated")
		}
		seen[state] = true
	}
}

func TestOAuthConfig(t *testing.T) {
	redirectURL := "http://127.0.0.1:12345/callback"
	cfg := OAuthConfig(redirectURL)

	if cfg.ClientID != ClientID {
		t.Errorf("ClientID = %q, want %q", cfg.ClientID, ClientID)
	}
	if cfg.Endpoint.AuthURL != AuthURL {
		t.Errorf("AuthURL = %q, want %q", cfg.Endpoint.AuthURL, AuthURL)
	}
	if cfg.Endpoint.TokenURL != TokenURL {
		t.Errorf("TokenURL = %q, want %q", cfg.Endpoint.TokenURL, TokenURL)
	}
	if cfg.RedirectURL != redirectURL {
		t.Errorf("RedirectURL = %q, want %q", cfg.RedirectURL, redirectURL)
	}
}

func TestAuthorizationURL(t *testing.T) {
	cfg := OAuthConfig("http://127.0.0.1:12345/callback")
	pkce := &PKCEParams{
		Verifier:  "test-verifier",
		Challenge: "test-challenge",
		Method:    "S256",
	}
	state := "test-state"

	url := AuthorizationURL(cfg, pkce, state)

	// Verify URL contains required parameters
	if !strings.Contains(url, "code_challenge=test-challenge") {
		t.Error("URL missing code_challenge")
	}
	if !strings.Contains(url, "code_challenge_method=S256") {
		t.Error("URL missing code_challenge_method")
	}
	if !strings.Contains(url, "state=test-state") {
		t.Error("URL missing state")
	}
	if !strings.Contains(url, "client_id="+ClientID) {
		t.Error("URL missing client_id")
	}
	if !strings.Contains(url, "redirect_uri=") {
		t.Error("URL missing redirect_uri")
	}
}

func TestExchangeAPIToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm() error = %v", err)
		}
		if got := r.PostForm.Get("grant_type"); got != "api_token" {
			t.Errorf("grant_type = %q, want %q", got, "api_token")
		}
		if got := r.PostForm.Get("api_token"); got != "my-api-token" {
			t.Errorf("api_token = %q, want %q", got, "my-api-token")
		}
		if got := r.PostForm.Get("client_id"); got != ClientID {
			t.Errorf("client_id = %q, want %q", got, ClientID)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"exchanged-token","token_type":"Bearer","expires_in":900}`)
	}))
	defer server.Close()

	token, err := ExchangeAPIToken(context.Background(), server.URL, ClientID, "my-api-token")
	if err != nil {
		t.Fatalf("ExchangeAPIToken() error = %v", err)
	}

	if token.AccessToken != "exchanged-token" {
		t.Errorf("AccessToken = %q, want %q", token.AccessToken, "exchanged-token")
	}
	if token.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want %q", token.TokenType, "Bearer")
	}

	// Expiry should be ~900s out.
	want := time.Now().Add(900 * time.Second)
	if diff := token.Expiry.Sub(want); diff < -10*time.Second || diff > 10*time.Second {
		t.Errorf("Expiry = %v, want ~%v", token.Expiry, want)
	}
}

func TestExchangeAPITokenDefaultExpiry(t *testing.T) {
	// expires_in is optional per RFC 6749. A missing value must not produce a
	// zero Expiry (which oauth2.Token.Valid() treats as "never expires").
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"access_token":"exchanged-token","token_type":"Bearer"}`)
	}))
	defer server.Close()

	token, err := ExchangeAPIToken(context.Background(), server.URL, ClientID, "my-api-token")
	if err != nil {
		t.Fatalf("ExchangeAPIToken() error = %v", err)
	}

	if token.Expiry.IsZero() {
		t.Fatal("Expiry is zero, want a default lifetime")
	}
	want := time.Now().Add(defaultExchangeExpiry)
	if diff := token.Expiry.Sub(want); diff < -10*time.Second || diff > 10*time.Second {
		t.Errorf("Expiry = %v, want ~%v", token.Expiry, want)
	}
}

func TestExchangeAPITokenErrors(t *testing.T) {
	tests := []struct {
		name         string
		status       int
		body         string
		wantSentinel error
		wantInMsg    string
	}{
		{
			name:         "400 with OAuth error body",
			status:       http.StatusBadRequest,
			body:         `{"error":"invalid_grant","error_description":"Invalid API token"}`,
			wantSentinel: ErrInvalidAPIToken,
			wantInMsg:    "invalid_grant: Invalid API token",
		},
		{
			name:         "401 without OAuth body",
			status:       http.StatusUnauthorized,
			body:         `nope`,
			wantSentinel: ErrInvalidAPIToken,
			wantInMsg:    "status 401",
		},
		{
			name:         "500 server error",
			status:       http.StatusInternalServerError,
			body:         `oops`,
			wantSentinel: ErrExchangeUnavailable,
			wantInMsg:    "status 500",
		},
		{
			name:         "200 with malformed JSON",
			status:       http.StatusOK,
			body:         `not json`,
			wantSentinel: ErrExchangeUnavailable,
			wantInMsg:    "parse response",
		},
		{
			name:         "200 missing access_token",
			status:       http.StatusOK,
			body:         `{"token_type":"Bearer","expires_in":900}`,
			wantSentinel: ErrExchangeUnavailable,
			wantInMsg:    "missing access_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				fmt.Fprint(w, tt.body)
			}))
			defer server.Close()

			_, err := ExchangeAPIToken(context.Background(), server.URL, ClientID, "my-api-token")
			if err == nil {
				t.Fatal("ExchangeAPIToken() error = nil, want error")
			}
			if !errors.Is(err, tt.wantSentinel) {
				t.Errorf("errors.Is(err, %v) = false, err = %v", tt.wantSentinel, err)
			}
			if !strings.Contains(err.Error(), tt.wantInMsg) {
				t.Errorf("error %q doesn't contain %q", err.Error(), tt.wantInMsg)
			}
			if strings.Contains(err.Error(), "my-api-token") {
				t.Errorf("error %q leaks the API token", err.Error())
			}
		})
	}
}

func TestExchangeAPITokenNetworkError(t *testing.T) {
	// A dead server must surface the *url.Error, not the auth sentinels:
	// the boundary classifies network failures via its url.Error branch.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close() // Immediately close so requests fail

	_, err := ExchangeAPIToken(context.Background(), server.URL, ClientID, "my-api-token")
	if err == nil {
		t.Fatal("ExchangeAPIToken() error = nil, want error")
	}

	var urlErr *url.Error
	if !errors.As(err, &urlErr) {
		t.Errorf("errors.As(err, *url.Error) = false, err = %v", err)
	}
	if errors.Is(err, ErrInvalidAPIToken) || errors.Is(err, ErrExchangeUnavailable) {
		t.Errorf("network error must not match auth sentinels, err = %v", err)
	}
}

func TestTokenConversion(t *testing.T) {
	stored := &StoredToken{
		AccessToken:  "access",
		RefreshToken: "refresh",
		TokenType:    "Bearer",
		Scope:        "read write",
	}

	// Convert to oauth2.Token
	token := StoredToToken(stored)
	if token.AccessToken != stored.AccessToken {
		t.Errorf("AccessToken mismatch")
	}
	if token.RefreshToken != stored.RefreshToken {
		t.Errorf("RefreshToken mismatch")
	}
	if token.TokenType != stored.TokenType {
		t.Errorf("TokenType mismatch")
	}

	// Convert back
	stored2 := TokenToStored(token)
	if stored2.AccessToken != stored.AccessToken {
		t.Errorf("Round-trip AccessToken mismatch")
	}
	if stored2.RefreshToken != stored.RefreshToken {
		t.Errorf("Round-trip RefreshToken mismatch")
	}
}
