package tailscale

import (
	"encoding/json"
	"fmt"
	"time"
)

// Device representa un dispositivo en la red Tailscale
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

// DevicesResponse representa la respuesta de la API al listar dispositivos
type DevicesResponse struct {
	Devices []Device `json:"devices"`
}

// ListDevices obtiene la lista de dispositivos conectados a la red Tailscale
func (c *Client) ListDevices() ([]Device, error) {
	endpoint := fmt.Sprintf("/tailnet/%s/devices", c.tailnet)

	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error listing devices: %w", err)
	}

	var devicesResp DevicesResponse
	if err := json.Unmarshal(respBody, &devicesResp); err != nil {
		return nil, fmt.Errorf("error parsing devices response: %w", err)
	}

	return devicesResp.Devices, nil
}

// GetDevice obtiene información de un dispositivo específico
func (c *Client) GetDevice(deviceID string) (*Device, error) {
	endpoint := fmt.Sprintf("/device/%s", deviceID)

	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting device: %w", err)
	}

	var device Device
	if err := json.Unmarshal(respBody, &device); err != nil {
		return nil, fmt.Errorf("error parsing device response: %w", err)
	}

	return &device, nil
}

// DeleteDevice elimina un dispositivo de la red
func (c *Client) DeleteDevice(deviceID string) error {
	endpoint := fmt.Sprintf("/device/%s", deviceID)

	_, err := c.doRequest("DELETE", endpoint, nil)
	if err != nil {
		return fmt.Errorf("error deleting device: %w", err)
	}

	return nil
}
