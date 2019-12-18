package core

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/gusmin/gate/pkg/agent"
	"github.com/gusmin/gate/pkg/backend"
	"github.com/gusmin/gate/pkg/database"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

var (
	// secureGateKeysDir is the directory used to store SSH key pairs for users.
	secureGateKeysDir = path.Join(os.Getenv("HOME"), ".sgsh")
)

// SecureGateCore is the core of Secure Gate.
type SecureGateCore struct {
	// SSH user used for SSH connection with nodes
	SSHUser string
	// Client communicating with Secure Gate backend
	BackendClient BackendClient
	// Client communicating with Secure Gate agents
	AgentClient AgentClient
	// Database repository
	DB DatabaseRepository
	// Logger with fields
	Logger *logrus.Logger
	// Translator for app internationalization
	Translator Translator

	// contains filtered or unexported fields
	loggedIn          bool    // set to true after successful SignUp
	session           session // updated by background polling
	stopPoll          chan struct{}
	stopPollListening chan struct{}
}

// Translator is a language translator.
type Translator interface {
	// Translate translates the given message to another language.
	Translate(message string) string
}

// DatabaseRepository interact with a Secure Gate compliant database.
type DatabaseRepository interface {
	// UpsertUser update the user in the database or insert it if none already exists.
	UpsertUser(user database.User) error
	// FindUser returns the user in the database with the given userID.
	GetUser(userID string) (database.User, error)
}

// BackendClient is a client which can interact with a Secure Gate server.
type BackendClient interface {
	// Auth authenticates a user with the given credentials.
	Auth(ctx context.Context, email, password string) (backend.AuthResponse, error)
	// Machines retrieves accessible nodes for the authenticated user from the server.
	Machines(ctx context.Context) (backend.MachinesResponse, error)
	// Me retrievves user related informations from the server.
	Me(ctx context.Context) (backend.MeResponse, error)
	// AddMachineLog sends logs to the server.
	AddMachineLog(ctx context.Context, inputs []backend.MachineLogInput) (backend.AddMachineLogResponse, error)
}

// AgentClient is a client which can interact with our agents.
type AgentClient interface {
	// AddAuthorizedKey add a new authorized key for the user
	// to the authorized_keys file in the agent running at the given endpoint.
	AddAuthorizedKey(ctx context.Context, endpoint, id string, key []byte) (agent.SSHAuthResponse, error)
	// DeleteAuthorizedKey delete the user authorized key from
	// the authorized_keys file in the agent running at the given endpoint.
	DeleteAuthorizedKey(ctx context.Context, endpoint, id string, key []byte) (agent.SSHAuthResponse, error)
}

// New creates a new Secure Gate core.
func New(sshUser string, backendClient BackendClient, agentClient AgentClient, logger *logrus.Logger, translator Translator, repo DatabaseRepository) *SecureGateCore {
	return &SecureGateCore{
		SSHUser:           sshUser,
		BackendClient:     backendClient,
		AgentClient:       agentClient,
		DB:                repo,
		Logger:            logger,
		Translator:        translator,
		stopPoll:          make(chan struct{}),
		stopPollListening: make(chan struct{}),
	}
}

// SignUp sign up the user to the backend and initialize the user session if successful.
func (core *SecureGateCore) SignUp(email, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	resp, err := core.BackendClient.Auth(ctx, email, string(password))
	if err != nil {
		return errors.Wrap(err, "authentication during sign up failed")
	}
	if resp.Auth.Success == false {
		return fmt.Errorf("authentication during sign up failed: %s", resp.Auth.Message)
	}

	err = core.initUserSession()
	if err != nil {
		return errors.Wrap(err, "could not initialize the user's session")
	}

	core.loggedIn = true

	return nil
}

