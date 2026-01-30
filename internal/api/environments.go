package api

import (
	"context"
	"fmt"
	"time"
)

// Environment represents a Upsun environment (branch).
type Environment struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	MachineName string    `json:"machine_name"`
	Title       string    `json:"title"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Parent      string    `json:"parent,omitempty"`
	Project     string    `json:"project"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Links       HALLinks  `json:"_links"`

	// HTTP access settings
	HTTPAccess *HTTPAccess `json:"http_access,omitempty"`
}

// HTTPAccess contains HTTP access settings for an environment.
type HTTPAccess struct {
	IsEnabled   bool              `json:"is_enabled"`
	Addresses   []AccessAddress   `json:"addresses,omitempty"`
	BasicAuth   map[string]string `json:"basic_auth,omitempty"`
}

// AccessAddress represents an IP allowlist entry.
type AccessAddress struct {
	Address string `json:"address"`
}

// ListEnvironments returns all environments for a project.
func (c *Client) ListEnvironments(ctx context.Context, projectID string) ([]Environment, error) {
	path := fmt.Sprintf("/v1/projects/%s/environments", projectID)
	return Collect[Environment](ctx, c, path)
}

// GetEnvironment returns a single environment by ID.
func (c *Client) GetEnvironment(ctx context.Context, projectID, environmentID string) (*Environment, error) {
	var env Environment
	path := fmt.Sprintf("/v1/projects/%s/environments/%s", projectID, environmentID)
	if err := c.Get(ctx, path, &env); err != nil {
		return nil, err
	}
	return &env, nil
}
