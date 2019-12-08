package shell

import (
	"io"

	"github.com/gusmin/gate/pkg/session"
	"github.com/pkg/errors"
)

// Prompt is the interface that wraps the Readline and ReadPassword methods.
//
// Readline return the user input or and error.
// ReadlinePasswords set the terminal in no echo mode and prompt the user for password.
type Prompt interface {
	Readline(prompt string) (string, error)
	ReadPassword(prompt string) (string, error)
}

// Command is the interface that wraps Execute method.
//
// Execute executes the command line and return an error in case of failure.
type Command interface {
	Execute(cmd string) error
}

// SecureGateShell is the interactive CLI application of Secure Gate.
// It is a prompt using the Secure Gate completer
type SecureGateShell struct {
	// contains filtered or unexported fields
	prompt  Prompt
	command Command
	sess    *session.SecureGateSession
}

// New instanciates a new SecureGateShell which executes
// the given command after every input and use the given
// SecureGateSession to enhance the completion (e.g. connect command completion).
func New(prompt Prompt, command Command, sess *session.SecureGateSession) *SecureGateShell {
	return &SecureGateShell{
		prompt:  prompt,
		command: command,
		sess:    sess,
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
		err = sh.sess.SignUp(email, password)
		if err != nil {
			sh.sess.Logger.Errorf("%v\n", err)
			continue mainLoop
		}

	inputLoop:
		for {
			cmd, err := sh.prompt.Readline("")
			if err != nil {
				switch err {
				case io.EOF:
					// Sign out the user if the input loop is broken
					sh.sess.SignOut()
					break inputLoop
				default:
					continue
				}
			}
			sh.sess.Logger.WithFields(session.Fields{
				"user": sh.sess.User().ID,
			}).Warnf("%v\n", cmd)

			err = sh.command.Execute(cmd)
			if err != nil {
				sh.sess.Logger.WithFields(session.Fields{
					"user": sh.sess.User().ID,
				}).Errorf("%v\n", err)
			}
		}
	}
}

// askForCredentials prompt the user to input his email and password.
func (sh *SecureGateShell) askForCredentials() (email, password string, err error) {
	email, err = sh.prompt.Readline("Email: ")
	if err != nil {
		err = errors.Wrap(err, "could not read email")
		return
	}

	b, err := sh.prompt.ReadPassword("Password: ")
	password = string(b)
	if err != nil {
		err = errors.Wrap(err, "could not read password")
		return
	}
	return
}
