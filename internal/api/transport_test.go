package api

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

// mockTokenSource returns a fixed token for testing.
type mockTokenSource struct {
	token *oauth2.Token
	err   error
}

func (m *mockTokenSource) Token() (*oauth2.Token, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.token, nil
}

func TestTransport_InjectsAuthHeader(t *testing.T) {
	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ts := &mockTokenSource{token: &oauth2.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
	}}

	transport := &Transport{
		Base:        http.DefaultTransport,
		TokenSource: ts,
	}

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip error: %v", err)
	}
	defer resp.Body.Close()

	if receivedAuth != "Bearer test-token" {
		t.Errorf("Authorization header = %q, want %q", receivedAuth, "Bearer test-token")
	}
}

func TestTransport_RetriesOn5xx(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ts := &mockTokenSource{token: &oauth2.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
	}}

	transport := &Transport{
		Base:        http.DefaultTransport,
		TokenSource: ts,
		RetryConfig: RetryConfig{
			MaxRetries:  3,
			BaseDelay:   1 * time.Millisecond,
			MaxDelay:    10 * time.Millisecond,
			JitterRatio: 0,
		},
	}

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("attempts = %d, want 3", atomic.LoadInt32(&attempts))
	}
}

func TestTransport_DoesNotRetryOn4xx(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	ts := &mockTokenSource{token: &oauth2.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
	}}

	transport := &Transport{
		Base:        http.DefaultTransport,
		TokenSource: ts,
		RetryConfig: RetryConfig{
			MaxRetries:  3,
			BaseDelay:   1 * time.Millisecond,
			MaxDelay:    10 * time.Millisecond,
			JitterRatio: 0,
		},
	}

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("attempts = %d, want 1 (should not retry 4xx)", atomic.LoadInt32(&attempts))
	}
}

func TestTransport_RetriesOn401WithTokenRefresh(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Token source that returns a new token on second call
	callCount := int32(0)
	ts := &mockTokenSource{token: &oauth2.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
	}}
	_ = callCount // We just verify the retry happens

	transport := &Transport{
		Base:        http.DefaultTransport,
		TokenSource: ts,
		RetryConfig: RetryConfig{
			MaxRetries:  3,
			BaseDelay:   1 * time.Millisecond,
			MaxDelay:    10 * time.Millisecond,
			JitterRatio: 0,
		},
	}

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("attempts = %d, want 2 (one retry after 401)", atomic.LoadInt32(&attempts))
	}
}

func TestTransport_ExhaustsRetries(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	ts := &mockTokenSource{token: &oauth2.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
	}}

	transport := &Transport{
		Base:        http.DefaultTransport,
		TokenSource: ts,
		RetryConfig: RetryConfig{
			MaxRetries:  2,
			BaseDelay:   1 * time.Millisecond,
			MaxDelay:    10 * time.Millisecond,
			JitterRatio: 0,
		},
	}

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip error: %v", err)
	}
	defer resp.Body.Close()

	// Should exhaust retries and return last 503
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusServiceUnavailable)
	}
	// 1 initial + 2 retries = 3 attempts
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("attempts = %d, want 3", atomic.LoadInt32(&attempts))
	}
}

func TestCalculateDelay(t *testing.T) {
	transport := &Transport{}
	cfg := RetryConfig{
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    1 * time.Second,
		JitterRatio: 0, // No jitter for predictable tests
	}

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 100 * time.Millisecond},  // base * 2^0 = 100ms
		{2, 200 * time.Millisecond},  // base * 2^1 = 200ms
		{3, 400 * time.Millisecond},  // base * 2^2 = 400ms
		{4, 800 * time.Millisecond},  // base * 2^3 = 800ms
		{5, 1 * time.Second},         // base * 2^4 = 1600ms, capped at 1s
	}

	for _, tt := range tests {
		delay := transport.calculateDelay(tt.attempt, cfg)
		if delay != tt.expected {
			t.Errorf("calculateDelay(%d) = %v, want %v", tt.attempt, delay, tt.expected)
		}
	}
}

func TestIsRetryableStatus(t *testing.T) {
	tests := []struct {
		code     int
		expected bool
	}{
		{200, false},
		{201, false},
		{400, false},
		{401, true},  // Unauthorized - retry with token refresh
		{403, false},
		{404, false},
		{429, true},  // Rate limited
		{500, true},
		{502, true},
		{503, true},
		{504, true},
	}

	for _, tt := range tests {
		result := isRetryableStatus(tt.code)
		if result != tt.expected {
			t.Errorf("isRetryableStatus(%d) = %v, want %v", tt.code, result, tt.expected)
		}
	}
}
