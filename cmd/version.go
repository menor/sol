package cmd

import "fmt"

// VersionCmd prints version information.
type VersionCmd struct{}

// Run executes the version command.
func (c *VersionCmd) Run(ctx *Context) error {
	fmt.Println("sol version 0.1.0-dev")
	return nil
}
