package commands

import (
	"github.com/gusmin/gate/pkg/session"
	"github.com/spf13/cobra"
)

func newLogoutCommand(sess *session.SecureGateSession) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: sess.Translator.Translate("LogoutShortDesc", nil),
		Long:  sess.Translator.Translate("LogoutShortDesc", nil),
		Run: func(cmd *cobra.Command, args []string) {
			sess.SignOut()
		},
	}
}
