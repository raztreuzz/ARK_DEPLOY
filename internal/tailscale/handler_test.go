package tailscale

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// MockClient simula el cliente de Tailscale para testing
type MockClient struct {
	devices      []Device
	shouldError  bool
	errorMessage string
}

func (m *MockClient) ListDevices() ([]Device, error) {
	if m.shouldError {
		return nil, assert.AnError
	}
	return m.devices, nil
}

func (m *MockClient) GetDevice(deviceID string) (*Device, error) {
	if m.shouldError {
		return nil, assert.AnError
	}
	for _, d := range m.devices {
		if d.ID == deviceID {
			return &d, nil
		}
	}
	return nil, assert.AnError
}

func (m *MockClient) DeleteDevice(deviceID string) error {
	if m.shouldError {
		return assert.AnError
	}
	return nil
}

func setupTest() (*gin.Engine, *Handler, *MockClient) {
	gin.SetMode(gin.TestMode)
	
	mockClient := &MockClient{
		devices: []Device{
			{
				ID:              "dev001",
				Name:            "server-01",
				Hostname:        "server-01.tail-scale.ts.net",
				Addresses:       []string{"100.64.0.1"},
				OS:              "linux",
				Created:         time.Now().Add(-24 * time.Hour),
				LastSeen:        time.Now(),
				IsOnline:        true,
				ClientVersion:   "1.56.0",
				UpdateAvailable: false,
				BlocksInbound:   false,
			},
			{
				ID:              "dev002",
				Name:            "laptop-dev",
				Hostname:        "laptop-dev.tail-scale.ts.net",
				Addresses:       []string{"100.64.0.2"},
				OS:              "windows",
				Created:         time.Now().Add(-48 * time.Hour),
				LastSeen:        time.Now().Add(-5 * time.Minute),
				IsOnline:        true,
				ClientVersion:   "1.56.0",
				UpdateAvailable: false,
				BlocksInbound:   true,
			},
		},
	}

	handler := &Handler{client: mockClient}
	router := gin.New()
	
	return router, handler, mockClient
}

func TestListDevices(t *testing.T) {
	router, handler, mockClient := setupTest()
	router.GET("/tailscale/devices", handler.ListDevices)

	t.Run("listar dispositivos exitosamente", func(t *testing.T) {
		mockClient.shouldError = false
		
		req := httptest.NewRequest("GET", "/tailscale/devices", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), response["count"])
		
		devices := response["devices"].([]interface{})
		assert.Len(t, devices, 2)
		
		firstDevice := devices[0].(map[string]interface{})
		assert.Equal(t, "dev001", firstDevice["id"])
		assert.Equal(t, "server-01", firstDevice["name"])
		assert.Equal(t, true, firstDevice["online"])
	})

	t.Run("error al listar dispositivos", func(t *testing.T) {
		mockClient.shouldError = true
		
		req := httptest.NewRequest("GET", "/tailscale/devices", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Failed to list devices")
	})
}

func TestGetDevice(t *testing.T) {
	router, handler, mockClient := setupTest()
	router.GET("/tailscale/devices/:id", handler.GetDevice)

	t.Run("obtener dispositivo existente", func(t *testing.T) {
		mockClient.shouldError = false
		
		req := httptest.NewRequest("GET", "/tailscale/devices/dev001", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var device Device
		err := json.Unmarshal(w.Body.Bytes(), &device)
		assert.NoError(t, err)
		assert.Equal(t, "dev001", device.ID)
		assert.Equal(t, "server-01", device.Name)
		assert.Equal(t, "linux", device.OS)
	})

	t.Run("dispositivo no encontrado", func(t *testing.T) {
		mockClient.shouldError = true
		
		req := httptest.NewRequest("GET", "/tailscale/devices/dev999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("sin ID de dispositivo", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tailscale/devices/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestDeleteDevice(t *testing.T) {
	router, handler, mockClient := setupTest()
	router.DELETE("/tailscale/devices/:id", handler.DeleteDevice)

	t.Run("eliminar dispositivo exitosamente", func(t *testing.T) {
		mockClient.shouldError = false
		
		req := httptest.NewRequest("DELETE", "/tailscale/devices/dev001", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Device deleted successfully", response["message"])
		assert.Equal(t, "dev001", response["device_id"])
	})

	t.Run("error al eliminar dispositivo", func(t *testing.T) {
		mockClient.shouldError = true
		
		req := httptest.NewRequest("DELETE", "/tailscale/devices/dev001", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("sin ID de dispositivo", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/tailscale/devices/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
