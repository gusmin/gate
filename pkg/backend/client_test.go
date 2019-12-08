package backend

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const expectedAuthorization = "JWT token"

func assertGQLVarsEq(assert *require.Assertions, expected string, body io.Reader) {
	b, err := ioutil.ReadAll(body)
	assert.NoError(err)

	// get GraphQL variables in the body
	vars := gjson.GetBytes(b, "variables")
	assert.JSONEq(expected, vars.Raw)
}

func TestAuth(t *testing.T) {
	assert := require.New(t)

	// start a local HTTP server which mocks the Auth GraphQL resolver beheviour
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		const expected = `
		{
			"email":"foo",
			"password":"bar"
		}
		`
		assertGQLVarsEq(assert, expected, req.Body)

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

	client := NewClient(server.URL)

	resp, err := client.Auth(context.Background(), "foo", "bar")
	assert.NoError(err)

	expected := AuthResponse{
		Auth{
			Success: true,
			Token:   "token",
			Message: "hello",
		},
	}
	assert.Equal(expected, resp)
}

func TestMachines(t *testing.T) {
	assert := require.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(expectedAuthorization, req.Header.Get("Authorization"))

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

	client := NewClient(server.URL)
	client.SetToken("token")

	resp, err := client.Machines(context.Background())
	assert.NoError(err)

	expected := MachinesResponse{
		Machines: []Machine{
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
		}}
	assert.Equal(expected, resp)
}

func TestMe(t *testing.T) {
	assert := require.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(expectedAuthorization, req.Header.Get("Authorization"))

		const res = `
		{
			"data": {
				"user": {
					"email": "admin",
					"firstName": "Super"
				}
			}
		}
		`
		rw.Write([]byte(res))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("token")

	resp, err := client.Me(context.Background())
	assert.NoError(err)

	expected := MeResponse{
		User{
			Email:     "admin",
			FirstName: "Super",
		},
	}
	assert.Equal(expected, resp)
}
