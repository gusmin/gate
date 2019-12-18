package commands

import (
	"strconv"
	"strings"

	"github.com/gusmin/gate/pkg/backend"
	"github.com/gusmin/gate/pkg/core"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// newListCommand creates a new "list" command tied to the given core.
func newListCommand(core *core.SecureGateCore) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: core.Translator.Translate("ListShortDesc"),
		Long:  core.Translator.Translate("ListShortDesc"),
		Run: func(cmd *cobra.Command, args []string) {
			list(core.User(), core.Machines(), core.Logger, core.Translator)
		},
	}
}

// list lists machines informations with the logger in table.
func list(
	user backend.User,
	machines []backend.Machine,
	logger *logrus.Logger, translator core.Translator) {

	var sb strings.Builder

	// Write table into the string.Builder.
	table := tablewriter.NewWriter(&sb)
	table.SetHeader([]string{
		translator.Translate("ID"),
		translator.Translate("Name"),
		translator.Translate("IP"),
		translator.Translate("AgentPort"),
	})
	table.SetCaption(true, translator.Translate("ListCaption"))

	// Fill the table.
	for _, machine := range machines {
		table.Append([]string{
			machine.ID,
			machine.Name,
			machine.IP,
			strconv.Itoa(machine.AgentPort),
		})
	}

	// Render the table into the string.Builder.
	table.Render()

	logger.WithFields(logrus.Fields{
		"user": user.ID,
	}).Infof(sb.String())

}
