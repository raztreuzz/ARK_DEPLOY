package tailscale

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockClient struct {
	devices     []Device
	shouldError bool
}

func (m *MockClient) ListDevices() ([]Device, error) {
	if m.shouldError {
		return nil, assert.AnError
	}
	return m.devices, nil
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

	handler := NewHandler(mockClient)
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

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, float64(2), response["total"])

		devices := response["devices"].([]any)
		assert.Len(t, devices, 2)

		first := devices[0].(map[string]any)
		assert.Equal(t, "dev001", first["id"])
		assert.Equal(t, "server-01", first["name"])
		assert.Equal(t, true, first["online"])
	})

	t.Run("error al listar dispositivos", func(t *testing.T) {
		mockClient.shouldError = true

		req := httptest.NewRequest("GET", "/tailscale/devices", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadGateway, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		_, ok := response["detail"]
		assert.True(t, ok)
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
		mockClient.shouldError = false

		req := httptest.NewRequest("GET", "/tailscale/devices/dev999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "device not found", response["detail"])
	})

	t.Run("error del cliente al listar", func(t *testing.T) {
		mockClient.shouldError = true

		req := httptest.NewRequest("GET", "/tailscale/devices/dev001", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadGateway, w.Code)
	})

	t.Run("ruta sin id (gin devuelve 404)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tailscale/devices/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}