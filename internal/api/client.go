package api

import (
	"context"
	"net/http"
	"time"
)

// Client is the Upsun API client
type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

// ClientOption configures the client
type ClientOption func(*Client)

// WithBaseURL sets a custom base URL
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithToken sets the authentication token
func WithToken(token string) ClientOption {
	return func(c *Client) {
		c.token = token
	}
}

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// New creates a new API client
func New(opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		baseURL: "https://api.upsun.com",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Do executes an HTTP request with authentication
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// TODO: implement retry logic
	// TODO: implement token refresh on 401

	return c.httpClient.Do(req)
}
