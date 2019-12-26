package core

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/gusmin/gate/pkg/agent"
	"github.com/gusmin/gate/pkg/backend"
	"github.com/gusmin/gate/pkg/database"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

type mockTranslator struct{}

func (t mockTranslator) Translate(msg string) string {
	return msg
}

type mockDatabaseRepository struct {
	db map[string]database.User
}

func (repo *mockDatabaseRepository) UpsertUser(user database.User) error {
	repo.db[user.ID] = user
	return nil
}

func (repo *mockDatabaseRepository) GetUser(userID string) (database.User, error) {
	user, ok := repo.db[userID]
	if !ok {
		return database.User{}, fmt.Errorf("could not find user with ID: %s", userID)
	}

	return user, nil
}

type mockAgentClient struct {
	agents map[string][]byte
}

func (c *mockAgentClient) AddAuthorizedKey(ctx context.Context, endpoint, id string, key []byte) (agent.SSHAuthResponse, error) {
	if key == nil {
		return agent.SSHAuthResponse{
			ErrorType: "NilKey",
			Message:   "nil key",
		}, nil
	}

	_, ok := c.agents[endpoint]
	if !ok {
		return agent.SSHAuthResponse{}, fmt.Errorf("no agent running")
	}
	c.agents[endpoint] = key

	return agent.SSHAuthResponse{}, nil
}

func (c *mockAgentClient) DeleteAuthorizedKey(ctx context.Context, endpoint, id string, key []byte) (agent.SSHAuthResponse, error) {
	if key == nil {
		return agent.SSHAuthResponse{
			ErrorType: "NilKey",
			Message:   "nil key",
		}, nil
	}

	_, ok := c.agents[endpoint]
	if !ok {
		return agent.SSHAuthResponse{}, fmt.Errorf("no agent running")
	}
	delete(c.agents, endpoint)

	return agent.SSHAuthResponse{}, nil
}

func init() {
	logrus.SetOutput(ioutil.Discard)
}

func TestUpdateMachines(t *testing.T) {
	assert := require.New(t)

	tt := []struct {
		name         string
		resp         string
		expectedResp []backend.Machine
		err          string
	}{
		{
			name: "valid response",
			resp: `
			{
				"data": {
					"machines": [
						{
							"name": "localhost",
							"ip": "127.0.0.1",
							"agentPort": 3001
						},
						{
							"name": "localhost",
							"ip": "127.0.0.2",
							"agentPort": 3002
						}
					]
				}
			}
			`,
			expectedResp: []backend.Machine{
				{
					Name:      "localhost",
					IP:        "127.0.0.1",
					AgentPort: 3001,
				},
				{
					Name:      "localhost",
					IP:        "127.0.0.2",
					AgentPort: 3002,
				},
			},
			err: "",
		},
		{
			name: "invalid response",
			resp: `
			{
				invalid json
			}
			`,
			expectedResp: nil,
			err:          "invalid character 'i' looking for beginning of object key string",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Write([]byte(tc.resp))
			}))

			backendClient := backend.NewClient(server.URL)

			core := New(
				"randomUser",
				backendClient,
				nil,
				logrus.StandardLogger(),
				&mockTranslator{},
				&mockDatabaseRepository{},
			)

			err := core.updateMachines(context.Background())
			if err != nil {
				assert.Equalf(tc.err, errors.Cause(err).Error(),
					"expected error was: %v, but actual is: %v", tc.err, err)
				return
			}

			resp := core.Machines()
			assert.Equalf(tc.expectedResp, resp,
				"expected response was %+v, but actual is %+v", tc.expectedResp, resp)

			server.Close()
		})
	}
}

func TestUpdateUser(t *testing.T) {
	assert := require.New(t)

	tt := []struct {
		name         string
		resp         string
		expectedResp backend.User
		err          string
	}{
		{
			name: "valid response",
			resp: `
			{
				"data": {
					"user": {
						"id": "superadmin1234",
						"email": "admin",
						"firstName": "Super",
						"lastName": "adopted",
						"job": "none"
					}
				}
			}
			`,
			expectedResp: backend.User{
				ID:        "superadmin1234",
				Email:     "admin",
				FirstName: "Super",
				LastName:  "adopted",
				Job:       "none",
			},
			err: "",
		},
		{
			name: "invalid response",
			resp: `
			{
				invalid json
			}
			`,
			expectedResp: backend.User{},
			err:          "invalid character 'i' looking for beginning of object key string",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Write([]byte(tc.resp))
			}))

			backendClient := backend.NewClient(server.URL)

			core := New(
				"randomUser",
				backendClient,
				nil,
				logrus.StandardLogger(),
				&mockTranslator{},
				&mockDatabaseRepository{},
			)

			err := core.updateUser(context.Background())
			if err != nil {
				assert.Equalf(tc.err, errors.Cause(err).Error(),
					"expected error was: %v, but actual is: %v", tc.err, err)
				return
			}

			resp := core.User()
			assert.Equalf(tc.expectedResp, resp,
				"expected response was %+v, but actual is %+v", tc.expectedResp, resp)

			server.Close()
		})
	}
}

