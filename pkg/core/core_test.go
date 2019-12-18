package core

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

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

type mockDatabaseRepository struct{}

func (repo *mockDatabaseRepository) UpsertUser(user database.User) error {
	return nil
}

func (repo *mockDatabaseRepository) GetUser(userID string) (database.User, error) {
	return database.User{
		ID: userID,
		Machines: []database.Machine{
			database.Machine{
				ID:        "pghrpighr",
				Name:      "clairementlenomonsenfout",
				IP:        "cacestlhostlocalsurleport",
				AgentPort: 22,
			},
		},
	}, nil
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

			sess := New(
				"randomUser",
				backendClient,
				nil,
				logrus.StandardLogger(),
				&mockTranslator{},
				&mockDatabaseRepository{},
			)

			err := sess.updateMachines(context.Background())
			if err != nil {
				assert.Equalf(tc.err, errors.Cause(err).Error(),
					"expected error was: %v, but actual is: %v", tc.err, err)
			}

			resp := sess.Machines()
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

			sess := New(
				"randomUser",
				backendClient,
				nil,
				logrus.StandardLogger(),
				&mockTranslator{},
				&mockDatabaseRepository{},
			)

			err := sess.updateUser(context.Background())
			if err != nil {
				assert.Equalf(tc.err, errors.Cause(err).Error(),
					"expected error was: %v, but actual is: %v", tc.err, err)
			}

			resp := sess.User()
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

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sess := New(
				"randomUser",
				nil,
				nil,
				logrus.StandardLogger(),
				&mockTranslator{},
				&mockDatabaseRepository{},
			)

			err := sess.initSSHKeys(tc.path)
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
