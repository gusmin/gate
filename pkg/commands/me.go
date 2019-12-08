package commands

import (
	"strings"

	"github.com/gusmin/gate/pkg/session"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newMeCommand(sess *session.SecureGateSession) *cobra.Command {
	return &cobra.Command{
		Use:   "me",
		Short: sess.Translator.Translate("MeShortDesc", nil),
		Long:  sess.Translator.Translate("MeShortDesc", nil),
		Run: func(cmd *cobra.Command, args []string) {
			var sb strings.Builder

			// write table into the string.Builder
			table := tablewriter.NewWriter(&sb)
			table.SetHeader([]string{
				sess.Translator.Translate("Email", nil),
				sess.Translator.Translate("Firstname", nil),
				sess.Translator.Translate("Lastname", nil),
				sess.Translator.Translate("Job", nil),
			})
			table.SetCaption(true, sess.Translator.Translate("MeCaption", nil))

			// fill the table with existing datas
			user := sess.User()
			var row []string
			if user.Email != "" {
				row = append(row, user.Email)
			}
			if user.FirstName != "" {
				row = append(row, user.FirstName)
			}
			if user.LastName != "" {
				row = append(row, user.LastName)
			}
			if user.Job != "" {
				row = append(row, user.Job)
			}
			table.Append(row)

			// render the table into the string.Builder
			table.Render()

			sess.Logger.WithFields(session.Fields{
				"user": user.ID,
			}).Infof(sb.String())
		},
	}
}
