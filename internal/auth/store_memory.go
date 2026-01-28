package auth

import "sync"

// MemoryStore implements TokenStore using in-memory storage.
// This is useful for testing without requiring OS keychain access.
type MemoryStore struct {
	mu    sync.RWMutex
	token *StoredToken
}

// Ensure MemoryStore implements TokenStore.
var _ TokenStore = (*MemoryStore)(nil)

// NewMemoryStore creates a new in-memory token store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

// Save stores the token in memory.
func (s *MemoryStore) Save(token *StoredToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Deep copy to avoid external modifications
	copied := *token
	s.token = &copied
	return nil
}

// Load retrieves the token from memory.
func (s *MemoryStore) Load() (*StoredToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.token == nil {
		return nil, nil
	}
	// Return a copy to avoid external modifications
	copied := *s.token
	return &copied, nil
}

// Delete removes the token from memory.
func (s *MemoryStore) Delete() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.token = nil
	return nil
}

// Exists returns true if a token exists in memory.
func (s *MemoryStore) Exists() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.token != nil
}
