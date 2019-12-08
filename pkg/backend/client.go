package backend

import (
	"context"

	"github.com/gusmin/graphql"
	"github.com/pkg/errors"
)

// Client is a GraphQL client interacting with the backend.
type Client struct {
	// contains filtered or unexported fields
	gqlClient *graphql.Client
	token     string // JWT token used in requests. Must be set with func SetToken
}

// NewClient creates a new GraphQL client pointing to the given backend endpoint.
func NewClient(endpoint string) *Client {
	return &Client{gqlClient: graphql.NewClient(endpoint)}
}

// SetToken set the JWT token used to authenticate on the server
// This token is given by the backend after a successful Auth request.
func (c *Client) SetToken(token string) { c.token = token }

// AuthResponse is the response sent by the server after an auth query.
type AuthResponse struct {
	Auth Auth `json:"auth"`
}

// Auth contains a success value that can be either true or false.
// When it is true the token identity is returned as well as a success message.
// Otherwise no token is returned and you should refer to the error message.
type Auth struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
	Message string `json:"message"`
}

// Auth authenticates the user with the given credentials.
func (c *Client) Auth(ctx context.Context, email, password string) (AuthResponse, error) {
	var res AuthResponse
	err := c.gqlClient.Run(ctx, makeAuthRequest(email, password), &res)
	if err != nil {
		return AuthResponse{}, errors.Wrap(err, "auth request failed")
	}
	return res, nil
}

// MachinesResponse is the response sent by the server after a Machines query.
type MachinesResponse struct {
	Machines []Machine `json:"machines"`
}

// Machine is the machine informations.
type Machine struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IP        string `json:"ip"`
	AgentPort int    `json:"agentPort"`
}

// Machines retrieves all the accessible nodes by the authenticated user.
func (c *Client) Machines(ctx context.Context) (MachinesResponse, error) {
	var res MachinesResponse
	err := c.gqlClient.Run(ctx, makeMachinesRequest(c.token), &res)
	if err != nil {
		return MachinesResponse{}, errors.Wrap(err, "machines request failed")
	}
	return res, nil
}

// MeResponse is the response sent by the server after a Me query.
type MeResponse struct {
	User User `json:"user"`
}

// User is the user informations.
type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Job       string `json:"job"`
}

// Me get informations related to the user.
func (c *Client) Me(ctx context.Context) (MeResponse, error) {
	var res MeResponse
	err := c.gqlClient.Run(ctx, makeMeRequest(c.token), &res)
	if err != nil {
		return MeResponse{}, errors.Wrap(err, "me request failed")
	}
	return res, nil
}

// AddMachineLogResponse is the response sent by the server
// after an AddMachineLog mutation.
type AddMachineLogResponse struct {
	AddMachineLog BaseResult `json:"addMachineLog"`
}

// BaseResult contains a success boolean which is true if the server
// received the log successfully otherwise it is set to false.
type BaseResult struct {
	Success bool `json:"success"`
}

// MachineLogInput is the input type required by the server for machine's logs.
type MachineLogInput struct {
	Timestamp float64 `json:"timestamp"`
	MachineID string  `json:"machineId"`
	UserID    string  `json:"userId"`
	Log       string  `json:"log"`
}

// AddMachineLog send session's recorded log.
func (c *Client) AddMachineLog(ctx context.Context, inputs []MachineLogInput) (AddMachineLogResponse, error) {
	var res AddMachineLogResponse
	err := c.gqlClient.Run(ctx, makeAddMachineLogRequest(c.token, inputs), &res)
	if err != nil {
		return AddMachineLogResponse{}, errors.Wrap(err, "addMachineLog request failed")
	}
	return res, nil
}
