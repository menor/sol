package api

import (
	"context"
	"sort"
)

// SSHKey represents a user's SSH key.
type SSHKey struct {
	KeyID       string `json:"key_id"`
	UID         string `json:"uid,omitempty"`
	Title       string `json:"title,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Value       string `json:"value,omitempty"`
	Changed     string `json:"changed,omitempty"`
}

// SSHKeySummary is a lean representation for list output.
type SSHKeySummary struct {
	KeyID       string `json:"key_id"`
	Title       string `json:"title,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

// ToSummary converts SSHKey to SSHKeySummary.
func (k SSHKey) ToSummary() SSHKeySummary {
	return SSHKeySummary{
		KeyID:       k.KeyID,
		Title:       k.Title,
		Fingerprint: k.Fingerprint,
	}
}

// ListSSHKeys returns all SSH keys for the current user.
func (c *Client) ListSSHKeys(ctx context.Context) ([]SSHKey, error) {
	var result struct {
		SSHKeys []SSHKey `json:"ssh_keys"`
	}
	if err := c.Get(ctx, "/me", &result); err != nil {
		return nil, err
	}

	// Sort for deterministic output
	sort.Slice(result.SSHKeys, func(i, j int) bool {
		return result.SSHKeys[i].KeyID < result.SSHKeys[j].KeyID
	})

	return result.SSHKeys, nil
}
