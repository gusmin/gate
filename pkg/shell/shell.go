// Package shell provides the implementation of the
// Secure Gate's interactive CLI.
package shell

import (
	"io"

	"github.com/gusmin/gate/pkg/core"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// SecureGateShell is the interactive CLI of Secure Gate.
type SecureGateShell struct {
	Prompt  Prompt
	Command Command
	Core    *core.SecureGateCore
}

// Prompt reads input from a user.
type Prompt interface {
	// Readline prompt for user input and returns
	// it or an error.
	Readline(prompt string) (string, error)
	// ReadlinePasswords set the terminal in no echo mode
	// and prompt for user input.
	ReadPassword(prompt string) (string, error)
}

// Command is an executable command tree.
type Command interface {
	// Execute executes the command line and returns an error in case of failure.
	Execute(cmd string) error
}

// NewSecureGateShell instanciates a new SecureGateShell which executes
// commands using the given command tree after every input and uses the
// SecureGateCore for user sessions management.
func NewSecureGateShell(prompt Prompt, command Command, core *core.SecureGateCore) *SecureGateShell {
	return &SecureGateShell{
		Prompt:  prompt,
		Command: command,
		Core:    core,
	}
}

// Run runs the shell loop.
func (sh *SecureGateShell) Run() error {
	// Try to authenticate the user until it succeeds.
mainLoop:
	for {
		email, password, err := sh.askForCredentials()
		if err != nil {
			return err
		}
		err = sh.Core.SignUp(email, password)
		if err != nil {
			sh.Core.Logger.Errorf("%v\n", err)
			continue mainLoop
		}
		user := sh.Core.User()

		// Executes every commands typed by the authenticated user.
	inputLoop:
		for sh.Core.LoggedIn() == true {
			cmd, err := sh.Prompt.Readline("")
			if err != nil {
				switch err {
				case io.EOF:
					// Sign out the user if the input loop is broken.
					sh.Core.SignOut()
					break inputLoop
				default:
					continue inputLoop
				}
			}
			// Log the input in log file if one exists.
			sh.Core.Logger.WithFields(logrus.Fields{
				"user": user.ID,
			}).Warnf("%s\n", cmd)

			// Executes the command line.
			err = sh.Command.Execute(cmd)
			if err != nil {
				sh.Core.Logger.WithFields(logrus.Fields{
					"user": user.ID,
				}).Errorf("%s\n", err)
			}
		}
	}
}

// askForCredentials prompt for user email and password.
func (sh *SecureGateShell) askForCredentials() (email, password string, err error) {
	email, err = sh.Prompt.Readline("Email: ")
	if err != nil {
		err = errors.Wrap(err, "could not read email")
		return
	}

	b, err := sh.Prompt.ReadPassword("Password: ")
	password = string(b)
	if err != nil {
		err = errors.Wrap(err, "could not read password")
		return
	}

	return
}
