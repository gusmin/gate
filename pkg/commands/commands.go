package commands

import (
	"strings"

	"github.com/gusmin/gate/pkg/session"
	"github.com/spf13/cobra"
)

// SecureGateCommand is the Secure Gate command tree.
type SecureGateCommand struct {
	// contains filtered or unexported fields
	root *cobra.Command
	sess *session.SecureGateSession
}

// NewSecureGateCommand creates a new command tree for the given Secure Gate session.
func NewSecureGateCommand(sess *session.SecureGateSession) *SecureGateCommand {
	root := &cobra.Command{SilenceErrors: true}
	root.AddCommand(
		newListCommand(sess),
		newMeCommand(sess),
		newConnectCommand(sess),
		newExitCommand(sess),
	)

	return &SecureGateCommand{
		root: root,
		sess: sess,
	}
}

// Execute executes the given command line if not empty.
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
