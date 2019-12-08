package session

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/gusmin/gate/pkg/agent"
	"github.com/gusmin/gate/pkg/backend"
	"github.com/pkg/errors"
)

var (
	// secureGateKeysDir is the directory used to store SSH key pairs for users.
	secureGateKeysDir = path.Join(os.Getenv("HOME"), ".sgsh")
)

// SecureGateSession is the session used by users when using Secure Gate.
type SecureGateSession struct {
	Translator Translator
	Logger     StructuredLogger
	// contains filtered or unexported fields
	sshUser       string
	backendClient BackendClient
	agentClient   AgentClient

	loggedIn  bool      // set to true after successful SignUp
	userInfos userInfos // updated by background polling
	stopPoll  chan struct{}
}

// BackendClient is a client which can interact with our server.
type BackendClient interface {
	SetToken(token string)
	Auth(ctx context.Context, email, password string) (backend.AuthResponse, error)
	Machines(ctx context.Context) (backend.MachinesResponse, error)
	Me(ctx context.Context) (backend.MeResponse, error)
	AddMachineLog(ctx context.Context, inputs []backend.MachineLogInput) (backend.AddMachineLogResponse, error)
}

// AgentClient is a client which can interact with our agents.
type AgentClient interface {
	AddSSHPublicKey(ctx context.Context, endpoint, id string, key []byte) (agent.SSHAuthResponse, error)
	DeleteSSHPublicKey(ctx context.Context, endpoint, id string, key []byte) (agent.SSHAuthResponse, error)
}

// Translator wraps the Translate method.
//
// Translate translates the given message to another language.
type Translator interface {
	Translate(message string, template map[string]interface{}) string
}

// New creates a new Secure Gate session.
func New(sshUser string, backendClient BackendClient, agentClient AgentClient, logger StructuredLogger, translator Translator) *SecureGateSession {
	return &SecureGateSession{
		sshUser:       sshUser,
		backendClient: backendClient,
		agentClient:   agentClient,
		stopPoll:      make(chan struct{}),
		Logger:        logger,
		Translator:    translator,
	}
}

// SignUp sign up the user to the backend and initialize the user session if successful.
func (sess *SecureGateSession) SignUp(email, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	resp, err := sess.backendClient.Auth(ctx, email, string(password))
	if err != nil {
		return errors.Wrap(err, "authentication during sign up failed")
	}
	if resp.Auth.Success == false {
		return fmt.Errorf("authentication during sign up failed: %s", resp.Auth.Message)
	}

	// Set the JWT token if successful authentication
	sess.backendClient.SetToken(resp.Auth.Token)

	sess.loggedIn = true

	return sess.initUserSession()
}

// initUserSession set up the authenticated user session by
// retrieving the user related informations, generating SSH keypair
// if none already exists, send them to accessible nodes agents
// and start a background polling of user informations.
func (sess *SecureGateSession) initUserSession() error {
	ctx := context.Background()

	// Update session's informations
	sess.update(ctx)

	// Check for existing SSH keys
	userSSHDir := path.Join(secureGateKeysDir, sess.User().ID)
	if exist(userSSHDir) == false {
		// Generate new ones if they do not exist already
		err := sess.initSSHKeys(userSSHDir)
		if err != nil {
			return err
		}

		// Agents running on accessible nodes must add our public key to authorized_keys
		for _, m := range sess.Machines() {
			err = sess.registerKeyInAgent(ctx, m)
			if err != nil {
				return err
			}
		}
	}

	user := sess.User()

	// Poll accessible nodes and user's informations periodically
	errC := make(chan error, 2)
	go poll(time.Second*10, errC, sess.stopPoll, sess.update)
	go func(ctx context.Context) {
		for {
			select {
			case <-sess.stopPoll:
				return
			case <-ctx.Done():
				return
			case err := <-errC:
				if err != nil {
					sess.Logger.WithFields(Fields{
						"user": user.ID,
					}).Warnf("Could not request server, you may have lost network\n")
				}
			}
		}
	}(ctx)

	sess.Logger.WithFields(Fields{
		"user": user.ID,
	}).Infof("%s\n", sess.Translator.Translate(
		"Hello",
		map[string]interface{}{
			"Firstname": user.FirstName,
			"Lastname":  user.LastName,
		}),
	)
	//"Welcome in Secure Gate %s %s!\n", user.FirstName, user.LastName)
	return nil
}

