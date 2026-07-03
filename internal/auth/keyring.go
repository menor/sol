package auth

import (
	"encoding/json"
	"time"

	"github.com/zalando/go-keyring"
)

// ServiceName is the keyring service identifier for Sol.
// All tokens are stored under this service name in the OS keychain.
const ServiceName = "sol-cli"

// KeyringKey identifies different secrets stored in the keyring.
const (
	KeyToken = "oauth_token" // Stores the full OAuth token as JSON
)

// TokenExpiryBuffer is how early we consider a token expired.
// This avoids edge cases where a token expires mid-request.
const TokenExpiryBuffer = 30 * time.Second

// StoredToken represents an OAuth token persisted in the keyring.
// We store all token data as a single JSON blob rather than separate entries
// to keep the keyring tidy and ensure atomic updates.
type StoredToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"` // Usually "Bearer"
	Expiry       time.Time `json:"expiry"`     // When access token expires
	Scope        string    `json:"scope"`      // OAuth scopes granted
}

// IsExpired returns true if the access token has expired.
// We consider it expired TokenExpiryBuffer early to avoid edge cases.
func (t *StoredToken) IsExpired() bool {
	if t.Expiry.IsZero() {
		return false // No expiry means it doesn't expire
	}
	return time.Now().Add(TokenExpiryBuffer).After(t.Expiry)
}

// SaveToken stores an OAuth token in the OS keychain.
// The token is serialized to JSON and stored as a single entry.
func SaveToken(token *StoredToken) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return keyring.Set(ServiceName, KeyToken, string(data))
}

// LoadToken retrieves the OAuth token from the OS keychain.
// Returns nil if no token is stored (not an error - just means not logged in).
func LoadToken() (*StoredToken, error) {
	data, err := keyring.Get(ServiceName, KeyToken)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, nil // No token stored, not an error
		}
		return nil, err
	}

	var token StoredToken
	if err := json.Unmarshal([]byte(data), &token); err != nil {
		return nil, err
	}
	return &token, nil
}

// DeleteToken removes the OAuth token from the OS keychain.
// Used during logout. Returns nil if token doesn't exist.
func DeleteToken() error {
	err := keyring.Delete(ServiceName, KeyToken)
	if err == keyring.ErrNotFound {
		return nil // Already deleted, not an error
	}
	return err
}

// HasToken returns true if a token is stored in the keychain.
// This is a quick check without loading the full token.
func HasToken() bool {
	_, err := keyring.Get(ServiceName, KeyToken)
	return err == nil
}

// KeyringStore implements TokenStore using the OS keychain.
type KeyringStore struct{}

// Ensure KeyringStore implements TokenStore.
var _ TokenStore = (*KeyringStore)(nil)

// Save stores the token in the OS keychain.
func (s *KeyringStore) Save(token *StoredToken) error {
	return SaveToken(token)
}

// Load retrieves the token from the OS keychain.
func (s *KeyringStore) Load() (*StoredToken, error) {
	return LoadToken()
}

// Delete removes the token from the OS keychain.
func (s *KeyringStore) Delete() error {
	return DeleteToken()
}

// Exists returns true if a token exists in the keychain.
func (s *KeyringStore) Exists() bool {
	return HasToken()
}
