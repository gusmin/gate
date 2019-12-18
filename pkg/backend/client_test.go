package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	assert := require.New(t)

	tt := []struct {
		name            string
		email, password string
		expectedVars    string
		resp            string
		expectedRes     AuthResponse
		err             string
	}{
		{
			name:     "valid JSON response",
			email:    "valid email",
			password: "valid password",
			expectedVars: `
				{
					"email": "valid email",
					"password": "valid password"
				}
			`,
			resp: `
			{
				"data": {
					"auth": {
						"success": true,
						"token": "token",
						"message": "hello"
					}
				}
			}
			`,
			expectedRes: AuthResponse{
				Auth{
					Success: true,
					Token:   "token",
					Message: "hello",
				},
			},
			err: "",
		},
		{
			name:     "invalid JSON response",
			email:    "valid email",
			password: "valid password",
			expectedVars: `
				{
					"email": "valid email",
					"password": "valid password"
				}
			`,
			resp: `
			{
				invalid json
			}
			`,
			expectedRes: AuthResponse{},
			err:         "auth request failed: decoding response: invalid character 'i' looking for beginning of object key string",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Start a local HTTP server which mocks the corresponding GraphQL resolver beheviour.
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				// Check GQL variables in the request.
				assertGQLVarsEq(assert, tc.expectedVars, req.Body)

				rw.Write([]byte(tc.resp))
			}))
			defer server.Close()

			client := NewClient(server.URL)

			resp, err := client.Auth(context.Background(), tc.email, tc.password)
			if err != nil {
				assert.Equalf(tc.err, err.Error(),
					"expected error to be: %v, but actual is: %v", tc.err, err)
			}

			assert.Equalf(tc.expectedRes, resp,
				"expected response to be: %+v, but actual is: %+v", tc.expectedRes, resp)
		})
	}
}

func TestMachines(t *testing.T) {
	assert := require.New(t)

	tt := []struct {
		name        string
		token       string
		resp        string
		expectedRes MachinesResponse
		err         string
	}{
		{
			name:  "valid JSON response",
			token: "token",
			resp: `
			{
				"data": {
					"machines": [
						{
							"name": "localhost",
							"ip": "127.0.0.1",
							"agentPort": 3001
						}
					]
				}
			}
			`,
			expectedRes: MachinesResponse{
				Machines: []Machine{
					{
						Name:      "localhost",
						IP:        "127.0.0.1",
						AgentPort: 3001,
					},
				},
			},
			err: "",
		},
		{
			name:  "invalid JSON response",
			token: "token",
			resp: `
			{
				invalid json
			}
			`,
			expectedRes: MachinesResponse{},
			err:         "machines request failed: decoding response: invalid character 'i' looking for beginning of object key string",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Start a local HTTP server which mocks the corresponding GraphQL resolver beheviour.
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				// Check JWT in the header.
				assert.Equal("JWT "+tc.token, req.Header.Get("Authorization"))

				rw.Write([]byte(tc.resp))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			client.token = tc.token

			resp, err := client.Machines(context.Background())
			if err != nil {
				assert.Equalf(tc.err, err.Error(),
					"expected error to be: %v, but actual is: %v", tc.err, err)
			}

			assert.Equalf(tc.expectedRes, resp,
				"expected response to be: %+v, but actual is: %+v", tc.expectedRes, resp)
		})
	}
}

func TestMe(t *testing.T) {
	assert := require.New(t)

	tt := []struct {
		name        string
		token       string
		resp        string
		expectedRes MeResponse
		err         string
	}{
		{
			name:  "valid JSON response",
			token: "token",
			resp: `
			{
				"data": {
					"user": {
						"email": "admin",
						"firstName": "Super"
					}
				}
			}
			`,
			expectedRes: MeResponse{
				User{
					Email:     "admin",
					FirstName: "Super",
				},
			},
			err: "",
		},
		{
			name:  "invalid JSON response",
			token: "token",
			resp: `
			{
				invalid json
			}
			`,
			expectedRes: MeResponse{},
			err:         "me request failed: decoding response: invalid character 'i' looking for beginning of object key string",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Start a local HTTP server which mocks the corresponding GraphQL resolver beheviour.
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				// Check JWT in header.
				assert.Equal("JWT "+tc.token, req.Header.Get("Authorization"))

				rw.Write([]byte(tc.resp))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			client.token = tc.token

			resp, err := client.Me(context.Background())
			if err != nil {
				assert.Equalf(tc.err, err.Error(),
					"expected error to be: %v, but actual is: %v", tc.err, err)
			}

			assert.Equalf(tc.expectedRes, resp,
				"expected response to be: %+v, but actual is: %+v", tc.expectedRes, resp)
		})
	}
}

func TestAddMachineLog(t *testing.T) {
	assert := require.New(t)

	tt := []struct {
		name           string
		token          string
		inputs         []MachineLogInput
		expectedInputs string
		resp           string
		expectedRes    AddMachineLogResponse
		err            string
	}{
		{
			name:  "valid JSON response",
			token: "token",
			inputs: []MachineLogInput{
				{
					Timestamp: 1337,
					MachineID: "randomMachineID",
					UserID:    "randomUserID",
					Log:       "randomLog",
				},
			},
			expectedInputs: `
			{
				"machineLogs": [
					{
						"timestamp": 1337,
						"machineId": "randomMachineID",
						"userId": "randomUserID",
						"log": "randomLog"
					}
				]
			}
			`,
			resp: `
			{
				"data": {
					"addMachineLog": {
						"success": true
					}
				}
			}
			`,
			expectedRes: AddMachineLogResponse{
				BaseResult{
					Success: true,
				},
			},
			err: "",
		},
		{
			name:   "invalid JSON response",
			token:  "token",
			inputs: nil,
			expectedInputs: `
			{
				"machineLogs": null
			}
			`,
			resp: `
			{
				invalid json
			}
			`,
			expectedRes: AddMachineLogResponse{},
			err:         "addMachineLog request failed: decoding response: invalid character 'i' looking for beginning of object key string",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Start a local HTTP server which mocks the corresponding GraphQL resolver beheviour.
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				// Check JWT in the header.
				assert.Equal("JWT "+tc.token, req.Header.Get("Authorization"))

				// Check GQL variables in the request.
				assertGQLVarsEq(assert, tc.expectedInputs, req.Body)

				rw.Write([]byte(tc.resp))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			client.token = tc.token

			resp, err := client.AddMachineLog(context.Background(), tc.inputs)
			if err != nil {
				assert.Equalf(tc.err, err.Error(),
					"expected error to be: %v, but actual is: %v", tc.err, err)
			}

			assert.Equalf(tc.expectedRes, resp,
				"expected response to be: %+v, but actual is: %+v", tc.expectedRes, resp)
		})
	}
}
