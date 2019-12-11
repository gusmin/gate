package shell

import (
	"io"

	"github.com/gusmin/gate/pkg/session"
	"github.com/pkg/errors"
)

// Prompt reads input from a user.
type Prompt interface {
	// Readline return the user input or and error.
	Readline(prompt string) (string, error)
	// ReadlinePasswords set the terminal in no echo mode and prompt the user for password.
	ReadPassword(prompt string) (string, error)
}

// Command is an executable command tree.
type Command interface {
	// Execute executes the command line and return an error in case of failure.
	Execute(cmd string) error
}

// SecureGateShell is the interactive CLI application of Secure Gate.
// It is a prompt using the Secure Gate completer
type SecureGateShell struct {
	Prompt  Prompt
	Command Command
	Sess    *session.SecureGateSession
}

// New instanciates a new SecureGateShell which executes
// the given command after every input and use the given
// SecureGateSession to enhance the completion (e.g. connect command completion).
func New(prompt Prompt, command Command, sess *session.SecureGateSession) *SecureGateShell {
	return &SecureGateShell{
		Prompt:  prompt,
		Command: command,
		Sess:    sess,
	}
}

// Run starts the shell.
func (sh *SecureGateShell) Run() error {
mainLoop:
	for {
		email, password, err := sh.askForCredentials()
		if err != nil {
			return err
		}
		err = sh.Sess.SignUp(email, password)
		if err != nil {
			sh.Sess.Logger.Errorf("%v\n", err)
			continue mainLoop
		}

	inputLoop:
		for sh.Sess.LoggedIn() == true {
			cmd, err := sh.Prompt.Readline("")
			if err != nil {
				switch err {
				case io.EOF:
					// Sign out the user if the input loop is broken
					sh.Sess.SignOut()
					break inputLoop
				default:
					continue inputLoop
				}
			}
			sh.Sess.Logger.WithFields(session.Fields{
				"user": sh.Sess.User().ID,
			}).Warnf("%v\n", cmd)

			err = sh.Command.Execute(cmd)
			if err != nil {
				sh.Sess.Logger.WithFields(session.Fields{
					"user": sh.Sess.User().ID,
				}).Errorf("%v\n", err)
			}
		}
	}
}

// askForCredentials prompt the user to input his email and password.
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
