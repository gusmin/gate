package commands

import (
	"os"

	"github.com/gusmin/gate/pkg/core"
	"github.com/spf13/cobra"
)

func newExitCommand(core *core.SecureGateCore) *cobra.Command {
	return &cobra.Command{
		Use:   "exit",
		Short: core.Translator.Translate("ExitShortDesc"),
		Long:  core.Translator.Translate("ExitShortDesc"),
		Run: func(cmd *cobra.Command, args []string) {
			// avoid leaking goroutines
			core.SignOut()

			os.Exit(0)
		},
	}
}
