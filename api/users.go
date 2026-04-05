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

// ProjectUserAccess represents a user's access to a project.
type ProjectUserAccess struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	Role        string   `json:"role,omitempty"`
}

// ProjectUserAccessSummary is a lean representation for list output.
type ProjectUserAccessSummary struct {
	UserID string `json:"user_id"`
	Email  string `json:"email,omitempty"`
	Role   string `json:"role,omitempty"`
}

// ToSummary converts ProjectUserAccess to ProjectUserAccessSummary.
func (p ProjectUserAccess) ToSummary() ProjectUserAccessSummary {
	return ProjectUserAccessSummary{
		UserID: p.UserID,
		Email:  p.Email,
		Role:   p.Role,
	}
}

// ListProjectUsers returns all users with access to a project.
func (c *Client) ListProjectUsers(ctx context.Context, projectID string) ([]ProjectUserAccess, error) {
	path := fmt.Sprintf("/projects/%s/user-access", url.PathEscape(projectID))

	var result struct {
		Items []ProjectUserAccess `json:"items"`
		Links HALLinks            `json:"_links"`
	}
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}

	// Resolve user details from HAL links if needed
	users := result.Items

	// Try to get user details from HAL links (ref:users:N pattern)
	for i := range users {
		linkName := fmt.Sprintf("ref:users:%d", i)
		if href, ok := result.Links.GetHREF(linkName); ok {
			var user struct {
				Email string `json:"email"`
			}
			if err := c.Get(ctx, href, &user); err == nil {
				users[i].Email = user.Email
			}
		}
	}

	// Sort for deterministic output
	sort.Slice(users, func(i, j int) bool {
		return users[i].Email < users[j].Email
	})

	return users, nil
}
