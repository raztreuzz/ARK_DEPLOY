package tailscale

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler maneja las peticiones HTTP relacionadas con Tailscale
type Handler struct {
	client *Client
}

// NewHandler crea un nuevo handler de Tailscale
func NewHandler(client *Client) *Handler {
	return &Handler{
		client: client,
	}
}

// ListDevices maneja GET /tailscale/devices
func (h *Handler) ListDevices(c *gin.Context) {
	devices, err := h.client.ListDevices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list devices",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"devices": devices,
		"count":   len(devices),
	})
}

// GetDevice maneja GET /tailscale/devices/:id
func (h *Handler) GetDevice(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Device ID is required",
		})
		return
	}

	device, err := h.client.GetDevice(deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get device",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, device)
}

// DeleteDevice maneja DELETE /tailscale/devices/:id
func (h *Handler) DeleteDevice(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Device ID is required",
		})
		return
	}

	if err := h.client.DeleteDevice(deviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete device",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Device deleted successfully",
		"device_id": deviceID,
	})
}
