package auth

import (
	"testing"
	"time"

	"github.com/zalando/go-keyring"
)

func TestStoredTokenIsExpired(t *testing.T) {
	tests := []struct {
		name   string
		expiry time.Time
		want   bool
	}{
		{"zero expiry (no expiration)", time.Time{}, false},
		{"future expiry", time.Now().Add(1 * time.Hour), false},
		{"past expiry", time.Now().Add(-1 * time.Hour), true},
		{"expiry within 30s buffer", time.Now().Add(15 * time.Second), true},
		{"expiry just outside buffer", time.Now().Add(45 * time.Second), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &StoredToken{Expiry: tt.expiry}
			if got := token.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSaveLoadDeleteToken(t *testing.T) {
	// This test uses the real keyring - skip in CI if needed
	if testing.Short() {
		t.Skip("skipping keyring test in short mode")
	}

	// Clean up before and after
	keyring.Delete(ServiceName, KeyToken)
	defer keyring.Delete(ServiceName, KeyToken)

	// Initially no token
	token, err := LoadToken()
	if err != nil {
		t.Fatalf("LoadToken() error = %v", err)
	}
	if token != nil {
		t.Error("LoadToken() should return nil when no token stored")
	}

	// HasToken should return false
	if HasToken() {
		t.Error("HasToken() should return false when no token stored")
	}

	// Save a token
	testToken := &StoredToken{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(1 * time.Hour),
		Scope:        "read write",
	}

	if err := SaveToken(testToken); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}

	// HasToken should return true
	if !HasToken() {
		t.Error("HasToken() should return true after saving token")
	}

	// Load the token back
	loaded, err := LoadToken()
	if err != nil {
		t.Fatalf("LoadToken() error = %v", err)
	}
	if loaded == nil {
		t.Fatal("LoadToken() returned nil after saving")
	}

	// Verify fields
	if loaded.AccessToken != testToken.AccessToken {
		t.Errorf("AccessToken = %q, want %q", loaded.AccessToken, testToken.AccessToken)
	}
	if loaded.RefreshToken != testToken.RefreshToken {
		t.Errorf("RefreshToken = %q, want %q", loaded.RefreshToken, testToken.RefreshToken)
	}
	if loaded.TokenType != testToken.TokenType {
		t.Errorf("TokenType = %q, want %q", loaded.TokenType, testToken.TokenType)
	}
	if loaded.Scope != testToken.Scope {
		t.Errorf("Scope = %q, want %q", loaded.Scope, testToken.Scope)
	}

	// Delete the token
	if err := DeleteToken(); err != nil {
		t.Fatalf("DeleteToken() error = %v", err)
	}

	// Verify it's gone
	token, err = LoadToken()
	if err != nil {
		t.Fatalf("LoadToken() after delete error = %v", err)
	}
	if token != nil {
		t.Error("LoadToken() should return nil after delete")
	}

	// Delete again should not error
	if err := DeleteToken(); err != nil {
		t.Errorf("DeleteToken() on non-existent token error = %v", err)
	}
}
