// agent/auth/credentials.go
package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	User      struct {
		ID     string  `json:"id"`
		Email  string  `json:"email"`
		Name   string  `json:"name"`
		TeamID *string `json:"teamId,omitempty"`
	} `json:"user"`
}

type CredentialsClient struct {
	serverURL string
	client    *http.Client
}

func NewCredentialsClient(serverURL string) *CredentialsClient {
	return &CredentialsClient{
		serverURL: serverURL,
		client:    sharedHTTPClient,
	}
}

func (c *CredentialsClient) Login(email, password string) (LoginResponse, error) {
	reqBody := LoginRequest{
		Email:    email,
		Password: password,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return LoginResponse{}, err
	}

	resp, err := c.client.Post(c.serverURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return LoginResponse{}, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		return LoginResponse{}, fmt.Errorf("invalid email or password")
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(bodyBytes, &errResp) == nil && errResp.Error != "" {
			return LoginResponse{}, fmt.Errorf("login failed: %s", errResp.Error)
		}
		return LoginResponse{}, fmt.Errorf("login failed: %s", string(bodyBytes))
	}

	var result LoginResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return LoginResponse{}, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}
