package tailscale

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Device struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Hostname        string    `json:"hostname"`
	Addresses       []string  `json:"addresses"`
	OS              string    `json:"os"`
	Created         time.Time `json:"created"`
	LastSeen        time.Time `json:"lastSeen"`
	IsOnline        bool      `json:"online"`
	ClientVersion   string    `json:"clientVersion,omitempty"`
	UpdateAvailable bool      `json:"updateAvailable"`
	BlocksInbound   bool      `json:"blocksIncomingConnections"`
}

type DevicesResponse struct {
	Devices []Device `json:"devices"`
}

func (c *Client) ListDevices() ([]Device, error) {
	endpoint := fmt.Sprintf("/tailnet/%s/devices", c.tailnet)

	b, _, err := c.doRequest(context.Background(), "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var resp DevicesResponse
	if err := decodeJSON(b, &resp); err != nil {
		return nil, fmt.Errorf("tailscale: decode devices response: %w", err)
	}

	return resp.Devices, nil
}

func (c *Client) FindDeviceByID(ctx context.Context, deviceID string) (*Device, error) {
	id := strings.TrimSpace(deviceID)
	if id == "" {
		return nil, fmt.Errorf("tailscale: deviceID is required")
	}

	devices, err := c.ListDevices()
	if err != nil {
		return nil, err
	}

	for i := range devices {
		if devices[i].ID == id {
			return &devices[i], nil
		}
	}

	return nil, fmt.Errorf("tailscale: device not found")
}

func (c *Client) FindDeviceByName(ctx context.Context, name string) (*Device, error) {
	n := strings.TrimSpace(name)
	if n == "" {
		return nil, fmt.Errorf("tailscale: name is required")
	}

	devices, err := c.ListDevices()
	if err != nil {
		return nil, err
	}

	for i := range devices {
		if devices[i].Name == n || devices[i].Hostname == n {
			return &devices[i], nil
		}
	}

	return nil, fmt.Errorf("tailscale: device not found")
}
