// Copyright 2026 José Menor
// Licensed under the Apache License, Version 2.0.
// See LICENSE and NOTICE files for details.

package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockBrowser records Open calls and can simulate failure.
type mockBrowser struct {
	opened []string
	err    error
}

func (b *mockBrowser) Open(url string) error {
	b.opened = append(b.opened, url)
	return b.err
}

// validToken returns a StoredToken that is not expired.
func validToken() *StoredToken {
	return &StoredToken{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(1 * time.Hour),
	}
}

// expiredToken returns a StoredToken that is already expired.
func expiredToken() *StoredToken {
	return &StoredToken{
		AccessToken:  "old-token",
		RefreshToken: "refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(-1 * time.Hour),
	}
}

// --- Login ---

func TestLogin_AlreadyLoggedIn(t *testing.T) {
	store := NewMemoryStore()
	store.Save(validToken())

	svc := NewService(store, &mockBrowser{})
	_, err := svc.Login(context.Background(), LoginOptions{})

	if !errors.Is(err, ErrAlreadyLoggedIn) {
		t.Errorf("err = %v, want ErrAlreadyLoggedIn", err)
	}
}

func TestLogin_AlreadyLoggedIn_ForceProceeds(t *testing.T) {
	store := NewMemoryStore()
	store.Save(validToken())

	svc := NewService(store, &mockBrowser{})

	// With Force: true the guard is skipped; the flow will time out since
	// no real OAuth callback arrives. That's the expected behaviour here.
	_, err := svc.Login(context.Background(), LoginOptions{
		Force:   true,
		Timeout: 50 * time.Millisecond,
	})

	if errors.Is(err, ErrAlreadyLoggedIn) {
		t.Error("Force: true should bypass the already-logged-in guard")
	}
}

func TestLogin_ExpiredToken_NotBlocked(t *testing.T) {
	store := NewMemoryStore()
	store.Save(expiredToken())

	svc := NewService(store, &mockBrowser{})

	_, err := svc.Login(context.Background(), LoginOptions{
		Timeout: 50 * time.Millisecond,
	})

	if errors.Is(err, ErrAlreadyLoggedIn) {
		t.Error("expired token should not block login")
	}
}

func TestLogin_Timeout(t *testing.T) {
	svc := NewService(NewMemoryStore(), &mockBrowser{})

	_, err := svc.Login(context.Background(), LoginOptions{
		Timeout: 50 * time.Millisecond,
	})

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if err.Error() != "authentication timed out" {
		t.Errorf("err = %q, want %q", err.Error(), "authentication timed out")
	}
}

