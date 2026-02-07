// agent/auth/device_flow.go
package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type DeviceAuthResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type DeviceFlowClient struct {
	serverURL string
	client    *http.Client
}

func NewDeviceFlowClient(serverURL string) *DeviceFlowClient {
	return &DeviceFlowClient{
		serverURL: serverURL,
		client:    sharedHTTPClient,
	}
}

func (c *DeviceFlowClient) InitiateDeviceAuth() (DeviceAuthResponse, error) {
	body := bytes.NewBufferString(`{"client_id":"edictflow-cli"}`)
	resp, err := c.client.Post(c.serverURL+"/api/v1/auth/device", "application/json", body)
	if err != nil {
		return DeviceAuthResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return DeviceAuthResponse{}, fmt.Errorf("server error: %s", string(bodyBytes))
	}

	var result DeviceAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return DeviceAuthResponse{}, err
	}
	return result, nil
}

func (c *DeviceFlowClient) PollForToken(deviceCode string, interval, expiresIn int) (TokenResponse, error) {
	deadline := time.Now().Add(time.Duration(expiresIn) * time.Second)
	pollInterval := time.Duration(interval) * time.Second

	for time.Now().Before(deadline) {
		body := bytes.NewBuffer(nil)
		json.NewEncoder(body).Encode(map[string]string{
			"device_code": deviceCode,
			"client_id":   "edictflow-cli",
		})

		resp, err := c.client.Post(c.serverURL+"/api/v1/auth/device/token", "application/json", body)
		if err != nil {
			time.Sleep(pollInterval)
			continue
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var token TokenResponse
			if err := json.Unmarshal(bodyBytes, &token); err != nil {
				return TokenResponse{}, err
			}
			return token, nil
		}

		var errResp ErrorResponse
		json.Unmarshal(bodyBytes, &errResp)

		switch errResp.Error {
		case "authorization_pending":
			time.Sleep(pollInterval)
			continue
		case "expired_token":
			return TokenResponse{}, fmt.Errorf("device code expired")
		default:
			return TokenResponse{}, fmt.Errorf("auth error: %s", errResp.Error)
		}
	}

	return TokenResponse{}, fmt.Errorf("timeout waiting for authorization")
}
