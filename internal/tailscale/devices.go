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

// AuthKeyRequest representa la petición para crear una auth key
type AuthKeyRequest struct {
	Capabilities    AuthKeyCapabilities `json:"capabilities"`
	ExpirySeconds   int                 `json:"expirySeconds"`
	Description     string              `json:"description,omitempty"`
}

// AuthKeyCapabilities define las capacidades de una auth key
type AuthKeyCapabilities struct {
	Devices AuthKeyDeviceCapabilities `json:"devices"`
}

// AuthKeyDeviceCapabilities define las capacidades de dispositivos
type AuthKeyDeviceCapabilities struct {
	Create AuthKeyCreate `json:"create"`
}

// AuthKeyCreate define configuración de creación
type AuthKeyCreate struct {
	Reusable      bool     `json:"reusable"`
	Ephemeral     bool     `json:"ephemeral"`
	Preauthorized bool     `json:"preauthorized"`
	Tags          []string `json:"tags,omitempty"`
}

// AuthKeyResponse representa la respuesta al crear una auth key
type AuthKeyResponse struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Created     time.Time `json:"created"`
	Expires     time.Time `json:"expires"`
	Capabilities AuthKeyCapabilities `json:"capabilities"`
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

// CreateAuthKey crea una auth key para registrar nuevos dispositivos
func (c *Client) CreateAuthKey(description string, reusable, ephemeral, preauthorized bool, expirySeconds int, tags []string) (*AuthKeyResponse, error) {
	if expirySeconds <= 0 {
		expirySeconds = 3600 // 1 hora por defecto
	}

	req := AuthKeyRequest{
		Capabilities: AuthKeyCapabilities{
			Devices: AuthKeyDeviceCapabilities{
				Create: AuthKeyCreate{
					Reusable:      reusable,
					Ephemeral:     ephemeral,
					Preauthorized: preauthorized,
					Tags:          tags,
				},
			},
		},
		ExpirySeconds: expirySeconds,
		Description:   description,
	}

	endpoint := fmt.Sprintf("/tailnet/%s/keys", c.tailnet)
	
	respBody, err := c.doRequest("POST", endpoint, req)
	if err != nil {
		return nil, fmt.Errorf("error creating auth key: %w", err)
	}

	var authKeyResp AuthKeyResponse
	if err := json.Unmarshal(respBody, &authKeyResp); err != nil {
		return nil, fmt.Errorf("error parsing auth key response: %w", err)
	}

	return &authKeyResp, nil
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