func TestLogin_ContextCancelled(t *testing.T) {
	svc := NewService(NewMemoryStore(), &mockBrowser{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before login starts

	_, err := svc.Login(ctx, LoginOptions{
		Timeout: 5 * time.Second,
	})

	if err == nil {
		t.Fatal("expected cancellation error, got nil")
	}
	if err.Error() != "authentication cancelled" {
		t.Errorf("err = %q, want %q", err.Error(), "authentication cancelled")
	}
}

func TestLogin_BrowserFailureNonFatal(t *testing.T) {
	browser := &mockBrowser{err: errors.New("no browser available")}
	svc := NewService(NewMemoryStore(), browser)

	// Browser failure is non-fatal: flow continues and times out normally.
	_, err := svc.Login(context.Background(), LoginOptions{
		Timeout: 50 * time.Millisecond,
	})

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if err.Error() != "authentication timed out" {
		t.Errorf("browser failure should not change error type, got %q", err.Error())
	}
	if len(browser.opened) != 1 {
		t.Errorf("Open() called %d times, want 1", len(browser.opened))
	}
}

func TestLogin_ProgressCallbackCalled(t *testing.T) {
	svc := NewService(NewMemoryStore(), &mockBrowser{})

	var messages []string
	_, _ = svc.Login(context.Background(), LoginOptions{
		Timeout: 50 * time.Millisecond,
		OnProgress: func(msg string) {
			messages = append(messages, msg)
		},
	})

	if len(messages) == 0 {
		t.Error("OnProgress was never called")
	}
}

// --- Logout ---

func TestLogout_DeletesToken(t *testing.T) {
	store := NewMemoryStore()
	store.Save(validToken())

	svc := NewService(store, &mockBrowser{})
	if err := svc.Logout(context.Background()); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	if store.Exists() {
		t.Error("token still exists after Logout()")
	}
}

func TestLogout_AlreadyLoggedOut(t *testing.T) {
	svc := NewService(NewMemoryStore(), &mockBrowser{})

	if err := svc.Logout(context.Background()); err != nil {
		t.Errorf("Logout() on empty store error = %v, want nil", err)
	}
}

// --- Status ---

func TestStatus_EnvVar(t *testing.T) {
	orig := getEnv
	getEnv = func(key string) string {
		if key == EnvTokenVar {
			return "env-api-token"
		}
		return ""
	}
	defer func() { getEnv = orig }()

	svc := NewService(NewMemoryStore(), &mockBrowser{})
	status, err := svc.Status(context.Background())

	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if !status.Authenticated {
		t.Error("Authenticated = false, want true")
	}
	if status.Method != "environment_variable" {
		t.Errorf("Method = %q, want %q", status.Method, "environment_variable")
	}
}

func TestStatus_NoToken(t *testing.T) {
	orig := getEnv
	getEnv = func(string) string { return "" }
	defer func() { getEnv = orig }()

	svc := NewService(NewMemoryStore(), &mockBrowser{})
	status, err := svc.Status(context.Background())

	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.Authenticated {
		t.Error("Authenticated = true, want false")
	}
	if status.Method != "none" {
		t.Errorf("Method = %q, want %q", status.Method, "none")
	}
	if status.Hint == "" {
		t.Error("Hint is empty, want a hint message")
	}
}

func TestStatus_ValidKeychainToken(t *testing.T) {
	orig := getEnv
	getEnv = func(string) string { return "" }
	defer func() { getEnv = orig }()

	store := NewMemoryStore()
	store.Save(validToken())

	svc := NewService(store, &mockBrowser{})
	status, err := svc.Status(context.Background())

	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if !status.Authenticated {
		t.Error("Authenticated = false, want true")
	}
	if status.Method != "keychain" {
		t.Errorf("Method = %q, want %q", status.Method, "keychain")
	}
	if status.Expired {
		t.Error("Expired = true, want false")
	}
	if status.ExpiresAt == "" {
		t.Error("ExpiresAt is empty, want RFC3339 timestamp")
	}
}

func TestStatus_ExpiredKeychainToken(t *testing.T) {
	orig := getEnv
	getEnv = func(string) string { return "" }
	defer func() { getEnv = orig }()

	store := NewMemoryStore()
	store.Save(expiredToken())

	svc := NewService(store, &mockBrowser{})
	status, err := svc.Status(context.Background())

	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if !status.Authenticated {
		t.Error("Authenticated = false, want true for expired token")
	}
	if !status.Expired {
		t.Error("Expired = false, want true")
	}
	if status.Hint == "" {
		t.Error("Hint is empty, want re-authentication hint")
	}
}

func TestStatus_StoreError(t *testing.T) {
	orig := getEnv
	getEnv = func(string) string { return "" }
	defer func() { getEnv = orig }()

	svc := NewService(&errorStore{}, &mockBrowser{})
	_, err := svc.Status(context.Background())

	if err == nil {
		t.Error("expected error from store, got nil")
	}
}

// errorStore is a TokenStore whose Load always fails.
type errorStore struct{}

func (s *errorStore) Save(*StoredToken) error  { return nil }
func (s *errorStore) Load() (*StoredToken, error) {
	return nil, errors.New("keyring unavailable")
}
func (s *errorStore) Delete() error { return nil }
func (s *errorStore) Exists() bool  { return false }
