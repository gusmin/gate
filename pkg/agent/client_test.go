package agent

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSSHAuthorization(t *testing.T) {
	assert := require.New(t)

	// start a local HTTP server which mocks agents behaviour
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// check endpoint
		assert.Equal(req.URL.String(), "/gate/users/foo/ssh-authorization")

		// check headers
		assert.Equal("Bearer token", req.Header.Get("Authorization"))
		assert.Equal("application/json", req.Header.Get("Accept"))
		assert.Equal("application/json", req.Header.Get("Content-Type"))

		// check body
		body, err := ioutil.ReadAll(req.Body)
		assert.NoError(err)
		defer req.Body.Close()
		const expected = `{"publicKey":"key"}`
		assert.Equal([]byte(expected), body)

		const resp = `
		{
			"ErrorType": "NoError",
			"Message": "all fine"
		}
		`
		rw.Write([]byte(resp))
	}))
	defer server.Close()

	client := NewClient("token", server.Client())

	resp, err := client.SSHAuthorization(context.Background(), server.URL, "foo", []byte("key"))
	assert.NoError(err)

	expected := SSHAuthResponse{
		ErrorType: "NoError",
		Message:   "all fine",
	}
	assert.Equal(expected, resp)
}
