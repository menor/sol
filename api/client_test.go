// Copyright 2026 José Menor
// Licensed under the Apache License, Version 2.0.
// See LICENSE and NOTICE files for details.

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/oauth2"
)

func TestClient_Get(t *testing.T) {
	expected := map[string]string{"message": "hello"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Method = %s, want %s", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/v1/test" {
			t.Errorf("Path = %s, want /v1/test", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	ts := &mockTokenSource{token: &oauth2.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
	}}

	client, err := New(context.Background(),
		WithBaseURL(server.URL),
		WithTokenSource(ts),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	var result map[string]string
	err = client.Get(context.Background(), "/v1/test", &result)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	if result["message"] != expected["message"] {
		t.Errorf("result = %v, want %v", result, expected)
	}
}

func TestClient_Post(t *testing.T) {
	requestBody := map[string]string{"name": "test"}
	responseBody := map[string]string{"id": "123"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %s, want %s", r.Method, http.MethodPost)
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != requestBody["name"] {
			t.Errorf("request body = %v, want %v", body, requestBody)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responseBody)
	}))
	defer server.Close()

	ts := &mockTokenSource{token: &oauth2.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
	}}

	client, err := New(context.Background(),
		WithBaseURL(server.URL),
		WithTokenSource(ts),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	var result map[string]string
	err = client.Post(context.Background(), "/v1/projects", requestBody, &result)
	if err != nil {
		t.Fatalf("Post() error: %v", err)
	}

	if result["id"] != responseBody["id"] {
		t.Errorf("result = %v, want %v", result, responseBody)
	}
}

func TestClient_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Project not found",
		})
	}))
	defer server.Close()

	ts := &mockTokenSource{token: &oauth2.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
	}}

	client, err := New(context.Background(),
		WithBaseURL(server.URL),
		WithTokenSource(ts),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	var result map[string]string
	err = client.Get(context.Background(), "/v1/projects/invalid", &result)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusNotFound)
	}
	if apiErr.Message != "Project not found" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Project not found")
	}
}

func TestClient_Delete(t *testing.T) {
	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Method = %s, want %s", r.Method, http.MethodDelete)
		}
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	ts := &mockTokenSource{token: &oauth2.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
	}}

	client, err := New(context.Background(),
		WithBaseURL(server.URL),
		WithTokenSource(ts),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	err = client.Delete(context.Background(), "/v1/projects/123")
	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	if !called {
		t.Error("expected DELETE request to be made")
	}
}

func TestClient_InvalidBaseURL(t *testing.T) {
	ts := &mockTokenSource{token: &oauth2.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
	}}

	_, err := New(context.Background(),
		WithBaseURL("not-a-url"),
		WithTokenSource(ts),
	)
	if err == nil {
		t.Fatal("expected error for invalid base URL")
	}
}

func TestClient_SetsHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/json")
		}
		if accept := r.Header.Get("Accept"); accept != "application/json" {
			t.Errorf("Accept = %q, want %q", accept, "application/json")
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ts := &mockTokenSource{token: &oauth2.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
	}}

	client, err := New(context.Background(),
		WithBaseURL(server.URL),
		WithTokenSource(ts),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	err = client.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
}

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		StatusCode: 404,
		Message:    "Not found",
	}

	expected := "API error 404: Not found"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}

func TestParseAPIError_JSONError(t *testing.T) {
	body := []byte(`{"error": "Something went wrong"}`)
	err := parseAPIError(500, body)

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Message != "Something went wrong" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Something went wrong")
	}
}

func TestParseAPIError_JSONMessage(t *testing.T) {
	body := []byte(`{"message": "Invalid request"}`)
	err := parseAPIError(400, body)

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Message != "Invalid request" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Invalid request")
	}
}

func TestParseAPIError_FallbackToStatusText(t *testing.T) {
	body := []byte(`not json`)
	err := parseAPIError(500, body)

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Message != "Internal Server Error" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Internal Server Error")
	}
}
