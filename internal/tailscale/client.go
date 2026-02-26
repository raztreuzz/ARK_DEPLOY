package tailscale

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	tailnet    string
	httpClient *http.Client
	userAgent  string
}

func NewClient(apiKey, tailnet string) *Client {
	return &Client{
		baseURL:   "https://api.tailscale.com/api/v2",
		apiKey:    strings.TrimSpace(apiKey),
		tailnet:   strings.TrimSpace(tailnet),
		userAgent: "ark_deploy/1.0",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Tailnet() string {
	return c.tailnet
}

func (c *Client) doRequest(ctx context.Context, method, endpoint string, body any) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("tailscale: marshal body: %w", err)
		}
		reqBody = bytes.NewBuffer(b)
	}

	u := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, u, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("tailscale: new request %s %s: %w", method, endpoint, err)
	}

	req.SetBasicAuth(c.apiKey, "")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("tailscale: do %s %s: %w", method, endpoint, err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("tailscale: read response %s %s: %w", method, endpoint, err)
	}

	if resp.StatusCode >= 400 {
		return nil, resp.StatusCode, fmt.Errorf("tailscale: api error %s %s status=%d body=%s", method, endpoint, resp.StatusCode, string(b))
	}

	return b, resp.StatusCode, nil
}

func decodeJSON[T any](b []byte, out *T) error {
	return json.Unmarshal(b, out)
}