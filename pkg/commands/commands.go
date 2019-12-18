// Package commands provides the command tree of Secure Gate.
package commands

import (
	"strings"

	"github.com/gusmin/gate/pkg/core"
	"github.com/spf13/cobra"
)

// SecureGateCommand is the Secure Gate command tree.
type SecureGateCommand struct {
	// contains filtered or unexported fields
	root *cobra.Command
}

// NewSecureGateCommand creates a new command tree using the given Secure Gate core.
func NewSecureGateCommand(core *core.SecureGateCore) *SecureGateCommand {
	root := &cobra.Command{SilenceErrors: true}
	root.AddCommand(
		newListCommand(core),
		newMeCommand(core),
		newConnectCommand(core),
		newLogoutCommand(core),
		newExitCommand(core),
	)

	return &SecureGateCommand{
		root: root,
	}
}

// Execute executes the given command line if not empty
// and returns an error if the command failed.
func (c *SecureGateCommand) Execute(cmd string) error {
	if cmd == "" {
		return nil
	}

	cmd = strings.TrimSpace(cmd)
	args := strings.Fields(cmd)
	if len(args) <= 0 {
		return nil
	}

	c.root.SetArgs(args)

	return c.root.Execute()
}
