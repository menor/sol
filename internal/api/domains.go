package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
)

// Domain represents a custom domain for a project.
type Domain struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	IsDefault bool            `json:"is_default"`
	CreatedAt string          `json:"created_at,omitempty"`
	UpdatedAt string          `json:"updated_at,omitempty"`
	SSL       json.RawMessage `json:"ssl,omitempty"`
	Links     HALLinks        `json:"_links,omitempty"`
}

// DomainSummary is a lean representation for list output.
type DomainSummary struct {
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

// ToSummary converts Domain to DomainSummary.
func (d Domain) ToSummary() DomainSummary {
	return DomainSummary{
		Name:      d.Name,
		IsDefault: d.IsDefault,
	}
}

// ListDomains returns all custom domains for a project.
func (c *Client) ListDomains(ctx context.Context, projectID string) ([]Domain, error) {
	path := fmt.Sprintf("/projects/%s/domains", url.PathEscape(projectID))

	var result struct {
		Items []Domain `json:"items"`
	}
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}

	// Sort for deterministic output
	sort.Slice(result.Items, func(i, j int) bool {
		return result.Items[i].Name < result.Items[j].Name
	})

	return result.Items, nil
}
