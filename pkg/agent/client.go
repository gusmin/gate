// Package agent contains a client for the Secure Gate agents API.
package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// Client is a REST client interacting with Secure Gate agents.
type Client struct {
	// contains filtered or unexported fields
	httpClient *http.Client
	authToken  string
}

// NewClient creates a new Secure Gate agent client with the given authorization token
// and httpClient. http.DefaultClient is used as HTTP client if none given.
func NewClient(authToken string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		httpClient: httpClient,
		authToken:  authToken,
	}
}

// SSHAuthResponse is the response sent by an agent after a POST request
// on the ssh-authorization route.
type SSHAuthResponse struct {
	ErrorType string `json:"ErrorType"`
	Message   string `json:"Message"`
}

// AddAuthorizedKey add the public SSH key to the authorized_keys file
// located on the agent running at the given endpoint for the given user id.
func (c Client) AddAuthorizedKey(ctx context.Context, endpoint, id string, key []byte) (SSHAuthResponse, error) {
	// Marshal the key as json body.
	body, err := json.Marshal(map[string]interface{}{"publicKey": strings.TrimSpace(string(key))})
	if err != nil {
		return SSHAuthResponse{}, errors.Wrap(err, "could not create body for ssh-authorization POST request")
	}

	req, err := makeAddAuthorizedKeyRequest(endpoint, id, c.authToken, bytes.NewBuffer(body))
	if err != nil {
		return SSHAuthResponse{}, err
	}

	var resp SSHAuthResponse
	err = c.do(ctx, req, &resp)
	if err != nil {
		return SSHAuthResponse{}, errors.Wrap(err, "ssh-authorization POST request failed")
	}
	return resp, nil
}

func makeAddAuthorizedKeyRequest(endpoint, id, token string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, endpoint+"/gate/users/"+id+"/ssh-authorization", body)
	if err != nil {
		return nil, errors.Wrap(err, "could not create ssh-authorization POST request")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// DeleteAuthorizedKey deletes the public SSH key from the authorized_keys file
// located on the agent running at the given endpoint for the given user id.
func (c Client) DeleteAuthorizedKey(ctx context.Context, endpoint, id string, key []byte) (SSHAuthResponse, error) {
	// Marshal the key as json body.
	body, err := json.Marshal(map[string]interface{}{"publicKey": strings.TrimSpace(string(key))})
	if err != nil {
		return SSHAuthResponse{}, errors.Wrap(err, "could not create body for ssh-authorization DELETE request")
	}

	req, err := makeDeleteSSHPublicKeyRequest(endpoint, id, c.authToken, bytes.NewBuffer(body))
	if err != nil {
		return SSHAuthResponse{}, err
	}

	var resp SSHAuthResponse
	err = c.do(ctx, req, &resp)
	if err != nil {
		return SSHAuthResponse{}, errors.Wrap(err, "ssh-authorization DELETE request failed")
	}
	return resp, nil
}

func makeDeleteSSHPublicKeyRequest(endpoint, id, token string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodDelete, endpoint+"/gate/users/"+id+"/ssh-authorization", body)
	if err != nil {
		return nil, errors.Wrap(err, "could not create ssh-authorization DELETE request")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c Client) do(ctx context.Context, req *http.Request, resp interface{}) error {
	httpResp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("request status code is not 200: %d", httpResp.StatusCode)
	}

	err = json.NewDecoder(httpResp.Body).Decode(&resp)
	if err != nil {
		return errors.Wrap(err, "could not decode response body")
	}
	return nil
}
