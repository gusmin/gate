package commands

import (
	"github.com/gusmin/gate/pkg/core"
	"github.com/spf13/cobra"
)

func newLogoutCommand(core *core.SecureGateCore) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: core.Translator.Translate("LogoutShortDesc"),
		Long:  core.Translator.Translate("LogoutShortDesc"),
		Run: func(cmd *cobra.Command, args []string) {
			core.SignOut()
		},
	}
}
