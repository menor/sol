package api

import (
	"context"
	"fmt"
	"net/url"
	"sort"
)

// Integration represents a project integration (webhook, CI, etc.).
type Integration struct {
	ID        string   `json:"id"`
	Type      string   `json:"type"`
	CreatedAt string   `json:"created_at,omitempty"`
	Role      string   `json:"role,omitempty"`
	TokenType string   `json:"token_type,omitempty"`
	Links     HALLinks `json:"_links,omitempty"`

	// Type-specific fields
	Repository           string   `json:"repository,omitempty"`
	FetchBranches        bool     `json:"fetch_branches,omitempty"`
	BuildPullRequests    bool     `json:"build_pull_requests,omitempty"`
	URL                  string   `json:"url,omitempty"`
	Username             string   `json:"username,omitempty"`
	Project              string   `json:"project,omitempty"`
	Environments         []string `json:"environments,omitempty"`
	ExcludedEnvironments []string `json:"excluded_environments,omitempty"`
	Channel              string   `json:"channel,omitempty"`
	ContinuousProfiling  bool     `json:"continuous_profiling,omitempty"`
}

// IntegrationSummary is a lean representation for list output.
type IntegrationSummary struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// ToSummary converts Integration to IntegrationSummary.
func (i Integration) ToSummary() IntegrationSummary {
	return IntegrationSummary{
		ID:   i.ID,
		Type: i.Type,
	}
}

// ListIntegrationsOptions contains options for listing integrations.
type ListIntegrationsOptions struct {
	Type string // Filter by integration type
}

// ListIntegrations returns all integrations for a project.
func (c *Client) ListIntegrations(ctx context.Context, projectID string, opts ListIntegrationsOptions) ([]Integration, error) {
	path := fmt.Sprintf("/projects/%s/integrations", url.PathEscape(projectID))

	var result struct {
		Items []Integration `json:"items"`
	}
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}

	integrations := result.Items

	// Client-side filtering by type if specified
	if opts.Type != "" {
		filtered := make([]Integration, 0)
		for _, i := range integrations {
			if i.Type == opts.Type {
				filtered = append(filtered, i)
			}
		}
		integrations = filtered
	}

	// Sort for deterministic output
	sort.Slice(integrations, func(i, j int) bool {
		return integrations[i].ID < integrations[j].ID
	})

	return integrations, nil
}

// GetIntegration returns details for a specific integration.
func (c *Client) GetIntegration(ctx context.Context, projectID, integrationID string) (*Integration, error) {
	path := fmt.Sprintf("/projects/%s/integrations/%s",
		url.PathEscape(projectID), url.PathEscape(integrationID))

	var integration Integration
	if err := c.Get(ctx, path, &integration); err != nil {
		return nil, err
	}

	return &integration, nil
}
