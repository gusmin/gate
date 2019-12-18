package commands

import (
	"strings"

	"github.com/gusmin/gate/pkg/backend"
	"github.com/gusmin/gate/pkg/core"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// newMeCommand creates a new "me" command tied to the given core.
func newMeCommand(core *core.SecureGateCore) *cobra.Command {
	return &cobra.Command{
		Use:   "me",
		Short: core.Translator.Translate("MeShortDesc"),
		Long:  core.Translator.Translate("MeShortDesc"),
		Run: func(cmd *cobra.Command, args []string) {
			me(core.User(), core.Logger, core.Translator)
		},
	}
}

// me display informations related to the user with the logger in a table.
func me(user backend.User, logger *logrus.Logger, translator core.Translator) {
	var sb strings.Builder

	// Write table into the string.Builder.
	table := tablewriter.NewWriter(&sb)
	table.SetHeader([]string{
		translator.Translate("Email"),
		translator.Translate("Firstname"),
		translator.Translate("Lastname"),
		translator.Translate("Job"),
	})
	table.SetCaption(true, translator.Translate("MeCaption"))

	// Fill the table with existing datas.
	var row []string
	row = append(
		row,
		[]string{
			user.Email,
			user.FirstName,
			user.LastName,
			user.Job,
		}...,
	)
	table.Append(row)

	// Render the table into the string.Builder.
	table.Render()

	logger.WithFields(logrus.Fields{
		"user": user.ID,
	}).Infof("%s\n", sb.String())
}
