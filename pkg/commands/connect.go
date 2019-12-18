package commands

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strconv"

	"github.com/gusmin/gate/pkg/backend"
	"github.com/gusmin/gate/pkg/core"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

func newConnectCommand(core *core.SecureGateCore) *cobra.Command {
	return &cobra.Command{
		Use:          "connect [machine]",
		Short:        core.Translator.Translate("ConnectShortDesc"),
		Long:         core.Translator.Translate("ConnectShortDesc"),
		SilenceUsage: true,
		Args:         cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return connect(core.SSHUser, args[0], core.User(), core.Machines(), core.Logger)
		},
	}
}

func connect(sshUser, machineName string, sgUser backend.User, machines []backend.Machine, logger *logrus.Logger) error {
	// Check for existing node
	var machine backend.Machine
	for _, m := range machines {
		if m.Name == machineName {
			machine = m
		}
	}
	if (backend.Machine{}) == machine {
		return fmt.Errorf("%s is not part of accessible machines", machineName)
	}

	// Setup the config
	signer, err := makePrivateKeySigner(path.Join(os.Getenv("HOME"), ".sgsh", sgUser.ID, "id_rsa"))
	if err != nil {
		return errors.Wrap(err, "could not make private key signer")
	}
	config := &ssh.ClientConfig{
		User:            sshUser,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Dial the server
	// conn, err := ssh.Dial("tcp", machine.IP+":22", config)
	conn, err := ssh.Dial("tcp", net.JoinHostPort(machine.IP, strconv.Itoa(machine.AgentPort)), config)
	if err != nil {
		return errors.Wrapf(err, "failed to dial with %s", machineName)
	}
	defer conn.Close()

	// Open a session
	sess, err := conn.NewSession()
	if err != nil {
		return errors.Wrap(err, "failed to create new SSH session")
	}
	defer sess.Close()

	// Loggers with session context
	logFn := logger.WithFields(logrus.Fields{
		"user":    sgUser.ID,
		"machine": machine.ID,
	})
	stdoutLogger := &sshTunnelLogger{log: logFn.Warnf}
	stderrLogger := &sshTunnelLogger{log: logFn.Warnf}

	stdinPipe, err := sess.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "could not pipe stdin")
	}
	go io.Copy(stdinPipe, os.Stdin)

	// Pipe stdout and sterr with logs
	stdoutPipe, err := sess.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "could not pipe stdout")
	}
	go io.Copy(io.MultiWriter(os.Stdout, stdoutLogger), stdoutPipe)

	stderrPipe, err := sess.StderrPipe()
	if err != nil {
		return errors.Wrap(err, "could not pipe stderr")
	}
	go io.Copy(io.MultiWriter(os.Stderr, stderrLogger), stderrPipe)

	// Put the terminal in raw mode and save the old state
	termFD := int(os.Stdin.Fd())
	termState, err := terminal.MakeRaw(termFD)
	if err != nil {
		return errors.Wrap(err, "could not put the terminal in raw mode")
	}
	// Restore terminal state
	defer terminal.Restore(termFD, termState)

	// Terminal attributes and size for pty
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,      // please print what I type
		ssh.ECHOCTL:       0,      // please don't print control chars
		ssh.TTY_OP_ISPEED: 115200, // baud in
		ssh.TTY_OP_OSPEED: 115200, // baud out
	}
	w, h, err := terminal.GetSize(termFD)
	if err != nil {
		return errors.Wrap(err, "could not get size of terminal")
	}

	// Request pty for the session
	err = sess.RequestPty("xterm-256color", h, w, modes)
	if err != nil {
		return errors.Wrap(err, "failed to request pty")
	}

	// Start a shell on the remote host
	err = sess.Shell()
	if err != nil {
		return errors.Wrap(err, "could not start shell on the remote host")
	}

	// Wait for the shell to exit
	return sess.Wait()
}

// makePrivateKeySigner creates a signer from a private SSH key.
func makePrivateKeySigner(file string) (ssh.Signer, error) {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read SSH key from %s", file)
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse private SSH key")
	}
	return key, nil
}

// logFunc logs the format with the given args.
type logFunc func(format string, args ...interface{})

// sshTunnelLogger bufferize written bytes until
// a CRLF, then flush the buffer with the logFunc.
type sshTunnelLogger struct {
	buffer []byte
	log    logFunc
}

func (w *sshTunnelLogger) Write(p []byte) (n int, err error) {
	// bufferize
	w.buffer = append(w.buffer, p...)

	// check for new lines
	for i := 0; i < len(w.buffer); i++ {
		if w.buffer[i] == '\n' {
			w.log("%s\n", string(w.buffer[:i])) // flush the buffer until i without the new line
			w.buffer = w.buffer[i+1:]           // trim the buffer without the new line at position
		}
	}

	return len(p), nil
}