// initUserSession set up the authenticated user session by
// retrieving the user related informations, generating SSH keypair
// if none already exists, send them to accessible nodes agents
// and start a background polling of user informations.
func (core *SecureGateCore) initUserSession() error {
	ctx := context.Background()

	// Update user's informations
	err := core.updateUser(ctx)
	if err != nil {
		return err
	}
	user := core.User()

	err = core.DB.UpsertUser(database.User{ID: user.ID})
	if err != nil {
		return err
	}

	// Update user's accessible nodes
	err = core.updateMachines(ctx)
	if err != nil {
		return err
	}

	// Check for existing SSH keys
	userSSHDir := path.Join(secureGateKeysDir, user.ID)
	if exist(userSSHDir) == false {
		// Generate new ones if they do not exist already
		err := core.initSSHKeys(userSSHDir)
		if err != nil {
			return errors.Wrap(err, "failed to init ssh keys")
		}
	}

	// Load the user public key to send it to agents if needed
	err = core.loadPublicSSHKey(userSSHDir)
	if err != nil {
		return errors.Wrap(err, "could not load public ssh key")
	}

	// Communicate user permissions changes to agents
	err = core.updateAgents(ctx)
	if err != nil {
		return err
	}

	// Poll accessible nodes and user's informations periodically
	errC := make(chan error, 3)
	go poll(
		time.Second*10,
		errC,
		core.stopPoll,
		// jobs
		core.updateUser,
		core.updateMachines,
		core.updateAgents,
	)
	go func(ctx context.Context) {
		for {
			select {
			case <-core.stopPollListening:
				return
			case err := <-errC:
				if err != nil {
					core.Logger.WithFields(logrus.Fields{
						"user": user.ID,
					}).Warnf("Could not request server, you may havetransform lost network\n")
				}
			}
		}
	}(ctx)

	core.Logger.WithFields(logrus.Fields{
		"user": user.ID,
	}).Infof(core.Translator.Translate("Hello"), user.FirstName, user.LastName)
	return nil
}

// initSSHKeys generate private and public SSH keys for the authenticated user.
func (core *SecureGateCore) initSSHKeys(keysPath string) error {
	err := os.MkdirAll(keysPath, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "could not create directory to store ssh keys")
	}

	privKeyPath := path.Join(keysPath, "id_rsa")
	pubKeyPath := path.Join(keysPath, "id_rsa.pub")
	err = generateSSHKeyPair(pubKeyPath, privKeyPath)
	if err != nil {
		return errors.Wrap(err, "could not generate ssh key pair for this session")
	}
	return nil
}

// loadPublicSSHKey parse the public ssh key located keysPath and
// set the user public key to the parsed key if no error occured.
func (core *SecureGateCore) loadPublicSSHKey(keysPath string) error {
	key, err := ioutil.ReadFile(path.Join(keysPath, "id_rsa.pub"))
	if err != nil {
		return errors.Wrap(err, "could not read public ssh key file")
	}

	// Check if the key is valid
	_, _, _, _, err = ssh.ParseAuthorizedKey(key)
	if err != nil {
		return errors.Wrap(err, "could not parse authorized key")
	}

	core.session.pubKey = key
	return nil
}

// registerKeyToAgent register the user SSH public key in a machine's agent.
func (core *SecureGateCore) registerKeyInAgent(ctx context.Context, machine backend.Machine) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()

	uri := net.JoinHostPort(machine.IP, strconv.Itoa(machine.AgentPort))
	resp, err := core.AgentClient.AddAuthorizedKey(ctx, "http://"+uri, core.User().ID, core.session.pubKey)
	if err != nil {
		return errors.Wrapf(err, "failed to send SSH keys to %s", machine.Name)
	}
	if resp.ErrorType != "" {
		return fmt.Errorf("failed to send SSH keys to %s: %s", machine.Name, resp.Message)
	}
	return nil
}

// unregisterKeyToAgent unregister the user SSH public key in a machine's agent.
func (core *SecureGateCore) unregisterKeyInAgent(ctx context.Context, machine backend.Machine) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()

	uri := net.JoinHostPort(machine.IP, strconv.Itoa(machine.AgentPort))
	resp, err := core.AgentClient.DeleteAuthorizedKey(
		ctx,
		"http://"+uri,
		core.User().ID,
		core.session.pubKey,
	)
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
func (core *SecureGateCore) updateMachines(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()
	resp, err := core.BackendClient.Machines(ctx)
	if err != nil {
		return err
	}

	core.session.machines.set(resp.Machines)
	return nil
}

