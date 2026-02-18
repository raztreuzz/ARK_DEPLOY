package tailscale

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client representa un cliente HTTP para la API de Tailscale
type Client struct {
	baseURL    string
	apiKey     string
	tailnet    string
	httpClient *http.Client
}

// NewClient crea un nuevo cliente de Tailscale
func NewClient(apiKey, tailnet string) *Client {
	return &Client{
		baseURL: "https://api.tailscale.com/api/v2",
		apiKey:  apiKey,
		tailnet: tailnet,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest ejecuta una petición HTTP con autenticación
func (c *Client) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.SetBasicAuth(c.apiKey, "")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
