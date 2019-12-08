package commands

import (
	"strconv"
	"strings"

	"github.com/gusmin/gate/pkg/session"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newListCommand(sess *session.SecureGateSession) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: sess.Translator.Translate("ListShortDesc", nil),
		Long:  sess.Translator.Translate("ListShortDesc", nil),
		Run: func(cmd *cobra.Command, args []string) {
			var sb strings.Builder

			// write table into the string.Builder
			table := tablewriter.NewWriter(&sb)
			table.SetHeader([]string{
				sess.Translator.Translate("ID", nil),
				sess.Translator.Translate("Name", nil),
				sess.Translator.Translate("IP", nil),
				sess.Translator.Translate("AgentPort", nil),
			})
			table.SetCaption(true, sess.Translator.Translate("ListCaption", nil))

			// fill the table
			for _, machine := range sess.Machines() {
				table.Append([]string{
					machine.ID,
					machine.Name,
					machine.IP,
					strconv.Itoa(machine.AgentPort),
				})
			}

			// render the table into the string.Builder
			table.Render()

			sess.Logger.WithFields(session.Fields{
				"user": sess.User().ID,
			}).Infof(sb.String())
		},
	}
}
