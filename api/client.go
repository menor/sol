// Copyright 2026 José Menor
// Licensed under the Apache License, Version 2.0.
// See LICENSE and NOTICE files for details.

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"golang.org/x/oauth2"

	"github.com/menor/sol/auth"
)

const (
	// DefaultBaseURL is the Upsun API base URL.
	DefaultBaseURL = "https://api.upsun.com"

	// DefaultTimeout is the default HTTP request timeout.
	DefaultTimeout = 30 * time.Second
)

// Client is the Upsun API client.
type Client struct {
	httpClient *http.Client
	baseURL    *url.URL
	logFunc    func(format string, args ...any)
}

// ClientOption configures the client.
type ClientOption func(*clientConfig)

// clientConfig holds configuration for building a Client.
type clientConfig struct {
	baseURL     string
	tokenSource oauth2.TokenSource
	timeout     time.Duration
	retryConfig *RetryConfig
	logFunc     func(format string, args ...any)
}

// WithBaseURL sets a custom base URL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *clientConfig) {
		c.baseURL = baseURL
	}
}

// WithTokenSource sets the token source for authentication.
func WithTokenSource(ts oauth2.TokenSource) ClientOption {
	return func(c *clientConfig) {
		c.tokenSource = ts
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *clientConfig) {
		c.timeout = timeout
	}
}

// WithRetryConfig sets custom retry configuration.
func WithRetryConfig(cfg RetryConfig) ClientOption {
	return func(c *clientConfig) {
		c.retryConfig = &cfg
	}
}

// WithLogFunc sets a logging function for debug output.
func WithLogFunc(f func(format string, args ...any)) ClientOption {
	return func(c *clientConfig) {
		c.logFunc = f
	}
}

// New creates a new API client with the given options.
// If no TokenSource is provided, it uses auth.TokenSource to get credentials.
func New(ctx context.Context, opts ...ClientOption) (*Client, error) {
	cfg := &clientConfig{
		baseURL: DefaultBaseURL,
		timeout: DefaultTimeout,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	// Parse base URL
	baseURL, err := url.Parse(cfg.baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}
	if baseURL.Host == "" {
		return nil, fmt.Errorf("invalid base URL (missing host): %s", cfg.baseURL)
	}

	// Get token source if not provided
	tokenSource := cfg.tokenSource
	if tokenSource == nil {
		tokenSource, err = auth.TokenSource(ctx)
		if err != nil {
			return nil, fmt.Errorf("get token source: %w", err)
		}
	}

	// Build retry config
	retryConfig := DefaultRetryConfig
	if cfg.retryConfig != nil {
		retryConfig = *cfg.retryConfig
	}

	// Build transport with retry and auth
	transport := &Transport{
		Base: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
		TokenSource: tokenSource,
		RetryConfig: retryConfig,
		LogFunc:     cfg.logFunc,
	}

	return &Client{
		httpClient: &http.Client{
			Timeout:   cfg.timeout,
			Transport: transport,
		},
		baseURL: baseURL,
		logFunc: cfg.logFunc,
	}, nil
}

// Get performs a GET request to the given path.
func (c *Client) Get(ctx context.Context, urlPath string, result any) error {
	return c.do(ctx, http.MethodGet, urlPath, nil, result)
}

// GetText performs a GET request that returns plain text.
func (c *Client) GetText(ctx context.Context, urlPath string) (string, error) {
	reqURL := c.resolveURL(urlPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "text/plain")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", parseAPIError(resp.StatusCode, body)
	}

	return string(body), nil
}

// Post performs a POST request to the given path.
func (c *Client) Post(ctx context.Context, urlPath string, body, result any) error {
	return c.do(ctx, http.MethodPost, urlPath, body, result)
}

// Patch performs a PATCH request to the given path.
func (c *Client) Patch(ctx context.Context, urlPath string, body, result any) error {
	return c.do(ctx, http.MethodPatch, urlPath, body, result)
}

// Delete performs a DELETE request to the given path.
func (c *Client) Delete(ctx context.Context, urlPath string) error {
	return c.do(ctx, http.MethodDelete, urlPath, nil, nil)
}

// log calls logFunc if it's set, used for debug output.
func (c *Client) log(format string, args ...any) {
	if c.logFunc != nil {
		c.logFunc(format, args...)
	}
}

// do executes an HTTP request.
func (c *Client) do(ctx context.Context, method, urlPath string, body, result any) error {
	// Build URL
	reqURL := c.resolveURL(urlPath)

	// Build request body
	var bodyReader io.Reader
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	c.log("REQUEST: %s %s", method, reqURL.String())
	if len(bodyBytes) > 0 {
		c.log("REQUEST BODY: %s", string(bodyBytes))
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	c.log("RESPONSE: %d (%d bytes)", resp.StatusCode, len(respBody))

	// Check for errors
	if resp.StatusCode >= 400 {
		c.log("RESPONSE BODY: %s", string(respBody))
		return parseAPIError(resp.StatusCode, respBody)
	}

	// Parse response
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}
	}

	return nil
}

// resolveURL resolves a path against the base URL.
// Handles query strings by parsing the urlPath as a URL reference.
func (c *Client) resolveURL(urlPath string) *url.URL {
	// Parse the urlPath to properly handle query strings
	ref, err := url.Parse(urlPath)
	if err != nil {
		// Log warning about URL parse failure
		if c.logFunc != nil {
			c.logFunc("warning: failed to parse URL path %q: %v, using fallback", urlPath, err)
		}
		// Fallback to old behavior if parsing fails
		ref = &url.URL{Path: path.Join(c.baseURL.Path, urlPath)}
	} else {
		// Join the paths properly
		ref.Path = path.Join(c.baseURL.Path, ref.Path)
	}
	return c.baseURL.ResolveReference(ref)
}


// parseAPIError creates an error from an API error response.
func parseAPIError(statusCode int, body []byte) error {
	// Try to parse as JSON error
	var apiErr struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		Detail  string `json:"detail"`
	}
	if err := json.Unmarshal(body, &apiErr); err == nil {
		msg := apiErr.Error
		if msg == "" {
			msg = apiErr.Message
		}
		if msg == "" {
			msg = apiErr.Detail
		}
		// Include raw body for debugging if no message found
		if msg == "" && len(body) > 0 {
			msg = string(body)
		}
		if msg != "" {
			return &APIError{
				StatusCode: statusCode,
				Message:    msg,
				Body:       body,
			}
		}
	}

	// Fall back to status text
	return &APIError{
		StatusCode: statusCode,
		Message:    http.StatusText(statusCode),
		Body:       body,
	}
}

// APIError represents an error response from the API.
type APIError struct {
	StatusCode int
	Message    string
	Body       []byte
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}