// initSSHKeys generate private and public SSH keys for the authenticated user.
func (sess *SecureGateSession) initSSHKeys(keysPath string) error {
	err := os.MkdirAll(keysPath, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "could not create directory to store ssh keys")
	}

	privKeyPath := path.Join(keysPath, "id_rsa")
	pubKeyPath := path.Join(keysPath, "id_rsa.pub")
	pubKey, err := generateSSHKeyPair(pubKeyPath, privKeyPath)
	if err != nil {
		return errors.Wrap(err, "could not generate ssh key pair for this session")
	}

	sess.userInfos.pubKey = pubKey
	return nil
}

// registerKeyToAgent register the user SSH public key in a machine's agent.
func (sess *SecureGateSession) registerKeyInAgent(ctx context.Context, machine backend.Machine) error {
	uri := net.JoinHostPort(machine.IP, strconv.Itoa(machine.AgentPort))
	client := agent.NewClient("http://"+uri, nil)

	resp, err := client.AddSSHPublicKey(ctx, uri, sess.User().ID, sess.userInfos.pubKey)
	if err != nil {
		return errors.Wrapf(err, "failed to send SSH keys to %s", machine.Name)
	}
	if resp.ErrorType != "" {
		return fmt.Errorf("failed to send SSH keys to %s: %s", machine.Name, resp.Message)
	}
	return nil
}

// updateMachines update the accessible nodes by newly retrieved machines
// from the backend.
func (sess *SecureGateSession) updateMachines(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()
	resp, err := sess.backendClient.Machines(ctx)
	if err != nil {
		return err
	}

	sess.userInfos.machines.set(resp.Machines)
	return nil
}

// updateUser update the user informations by newly retrieves user
// informations from the backend.
func (sess *SecureGateSession) updateUser(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()
	resp, err := sess.backendClient.Me(ctx)
	if err != nil {
		return err
	}

	sess.userInfos.user.set(resp.User)
	return nil
}

func (sess *SecureGateSession) update(ctx context.Context) error {
	// retrieve informations related to the user
	err := sess.updateUser(ctx)
	if err != nil {
		return errors.Wrap(err, "could not retrieve user infos")
	}
	// retrieve accessibles nodes by the user
	err = sess.updateMachines(ctx)
	if err != nil {
		return errors.Wrap(err, "could not retrieve machines")
	}
	return nil
}

// SignOut sign out the user from the current session.
func (sess *SecureGateSession) SignOut() {
	currentUser := sess.User()
	sess.Logger.WithFields(Fields{
		"user": currentUser.ID,
	}).Infof("%s\n", sess.Translator.Translate(
		"Goodbye",
		map[string]interface{}{
			"Firstname": currentUser.FirstName,
			"Lastname":  currentUser.LastName,
		}),
	)

	sess.loggedIn = false

	//reset user informations
	sess.userInfos.user.set(backend.User{})
	sess.userInfos.machines.set(nil)
	sess.userInfos.pubKey = nil

	// stop the background polling
	sess.stopPoll <- struct{}{}
}

// SSHUser returns the SSH user used to connect to nodes.
func (sess *SecureGateSession) SSHUser() string {
	return sess.sshUser
}

// User returns the current logged in user.
func (sess *SecureGateSession) User() backend.User {
	return sess.userInfos.user.get()
}

// Machines returns the accessible nodes.
func (sess *SecureGateSession) Machines() []backend.Machine {
	return sess.userInfos.machines.get()
}

// poll executes the job and returning his error in errC if one returned,
// following the given interval until it receives stop from stopC.
func poll(interval time.Duration, errC chan<- error, stopC <-chan struct{}, job func(ctx context.Context) error) {
	ctx := context.Background()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			errC <- job(ctx)
		case <-stopC:
			return
		}
	}
}