func TestInitSSHKeys(t *testing.T) {
	assert := require.New(t)

	tt := []struct {
		name  string
		path  string
		cause string
	}{
		{
			name:  "valid keysDir path",
			path:  os.TempDir(),
			cause: "",
		},
		{
			name:  "invalid keysDir path",
			path:  "",
			cause: "mkdir : no such file or directory",
		},
	}

	core := New(
		"randomUser",
		nil,
		nil,
		logrus.StandardLogger(),
		&mockTranslator{},
		&mockDatabaseRepository{},
	)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := core.initSSHKeys(tc.path)
			if err != nil {
				assert.Equalf(tc.cause, errors.Cause(err).Error(),
					"expected error was %v, but actual is %v", tc.cause, err)
				return
			}
			publicKeyPath := path.Join(tc.path, "id_rsa.pub")
			defer os.Remove(publicKeyPath)
			privateKeyPath := path.Join(tc.path, "id_rsa")
			defer os.Remove(privateKeyPath)

			// check wether generated authorized key exists and is valid
			authorizedKey, err := ioutil.ReadFile(publicKeyPath)
			assert.NoError(err)
			_, _, _, _, err = ssh.ParseAuthorizedKey(authorizedKey)
			assert.NoError(err)

			// check wether generated private key exists and is valid
			privateKey, err := ioutil.ReadFile(privateKeyPath)
			assert.NoError(err)
			_, err = ssh.ParsePrivateKey(privateKey)
			assert.NoError(err)
		})
	}
}

func TestLoadPublicSSHKey(t *testing.T) {
	assert := require.New(t)

	tt := []struct {
		name        string
		path        string
		content     []byte
		expectedErr string
	}{
		{
			name:        "valid path",
			path:        os.TempDir(),
			expectedErr: "",
		},
		{
			name:        "invalid path",
			path:        "",
			expectedErr: "could not read public ssh key file: open id_rsa.pub: no such file or directory",
		},
		{
			name:        "invalid content",
			path:        os.TempDir(),
			content:     []byte("owghwroghwrh"),
			expectedErr: "could not parse authorized key: ssh: no key found",
		},
	}

	core := New(
		"",
		nil,
		nil,
		logrus.StandardLogger(),
		&mockTranslator{},
		&mockDatabaseRepository{},
	)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.path != "" {
				core.initSSHKeys(tc.path)
			}

			if tc.content != nil {
				ioutil.WriteFile(filepath.Join(tc.path, "id_rsa.pub"), tc.content, os.ModePerm)
			}

			err := core.loadPublicSSHKey(tc.path)
			if err != nil {
				assert.Equalf(tc.expectedErr, err.Error(),
					"expected error was: %v, but actual is: %v", tc.expectedErr, err)
				return
			}

			assert.NotEqual("", core.session.pubKey)
		})
	}
}

func TestRegisterKeyInAgent(t *testing.T) {
	assert := require.New(t)

	tt := []struct {
		name        string
		machine     backend.Machine
		key         []byte
		expectedErr string
	}{
		{
			name: "valid agent",
			machine: backend.Machine{
				IP:        "foo",
				AgentPort: 3000,
			},
			key:         []byte("test"),
			expectedErr: "",
		},
		{
			name:        "no running agent",
			machine:     backend.Machine{},
			key:         []byte("test"),
			expectedErr: "failed to send SSH keys to : no agent running",
		},
		{
			name: "agent error",
			machine: backend.Machine{
				IP:        "foo",
				AgentPort: 3000,
			},
			key:         nil,
			expectedErr: "failed to send SSH keys to : nil key",
		},
	}

	agentClient := mockAgentClient{
		agents: map[string][]byte{
			"http://foo:3000": nil,
		},
	}

	core := New(
		"",
		nil,
		&agentClient,
		logrus.StandardLogger(),
		&mockTranslator{},
		&mockDatabaseRepository{},
	)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			core.session.pubKey = tc.key

			err := core.registerKeyInAgent(context.Background(), tc.machine)
			if err != nil {
				assert.Equalf(tc.expectedErr, err.Error(),
					"expected error was: %v, but actual is: %v", tc.expectedErr, err)
				return
			}

			key := "http://" + net.JoinHostPort(tc.machine.IP, strconv.Itoa(tc.machine.AgentPort))
			assert.Equalf(tc.key, agentClient.agents[key],
				"sent key isn't present in agent's authorized keys")
		})
	}
}

