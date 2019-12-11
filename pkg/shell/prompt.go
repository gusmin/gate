package shell

import (
	"log"

	"github.com/chzyer/readline"
	"github.com/gusmin/gate/pkg/session"
)

// secureGatePrompt is the prompt shown in Secure Gate shell.
const securegatePrompt = "\033[36;1;1msecuregate$\033[0m "

// SecureGatePrompt is the default prompt for Secure Gate.
type SecureGatePrompt struct {
	// contains filtered or unexported fields
	prompt *readline.Instance
}

// makeConnectCommandCompleter returns a function which is used
// to make dynamic completion with accessible nodes for the connect command.
func makeConnectCommandCompleter(sess *session.SecureGateSession) readline.DynamicCompleteFunc {
	return func(line string) []string {
		var machineNames []string

		for _, m := range sess.Machines() {
			machineNames = append(machineNames, m.Name)
		}

		return machineNames
	}
}

// NewSecureGatePrompt instanciates a prompt for a Secure Gate shell
// with the appropriate completer.
func NewSecureGatePrompt(sess *session.SecureGateSession) *SecureGatePrompt {
	completer := readline.NewPrefixCompleter(
		readline.PcItem("connect",
			readline.PcItemDynamic(makeConnectCommandCompleter(sess)),
		),
		readline.PcItem("list"),
		readline.PcItem("me"),
		readline.PcItem("logout"),
		readline.PcItem("exit"),
	)

	prompt, err := readline.NewEx(&readline.Config{
		Prompt:            securegatePrompt,
		AutoComplete:      completer,
		InterruptPrompt:   "^C",
		HistorySearchFold: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	return &SecureGatePrompt{prompt}
}

// Readline override the prompt if one given,
// reads the current input and return it or an error
// if it finds EOF or if the user sent a SIGINT signal.
func (p *SecureGatePrompt) Readline(prompt string) (string, error) {
	if prompt != "" {
		oldPrompt := p.prompt.Config.Prompt
		p.prompt.SetPrompt(prompt)
		defer p.prompt.SetPrompt(oldPrompt)
	}
	return p.prompt.Readline()
}

// ReadPassword display the given prompt, reads the input
// in no echo mode and return it or an error if if finds EOF
// or if the user sent a SIGINT signal.
func (p *SecureGatePrompt) ReadPassword(prompt string) (string, error) {
	b, err := p.prompt.ReadPassword(prompt)
	password := string(b)
	return password, err
}

// Close the prompt.
// Make sure to call this method after using the prompt.
func (p *SecureGatePrompt) Close() error { return p.prompt.Close() }
