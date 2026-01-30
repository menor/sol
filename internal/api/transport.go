package api

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// RetryConfig controls retry behavior for transient failures.
type RetryConfig struct {
	MaxRetries  int           // Maximum number of retry attempts (default: 3)
	BaseDelay   time.Duration // Initial delay before first retry (default: 100ms)
	MaxDelay    time.Duration // Maximum delay between retries (default: 5s)
	JitterRatio float64       // Jitter as a ratio of delay (default: 0.2 = 20%)
}

// DefaultRetryConfig provides sensible defaults for API retries.
var DefaultRetryConfig = RetryConfig{
	MaxRetries:  3,
	BaseDelay:   100 * time.Millisecond,
	MaxDelay:    5 * time.Second,
	JitterRatio: 0.2,
}

// Transport is an http.RoundTripper that:
// - Injects Authorization headers from an oauth2.TokenSource
// - Retries requests on transient failures (5xx, network errors)
// - Refreshes tokens on 401 Unauthorized and retries once
type Transport struct {
	// Base transport (usually http.DefaultTransport)
	Base http.RoundTripper

	// TokenSource provides access tokens for requests
	TokenSource oauth2.TokenSource

	// RetryConfig controls retry behavior
	RetryConfig RetryConfig

	// LogFunc is called for retry/refresh events (optional)
	LogFunc func(format string, args ...any)
}

// RoundTrip implements http.RoundTripper.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Get token and add to request
	token, err := t.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("get access token: %w", err)
	}

	// Clone request and set auth header
	req = cloneRequest(req)
	req.Header.Set("Authorization", token.Type()+" "+token.AccessToken)

	// Attempt with retries
	cfg := t.retryConfig()
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := t.calculateDelay(attempt, cfg)
			t.log("Retrying request (attempt %d/%d) after %v", attempt, cfg.MaxRetries, delay)
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-time.After(delay):
			}
			// Re-clone for retry (body needs to be re-readable)
			req = cloneRequest(req)
			req.Header.Set("Authorization", token.Type()+" "+token.AccessToken)
		}

		resp, lastErr = t.base().RoundTrip(req)

		// Success or non-retryable error
		if lastErr == nil && !isRetryableStatus(resp.StatusCode) {
			break
		}

		// Network error - retry
		if lastErr != nil {
			t.log("Request failed: %v", lastErr)
			continue
		}

		// 401 Unauthorized - try token refresh once
		if resp.StatusCode == http.StatusUnauthorized && attempt == 0 {
			t.log("Received 401, attempting token refresh")
			drainBody(resp.Body)

			// Force new token by getting it again
			// Note: The TokenSource handles refresh internally
			token, err = t.TokenSource.Token()
			if err != nil {
				return nil, fmt.Errorf("refresh access token: %w", err)
			}
			continue
		}

		// 5xx - retry with backoff
		if resp.StatusCode >= 500 {
			t.log("Received %d, will retry", resp.StatusCode)
			drainBody(resp.Body)
			continue
		}

		// 4xx (non-401) - don't retry
		break
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return resp, nil
}

func (t *Transport) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}
	return http.DefaultTransport
}

func (t *Transport) retryConfig() RetryConfig {
	cfg := t.RetryConfig
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = DefaultRetryConfig.MaxRetries
	}
	if cfg.BaseDelay == 0 {
		cfg.BaseDelay = DefaultRetryConfig.BaseDelay
	}
	if cfg.MaxDelay == 0 {
		cfg.MaxDelay = DefaultRetryConfig.MaxDelay
	}
	if cfg.JitterRatio == 0 {
		cfg.JitterRatio = DefaultRetryConfig.JitterRatio
	}
	return cfg
}

func (t *Transport) calculateDelay(attempt int, cfg RetryConfig) time.Duration {
	// Exponential backoff: base * 2^attempt
	delay := float64(cfg.BaseDelay) * math.Pow(2, float64(attempt-1))

	// Cap at max delay
	if delay > float64(cfg.MaxDelay) {
		delay = float64(cfg.MaxDelay)
	}

	// Add jitter
	jitter := delay * cfg.JitterRatio * (rand.Float64()*2 - 1) // -jitter to +jitter
	delay += jitter

	return time.Duration(delay)
}

func (t *Transport) log(format string, args ...any) {
	if t.LogFunc != nil {
		t.LogFunc(format, args...)
	}
}

// isRetryableStatus returns true for status codes that warrant a retry.
func isRetryableStatus(code int) bool {
	return code == http.StatusUnauthorized || // 401 - will try token refresh
		code == http.StatusTooManyRequests || // 429 - rate limited
		code >= 500 // 5xx - server errors
}

// cloneRequest creates a shallow copy of a request with a re-readable body.
func cloneRequest(req *http.Request) *http.Request {
	clone := req.Clone(req.Context())

	if req.Body != nil && req.Body != http.NoBody {
		bodyBytes, _ := io.ReadAll(req.Body)
		req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		clone.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	return clone
}

// drainBody reads and closes a response body to allow connection reuse.
func drainBody(body io.ReadCloser) {
	if body != nil {
		io.Copy(io.Discard, body)
		body.Close()
	}
}