func TestUnregisterKeyInAgent(t *testing.T) {
	assert := require.New(t)

	tt := []struct {
		name        string
		machine     backend.Machine
		key         []byte
		expectedErr string
	}{
		{
			name: "valid agent",
			machine: backend.Machine{
				IP:        "foo",
				AgentPort: 3000,
			},
			key:         []byte("test"),
			expectedErr: "",
		},
		{
			name:        "no running agent",
			machine:     backend.Machine{},
			key:         []byte("test"),
			expectedErr: "failed to send SSH keys to : no agent running",
		},
		{
			name: "agent error",
			machine: backend.Machine{
				IP:        "foo",
				AgentPort: 3000,
			},
			key:         nil,
			expectedErr: "failed to send SSH keys to : nil key",
		},
	}

	agentClient := mockAgentClient{
		agents: map[string][]byte{
			"http://foo:3000": nil,
		},
	}

	core := New(
		"",
		nil,
		&agentClient,
		logrus.StandardLogger(),
		&mockTranslator{},
		&mockDatabaseRepository{},
	)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			core.session.pubKey = tc.key

			err := core.unregisterKeyInAgent(context.Background(), tc.machine)
			if err != nil {
				assert.Equalf(tc.expectedErr, err.Error(),
					"expected error was: %v, but actual is: %v", tc.expectedErr, err)
				return
			}

			key := "http://" + net.JoinHostPort(tc.machine.IP, strconv.Itoa(tc.machine.AgentPort))
			assert.Nil(agentClient.agents[key],
				"sent key isn't deleted in agent")
		})
	}
}

func TestUpdateAgentsAddition(t *testing.T) {
	assert := require.New(t)

	tt := []struct {
		name        string
		user        backend.User
		machines    []backend.Machine
		repo        mockDatabaseRepository
		agentClient mockAgentClient
		expectedErr string
	}{
		{
			name: "unknown user",
			user: backend.User{ID: "qew"},
			repo: mockDatabaseRepository{
				db: map[string]database.User{},
			},
			expectedErr: "could not find user with ID: qew",
		},
		{
			name: "access added",
			user: backend.User{ID: "foobar"},
			machines: []backend.Machine{
				backend.Machine{
					ID:        "anything",
					Name:      "qwe",
					IP:        "foo",
					AgentPort: 3000,
				},
			},
			repo: mockDatabaseRepository{
				db: map[string]database.User{
					"foobar": database.User{
						ID: "foobar",
					},
				},
			},
			agentClient: mockAgentClient{
				agents: map[string][]byte{
					"http://foo:3000": nil,
				},
			},
			expectedErr: "",
		},
		{
			name:     "access removed",
			user:     backend.User{ID: "foobar"},
			machines: []backend.Machine{},
			repo: mockDatabaseRepository{
				db: map[string]database.User{
					"foobar": database.User{
						ID: "foobar",
						Machines: []database.Machine{
							database.Machine{
								ID:        "anything",
								Name:      "qwe",
								IP:        "foo",
								AgentPort: 3000,
							},
						},
					},
				},
			},
			agentClient: mockAgentClient{
				agents: map[string][]byte{
					"http://foo:3000": []byte("test"),
				},
			},
			expectedErr: "",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			core := New(
				"",
				nil,
				&tc.agentClient,
				logrus.StandardLogger(),
				&mockTranslator{},
				&tc.repo,
			)
			core.session.user.set(tc.user)
			core.session.machines.set(tc.machines)
			core.session.pubKey = []byte("test")

			err := core.updateAgents(context.Background())
			if err != nil {
				assert.Equalf(tc.expectedErr, err.Error(),
					"expected error was: %v, but actual is: %v", tc.expectedErr, err)
				return
			}

			dbMachines := transformInDBMachines(tc.machines)
			assert.ElementsMatch(tc.repo.db["foobar"].Machines, dbMachines)

			// TODO: Check for additions and deletions.
		})
	}
}

func TestSignout(t *testing.T) {
	assert := require.New(t)

	core := New(
		"",
		nil,
		&mockAgentClient{},
		logrus.StandardLogger(),
		&mockTranslator{},
		&mockDatabaseRepository{},
	)
	core.loggedIn = true

	go func() {
		for {
			select {
			case <-core.stopPoll:
			case <-core.stopPollListening:
				break
			case <-time.After(3 * time.Second):
				assert.Fail("polling not stopped")
			}
		}
	}()

	core.SignOut()

	assert.False(core.LoggedIn())
	assert.Empty(core.session)
}

func TestPoll(t *testing.T) {
	assert := require.New(t)

	var inc int

	errJob := func(ctx context.Context) error {
		return errors.New("dummy error")
	}

	errC := make(chan error, 5)
	stopC := make(chan struct{})

	go poll(time.Microsecond, errC, stopC, errJob)

	for err := range errC {
		assert.EqualError(err, "dummy error")

		inc++
		if inc == 5 {
			stopC <- struct{}{}
			close(errC)
		}
	}
}
