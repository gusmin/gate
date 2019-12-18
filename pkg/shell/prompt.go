package shell

import (
	"io"

	"github.com/chzyer/readline"
	"github.com/gusmin/gate/pkg/core"
	"github.com/pkg/errors"
)

// secureGatePrompt is the default prompt shown in Secure Gate shell.
var securegatePrompt = "\033[36;1;1msecuregate$\033[0m "

// SecureGatePrompt is the default prompt for Secure Gate.
type SecureGatePrompt struct {
	// contains filtered or unexported fields
	prompt *readline.Instance
}

// makeConnectCommandCompleter returns a function which is used
// to make dynamic completion on connect command with accessible nodes
// of the current user.
func makeConnectCommandCompleter(core *core.SecureGateCore) readline.DynamicCompleteFunc {
	return func(line string) []string {
		var machineNames []string

		for _, m := range core.Machines() {
			machineNames = append(machineNames, m.Name)
		}

		return machineNames
	}
}

// NewSecureGatePrompt instanciates a prompt reading the given io.ReadCloser
// with an enhanced completer. The enhancement is based on the given core.
func NewSecureGatePrompt(in io.ReadCloser, core *core.SecureGateCore) (*SecureGatePrompt, error) {
	completer := readline.NewPrefixCompleter(
		readline.PcItem("connect",
			readline.PcItemDynamic(makeConnectCommandCompleter(core)),
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
		Stdin:             in,
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not instanciate the prompt")
	}
	return &SecureGatePrompt{prompt}, nil
}

// Readline overrides the default prompt if one given,
// reads the user input and returns it or an error
// if it finds EOF or if the user sent a SIGINT signal.
func (p *SecureGatePrompt) Readline(prompt string) (string, error) {
	// Override the default prompt.
	if prompt != "" {
		oldPrompt := p.prompt.Config.Prompt
		p.prompt.SetPrompt(prompt)
		defer p.prompt.SetPrompt(oldPrompt)
	}

	return p.prompt.Readline()
}

// ReadPassword display the given prompt, reads the user input
// without echoing and returns it or an error if it finds EOF
// or if the user sent a SIGINT signal. This is usually the function
// you need when you want a user to input a password.
func (p *SecureGatePrompt) ReadPassword(prompt string) (string, error) {
	b, err := p.prompt.ReadPassword(prompt)
	password := string(b)
	return password, err
}

// Close the prompt.
// Make sure to call this method after using the prompt.
func (p *SecureGatePrompt) Close() error { return p.prompt.Close() }
