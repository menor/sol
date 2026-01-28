package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"
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
