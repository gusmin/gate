package agent

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	assert := require.New(t)

	tt := []struct {
		name                   string
		client, expectedClient *http.Client
	}{
		{
			client:         &http.Client{},
			expectedClient: &http.Client{},
		},
		{
			client:         nil,
			expectedClient: http.DefaultClient,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			client := NewClient("whocareaboutatoken", tc.client)
			assert.Equalf(tc.expectedClient, client.httpClient,
				"expected httpClient to be: %v but actual is: %v\n", tc.expectedClient, client.httpClient)
		})
	}
}

func TestAddAuthorizedKey(t *testing.T) {
	assert := require.New(t)

	tt := []struct {
		name         string
		userID       string
		key          []byte
		expectedBody []byte
		respp        string
		expectedResp SSHAuthResponse
		err          string
	}{
		{
			name:         "valid key",
			userID:       "foo",
			key:          []byte("key"),
			expectedBody: []byte(`{"publicKey":"key"}`),
			respp: `
			{
				"ErrorType": "NoError",
				"Message": "all fine"
			}
			`,
			expectedResp: SSHAuthResponse{
				ErrorType: "NoError",
				Message:   "all fine",
			},
			err: "",
		},
		{
			name:         "emtpy key",
			userID:       "foo",
			key:          nil,
			expectedBody: []byte(`{"publicKey":""}`),
			respp: `
			{
				"ErrorType": "InvalidKey",
				"Message": "empty publicKey"
			}
			`,
			expectedResp: SSHAuthResponse{
				ErrorType: "InvalidKey",
				Message:   "empty publicKey",
			},
			err: "",
		},
		{
			name:         "invalid JSON respponse",
			userID:       "foo",
			key:          []byte("key"),
			expectedBody: []byte(`{"publicKey":"key"}`),
			respp: `
			{
				invalid JSON
			}
			`,
			expectedResp: SSHAuthResponse{},
			err:          "invalid character 'i' looking for beginning of object key string",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Start a local HTTP server which mocks agents behaviour.
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				// Check if endpoint is correctly formatted.
				assert.Equal(req.URL.String(), "/gate/users/foo/ssh-authorization")

				// Check the requestmethod.
				assert.Equal(req.Method, http.MethodPost)

				// Check the headers.
				assert.Equal("Bearer token", req.Header.Get("Authorization"))
				assert.Equal("application/json", req.Header.Get("Accept"))
				assert.Equal("application/json", req.Header.Get("Content-Type"))

				// Check the body.
				body, err := ioutil.ReadAll(req.Body)
				assert.NoError(err)
				defer req.Body.Close()
				assert.Equal(tc.expectedBody, body)

				rw.Write([]byte(tc.respp))
			}))
			defer server.Close()

			client := NewClient("token", server.Client())

			respp, err := client.AddAuthorizedKey(context.Background(), server.URL, tc.userID, tc.key)
			if err != nil {
				assert.Equal(tc.err, errors.Cause(err).Error(),
					"expected error was: %v, but actual is: %v", tc.err, err)
			}

			assert.Equal(tc.expectedResp, respp,
				"expected respponse was: %v, but actual is: %v", tc.expectedResp, respp)
		})
	}
}

func TestDeleteAuthorizedKey(t *testing.T) {
	assert := require.New(t)

	tt := []struct {
		name         string
		userID       string
		key          []byte
		expectedBody []byte
		resp         string
		expectedResp SSHAuthResponse
		err          string
	}{
		{
			name:         "valid key",
			userID:       "foo",
			key:          []byte("key"),
			expectedBody: []byte(`{"publicKey":"key"}`),
			resp: `
			{
				"ErrorType": "NoError",
				"Message": "all fine"
			}
			`,
			expectedResp: SSHAuthResponse{
				ErrorType: "NoError",
				Message:   "all fine",
			},
			err: "",
		},
		{
			name:         "empty key",
			userID:       "foo",
			key:          nil,
			expectedBody: []byte(`{"publicKey":""}`),
			resp: `
			{
				"ErrorType": "InvalidKey",
				"Message": "empty publicKey"
			}
			`,
			expectedResp: SSHAuthResponse{
				ErrorType: "InvalidKey",
				Message:   "empty publicKey",
			},
			err: "",
		},
		{
			name:         "invalid JSON respponse",
			userID:       "foo",
			key:          []byte("key"),
			expectedBody: []byte(`{"publicKey":"key"}`),
			resp: `
			{
				invalid JSON
			}
			`,
			expectedResp: SSHAuthResponse{},
			err:          "invalid character 'i' looking for beginning of object key string",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Start a local HTTP server which mocks agents behaviour.
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				// Check if the request endpoint is correctly formatted.
				assert.Equal(req.URL.String(), "/gate/users/foo/ssh-authorization")

				// Check the request method.
				assert.Equal(req.Method, http.MethodDelete)

				// Check the headers.
				assert.Equal("Bearer token", req.Header.Get("Authorization"))
				assert.Equal("application/json", req.Header.Get("Accept"))
				assert.Equal("application/json", req.Header.Get("Content-Type"))

				// Check the body.
				body, err := ioutil.ReadAll(req.Body)
				assert.NoError(err)
				defer req.Body.Close()
				assert.Equal(tc.expectedBody, body)

				rw.Write([]byte(tc.resp))
			}))
			defer server.Close()

			client := NewClient("token", server.Client())

			respp, err := client.DeleteAuthorizedKey(context.Background(), server.URL, tc.userID, tc.key)
			if err != nil {
				assert.Equal(tc.err, errors.Cause(err).Error(),
					"expected error was: %v, but actual is: %v", tc.err, err)
			}

			assert.Equal(tc.expectedResp, respp,
				"expected respponse was: %v, but actual is: %v", tc.expectedResp, respp)
		})
	}
}
