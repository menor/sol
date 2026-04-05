// Copyright 2026 José Menor
// Licensed under the Apache License, Version 2.0.
// See LICENSE and NOTICE files for details.

package api

import (
	"context"
	"fmt"
	"net/url"
	"sort"
)

// Organization represents an Upsun organization.
type Organization struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Label        string   `json:"label,omitempty"`
	Owner        string   `json:"owner,omitempty"`
	Country      string   `json:"country,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// OrganizationSummary is a lean representation for list output.
type OrganizationSummary struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Label string `json:"label,omitempty"`
}

// ToSummary converts an Organization to OrganizationSummary.
func (o Organization) ToSummary() OrganizationSummary {
	return OrganizationSummary{
		ID:    o.ID,
		Name:  o.Name,
		Label: o.Label,
	}
}

// ListOrganizations returns all organizations for the current user.
func (c *Client) ListOrganizations(ctx context.Context) ([]Organization, error) {
	// First get the current user ID
	userID, err := c.getCurrentUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("get current user: %w", err)
	}

	path := fmt.Sprintf("/users/%s/organizations", url.PathEscape(userID))

	var result struct {
		Items []Organization `json:"items"`
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

// GetOrganization returns details for a specific organization.
func (c *Client) GetOrganization(ctx context.Context, orgID string) (*Organization, error) {
	path := fmt.Sprintf("/organizations/%s", url.PathEscape(orgID))

	var org Organization
	if err := c.Get(ctx, path, &org); err != nil {
		return nil, err
	}

	return &org, nil
}

// getCurrentUserID retrieves the current user's ID.
func (c *Client) getCurrentUserID(ctx context.Context) (string, error) {
	var user struct {
		ID string `json:"id"`
	}
	if err := c.Get(ctx, "/users/me", &user); err != nil {
		return "", err
	}
	return user.ID, nil
}