// updateUser update the user informations by newly retrieves user
// informations from the backend.
func (core *SecureGateCore) updateUser(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()
	resp, err := core.BackendClient.Me(ctx)
	if err != nil {
		return err
	}

	core.session.user.set(resp.User)
	return nil
}

// updateAgents update agents authorized_keys file depending on permissions
// changes.
func (core *SecureGateCore) updateAgents(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()
	user, err := core.DB.GetUser(core.User().ID)
	if err != nil {
		return err
	}

	var insertions []backend.Machine
	var deletions []backend.Machine

	current := make(map[string]database.Machine)
	for _, m := range user.Machines {
		current[m.ID] = m
	}
	received := make(map[string]database.Machine)
	for _, m := range core.Machines() {
		received[m.ID] = database.Machine{
			ID:        m.ID,
			Name:      m.Name,
			IP:        m.IP,
			AgentPort: m.AgentPort,
		}
	}

	for k := range current {
		if _, ok := received[k]; !ok {
			deletions = append(deletions, backend.Machine{
				ID:        current[k].ID,
				Name:      current[k].Name,
				IP:        current[k].IP,
				AgentPort: current[k].AgentPort,
			})
		}
	}
	for k := range received {
		if _, ok := current[k]; !ok {
			insertions = append(insertions, backend.Machine{
				ID:        received[k].ID,
				Name:      received[k].Name,
				IP:        received[k].IP,
				AgentPort: received[k].AgentPort,
			})
		}
	}

	// Agent running on accessible node must add our public key to authorized_keys
	// if the user got rights to access the node.
	for _, m := range insertions {
		err := core.registerKeyInAgent(ctx, m)
		if err != nil {
			core.Logger.WithFields(logrus.Fields{
				"user": user.ID,
			}).Warnf("Could not register key in %s: %v\n", m.Name, err)
		}
	}

	// Agent running on accessible node must delete our public key from authorized_keys
	// if the user lost rights to access the node.
	for _, m := range deletions {
		err := core.unregisterKeyInAgent(ctx, m)
		if err != nil {
			core.Logger.WithFields(logrus.Fields{
				"user": user.ID,
			}).Warnf("Could not unregister key in %s: %v\n", m.Name, err)
		}
	}

	// Update the user machines
	err = core.DB.UpsertUser(database.User{
		ID:       user.ID,
		Machines: transformInDBMachines(core.Machines()),
	})
	if err != nil {
		return err
	}

	return nil
}

// SignOut sign out the user from the current session.
func (core *SecureGateCore) SignOut() {
	user := core.User()
	core.Logger.WithFields(logrus.Fields{
		"user": user.ID,
	}).Infof(core.Translator.Translate("Goodbye"), user.FirstName, user.LastName)

	core.loggedIn = false

	// stop the background polling
	core.stopPoll <- struct{}{}
	// and stop listening to it
	core.stopPollListening <- struct{}{}

	// reset user informations
	core.session = session{}
}

// User returns the current logged in user.
func (core *SecureGateCore) User() backend.User {
	return core.session.user.get()
}

// Machines returns the accessible nodes.
func (core *SecureGateCore) Machines() []backend.Machine {
	return core.session.machines.get()
}

// LoggedIn returns wether an user is logged in
func (core *SecureGateCore) LoggedIn() bool {
	return core.loggedIn
}

type pollingFunc func(ctx context.Context) error

// poll executes the job and returning his error in errC if one returned,
// following the given interval until it receives stop from stopC.
func poll(interval time.Duration, errC chan<- error, stopC <-chan struct{}, jobs ...pollingFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var wg sync.WaitGroup
	for {
		select {
		case <-ticker.C:
			wg.Add(len(jobs))
			for _, job := range jobs {
				go func(job pollingFunc) {
					defer wg.Done()
					errC <- job(ctx)
				}(job)
			}
		case <-stopC:
			return
		}
	}
}

func transformInDBMachines(machines []backend.Machine) []database.Machine {
	var dbMachines []database.Machine

	for _, m := range machines {
		dbMachines = append(dbMachines, database.Machine{
			ID:        m.ID,
			Name:      m.Name,
			IP:        m.IP,
			AgentPort: m.AgentPort,
		})
	}

	return dbMachines
}
