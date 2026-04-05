package cmd

import (
	"context"
	"testing"

	"github.com/menor/sol/api"
)

func TestOrganizationListCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		ListOrganizationsFunc: func(ctx context.Context) ([]api.Organization, error) {
			return []api.Organization{
				{ID: "org1", Name: "my-org", Label: "My Organization"},
				{ID: "org2", Name: "other-org", Label: "Other Org"},
			}, nil
		},
	}

	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		apiClientFactory: func(ctx context.Context) (api.API, error) {
			return mockClient, nil
		},
	}

	cmd := &OrganizationListCmd{}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "ListOrganizations" {
		t.Errorf("expected ListOrganizations call, got %s", mockClient.Calls[0].Method)
	}
}

func TestOrganizationListCmd_Full(t *testing.T) {
	mockClient := &api.MockClient{
		ListOrganizationsFunc: func(ctx context.Context) ([]api.Organization, error) {
			return []api.Organization{
				{ID: "org1", Name: "my-org", Label: "My Organization", Owner: "user123", Country: "US"},
			}, nil
		},
	}

	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		apiClientFactory: func(ctx context.Context) (api.API, error) {
			return mockClient, nil
		},
	}

	cmd := &OrganizationListCmd{Full: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrganizationInfoCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		GetOrganizationFunc: func(ctx context.Context, orgID string) (*api.Organization, error) {
			return &api.Organization{
				ID:      orgID,
				Name:    "my-org",
				Label:   "My Organization",
				Owner:   "user123",
				Country: "US",
			}, nil
		},
	}

	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		apiClientFactory: func(ctx context.Context) (api.API, error) {
			return mockClient, nil
		},
	}

	cmd := &OrganizationInfoCmd{OrgID: "org123"}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mockClient.Calls[0].Args[0] != "org123" {
		t.Errorf("expected org ID 'org123', got %v", mockClient.Calls[0].Args[0])
	}
}
