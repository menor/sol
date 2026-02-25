package api

import (
	"context"
	"fmt"
	"net/url"
	"sort"
)

// Certificate represents an SSL certificate for a project.
type Certificate struct {
	ID            string            `json:"id"`
	Key           string            `json:"key,omitempty"`
	Certificate   string            `json:"certificate,omitempty"`
	Chain         []string          `json:"chain,omitempty"`
	Domains       []string          `json:"domains,omitempty"`
	ExpiresAt     string            `json:"expires_at,omitempty"`
	CreatedAt     string            `json:"created_at,omitempty"`
	UpdatedAt     string            `json:"updated_at,omitempty"`
	IsProvisioned bool              `json:"is_provisioned"`
	IsInvalid     bool              `json:"is_invalid,omitempty"`
	Issuer        []CertificateAttr `json:"issuer,omitempty"`
	Links         HALLinks          `json:"_links,omitempty"`
}

// CertificateAttr represents a certificate attribute (issuer fields).
type CertificateAttr struct {
	OID   string `json:"oid,omitempty"`
	Alias string `json:"alias,omitempty"`
	Value string `json:"value,omitempty"`
}

// CertificateSummary is a lean representation for list output.
type CertificateSummary struct {
	ID            string   `json:"id"`
	Domains       []string `json:"domains,omitempty"`
	ExpiresAt     string   `json:"expires_at,omitempty"`
	IsProvisioned bool     `json:"is_provisioned"`
}

// ToSummary converts Certificate to CertificateSummary.
func (c Certificate) ToSummary() CertificateSummary {
	return CertificateSummary{
		ID:            c.ID,
		Domains:       c.Domains,
		ExpiresAt:     c.ExpiresAt,
		IsProvisioned: c.IsProvisioned,
	}
}

// ListCertificates returns all SSL certificates for a project.
func (c *Client) ListCertificates(ctx context.Context, projectID string) ([]Certificate, error) {
	path := fmt.Sprintf("/projects/%s/certificates", url.PathEscape(projectID))

	var result struct {
		Items []Certificate `json:"items"`
	}
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}

	// Sort for deterministic output
	sort.Slice(result.Items, func(i, j int) bool {
		return result.Items[i].ID < result.Items[j].ID
	})

	return result.Items, nil
}
