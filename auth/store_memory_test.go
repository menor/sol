// Copyright 2026 José Menor
// Licensed under the Apache License, Version 2.0.
// See LICENSE and NOTICE files for details.

package auth

import (
	"testing"
	"time"
)

func TestMemoryStore_SaveAndLoad(t *testing.T) {
	s := NewMemoryStore()
	token := &StoredToken{
		AccessToken:  "access",
		RefreshToken: "refresh",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	if err := s.Save(token); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got == nil {
		t.Fatal("Load() returned nil, want token")
	}
	if got.AccessToken != token.AccessToken {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, token.AccessToken)
	}
	if got.RefreshToken != token.RefreshToken {
		t.Errorf("RefreshToken = %q, want %q", got.RefreshToken, token.RefreshToken)
	}
	if got.TokenType != token.TokenType {
		t.Errorf("TokenType = %q, want %q", got.TokenType, token.TokenType)
	}
}

func TestMemoryStore_LoadFromEmpty(t *testing.T) {
	s := NewMemoryStore()

	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got != nil {
		t.Errorf("Load() = %v, want nil", got)
	}
}

func TestMemoryStore_SaveIsACopy(t *testing.T) {
	s := NewMemoryStore()
	token := &StoredToken{AccessToken: "original"}

	s.Save(token)
	token.AccessToken = "mutated"

	got, _ := s.Load()
	if got.AccessToken != "original" {
		t.Errorf("Save() stored a reference — mutation affected stored value")
	}
}

func TestMemoryStore_LoadIsACopy(t *testing.T) {
	s := NewMemoryStore()
	s.Save(&StoredToken{AccessToken: "original"})

	got, _ := s.Load()
	got.AccessToken = "mutated"

	got2, _ := s.Load()
	if got2.AccessToken != "original" {
		t.Errorf("Load() returned a reference — mutation affected stored value")
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	s := NewMemoryStore()
	s.Save(&StoredToken{AccessToken: "access"})

	if err := s.Delete(); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load() after Delete() error = %v", err)
	}
	if got != nil {
		t.Error("Load() after Delete() returned token, want nil")
	}
}

func TestMemoryStore_DeleteOnEmpty(t *testing.T) {
	s := NewMemoryStore()

	if err := s.Delete(); err != nil {
		t.Errorf("Delete() on empty store error = %v, want nil", err)
	}
}

func TestMemoryStore_Exists(t *testing.T) {
	s := NewMemoryStore()

	if s.Exists() {
		t.Error("Exists() = true on empty store, want false")
	}

	s.Save(&StoredToken{AccessToken: "access"})

	if !s.Exists() {
		t.Error("Exists() = false after Save(), want true")
	}

	s.Delete()

	if s.Exists() {
		t.Error("Exists() = true after Delete(), want false")
	}
}
