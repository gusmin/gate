package commands

import (
	"os"

	"github.com/gusmin/gate/pkg/session"
	"github.com/spf13/cobra"
)

func newExitCommand(sess *session.SecureGateSession) *cobra.Command {
	return &cobra.Command{
		Use:   "exit",
		Short: sess.Translator.Translate("ExitShortDesc", nil),
		Long:  sess.Translator.Translate("ExitShortDesc", nil),
		Run: func(cmd *cobra.Command, args []string) {
			// avoid leaking goroutines
			sess.SignOut()

			os.Exit(0)
		},
	}
}
