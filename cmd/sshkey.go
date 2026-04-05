package cmd

import (
	"github.com/menor/sol/api"
	"github.com/menor/sol/internal/errors"
)

// SSHKeyListCmd lists SSH keys for the current user.
type SSHKeyListCmd struct {
	Full bool `help:"Include all fields" short:"f"`
}

// Run executes the ssh-key:list command.
func (c *SSHKeyListCmd) Run(ctx *Context) error {
	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	keys, err := client.ListSSHKeys(ctx)
	if err != nil {
		return handleAPIError(err, "ssh-keys", "")
	}

	if c.Full {
		return ctx.Output(keys)
	}

	// Return lean summaries
	summaries := make([]api.SSHKeySummary, len(keys))
	for i, key := range keys {
		summaries[i] = key.ToSummary()
	}

	return ctx.Output(summaries)
}
