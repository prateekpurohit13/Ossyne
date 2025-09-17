package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ossyne/internal/services"
)

type APIClient struct {
	BaseURL string
	Client  *http.Client
	Keyring services.KeyringService
}

func NewAPIClient() *APIClient {
	return &APIClient{
		BaseURL: "http://localhost:8080/api",
		Client:  &http.Client{},
		Keyring: services.KeyringService{},
	}
}

func (c *APIClient) DoAuthenticatedRequest(method, endpoint string, payload interface{}) (*http.Response, error) {
	token, err := c.Keyring.GetToken()
	if err != nil {
		return nil, err
	}
	var body io.Reader
	if payload != nil {
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonPayload)
	}
	req, err := http.NewRequest(method, c.BaseURL+endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	return c.Client.Do(req)
}