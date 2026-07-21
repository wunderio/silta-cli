package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HubClient talks to the Silta hub REST API, optionally using a
// stored CLI bearer token for authenticated requests.
type HubClient struct {
	BaseURL string
	Token   string
	http    *http.Client
}

// NewHubClient builds a client for the given hub base URL.
func NewHubClient(baseURL, token string) *HubClient {
	return &HubClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// NewHubClientFromCredentials builds a client using stored credentials.
func NewHubClientFromCredentials() (*HubClient, *Credentials, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return nil, nil, err
	}
	return NewHubClient(creds.HubURL, creds.Token), creds, nil
}

// PostJSON sends a JSON body to the given path and decodes the JSON response
// into out (when non-nil). It returns the HTTP status code and any error.
func (c *HubClient) PostJSON(path string, body interface{}, out interface{}) (int, error) {
	return c.doJSON(http.MethodPost, path, body, out)
}

// GetJSON performs a GET request and decodes the JSON response into out.
func (c *HubClient) GetJSON(path string, out interface{}) (int, error) {
	return c.doJSON(http.MethodGet, path, nil, out)
}

func (c *HubClient) doJSON(method, path string, body interface{}, out interface{}) (int, error) {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return 0, err
		}
		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reader)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if out != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil && err != io.EOF {
			return resp.StatusCode, err
		}
		return resp.StatusCode, nil
	}

	if resp.StatusCode >= 400 {
		return resp.StatusCode, apiError(resp)
	}
	return resp.StatusCode, nil
}

// apiError extracts an error message from a non-2xx JSON response.
func apiError(resp *http.Response) error {
	var body struct {
		Error string `json:"error"`
	}
	data, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(data, &body); err == nil && body.Error != "" {
		return fmt.Errorf("%s", body.Error)
	}
	return fmt.Errorf("request failed with status %d", resp.StatusCode)
}
