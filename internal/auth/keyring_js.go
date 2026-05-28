//go:build js

package auth

import (
	"time"

	"github.com/menor/sol/internal/errors"
)

// Constants kept in sync with the !js implementation so shared code compiles.
const ServiceName = "sol-cli"

const KeyToken = "oauth_token"

const TokenExpiryBuffer = 30 * time.Second

// StoredToken mirrors the !js definition so shared files (token.go, oauth.go) compile.
type StoredToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	Expiry       time.Time `json:"expiry"`
	Scope        string    `json:"scope"`
}

func (t *StoredToken) IsExpired() bool {
	if t.Expiry.IsZero() {
		return false
	}
	return time.Now().Add(TokenExpiryBuffer).After(t.Expiry)
}

// SaveToken is unsupported in browser builds.
func SaveToken(token *StoredToken) error {
	return errors.NewUnsupportedError("keyring not available in browser")
}

// LoadToken returns nil/nil in browser builds — semantically "no stored token".
// This lets shared code paths fall through to the env-token check without erroring.
func LoadToken() (*StoredToken, error) {
	return nil, nil
}

// DeleteToken is a no-op in browser builds.
func DeleteToken() error {
	return nil
}

// HasToken always returns false in browser builds.
func HasToken() bool {
	return false
}

// KeyringStore is a stub that errors on every write operation.
type KeyringStore struct{}

var _ TokenStore = (*KeyringStore)(nil)

func (s *KeyringStore) Save(token *StoredToken) error {
	return errors.NewUnsupportedError("keyring not available in browser")
}

func (s *KeyringStore) Load() (*StoredToken, error) {
	return nil, nil
}

func (s *KeyringStore) Delete() error {
	return nil
}

func (s *KeyringStore) Exists() bool {
	return false
}
