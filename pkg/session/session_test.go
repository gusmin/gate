package session

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"github.com/gusmin/gate/pkg/backend"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func mockInput(assert *require.Assertions, input string) *os.File {
	tmp, err := ioutil.TempFile("", "")
	assert.NoError(err)

	_, err = tmp.WriteString(input)
	assert.NoError(err)

	_, err = tmp.Seek(0, os.SEEK_SET)
	assert.NoError(err)
	return tmp
}

func TestUpdateMachines(t *testing.T) {
	assert := require.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		const res = `
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
		`
		rw.Write([]byte(res))
	}))
	defer server.Close()
	backendClient := backend.NewClient(server.URL)

	logger := NewLogrusLogger(logrus.New())
	sg := New("randomUser", backendClient, nil, logger)

	sg.updateMachines(context.Background())

	expected := []backend.Machine{
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
	}
	assert.Equal(expected, sg.Machines())
}

func TestUpdateUserInfos(t *testing.T) {
	assert := require.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		const res = `
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
		`
		rw.Write([]byte(res))
	}))
	defer server.Close()
	backendClient := backend.NewClient(server.URL)

	logger := NewLogrusLogger(logrus.New())
	sg := New("randomUser", backendClient, nil, logger)

	sg.updateUser(context.Background())

	expected := backend.User{
		ID:        "superadmin1234",
		Email:     "admin",
		FirstName: "Super",
		LastName:  "adopted",
		Job:       "none",
	}
	assert.Equal(expected, sg.User())
}

func TestSignUp(t *testing.T) {
	assert := require.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		const res = `
		{
			"data": {
				"auth": {
					"success": true,
					"token": "token",
					"message": "hello"
				}
			}
		}
		`
		rw.Write([]byte(res))
	}))
	defer server.Close()
	backendClient := backend.NewClient(server.URL)

	logger := NewLogrusLogger(logrus.New())
	sg := New("randomUser", backendClient, nil, logger)

	err := sg.SignUp("any email", "any password")
	assert.NoError(err)
	assert.True(sg.loggedIn)
}

func TestInitSSHKeys(t *testing.T) {
	assert := require.New(t)

	logger := NewLogrusLogger(logrus.New())
	sg := New("randomUser", nil, nil, logger)

	err := sg.initSSHKeys(os.TempDir())
	_, _, _, _, err = ssh.ParseAuthorizedKey(sg.userInfos.pubKey)
	assert.NoError(err)

	assert.True(exist(path.Join(os.TempDir(), "id_rsa")))
	assert.True(exist(path.Join(os.TempDir(), "id_rsa.pub")))
}

func TestPoll(t *testing.T) {
	assert := require.New(t)

	var inc int

	stopC := make(chan struct{}, 5)
	job := func(ctx context.Context) error {
		return errors.New("dummy error")
	}
	errC := make(chan error)
	go poll(time.Second*1, errC, stopC, job)
	for err := range errC {
		assert.Error(err)
		fmt.Println(err)

		inc++
		if inc == 5 {
			stopC <- struct{}{}
			close(errC)
		}
	}
}
