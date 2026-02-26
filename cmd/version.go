package cmd

import "fmt"

// Version information set at build time via -ldflags.
// Example: go build -ldflags "-X github.com/platformsh/sol/cmd.version=1.0.0"
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// VersionCmd prints version information.
type VersionCmd struct{}

// Run executes the version command.
func (c *VersionCmd) Run(ctx *Context) error {
	short := commit
	if len(commit) > 7 {
		short = commit[:7]
	}
	fmt.Printf("sol %s (%s) built %s\n", version, short, date)
	return nil
}

// Version returns the current version string.
func Version() string {
	return version
}
