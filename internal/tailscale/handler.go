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

// CreateAuthKeyRequest representa la petición para crear una auth key
type CreateAuthKeyRequest struct {
	Description   string   `json:"description"`
	Reusable      bool     `json:"reusable"`
	Ephemeral     bool     `json:"ephemeral"`
	Preauthorized bool     `json:"preauthorized"`
	ExpirySeconds int      `json:"expiry_seconds"`
	Tags          []string `json:"tags,omitempty"`
}

// CreateAuthKey maneja POST /tailscale/auth-keys
// Crea una auth key para registrar nuevos dispositivos
func (h *Handler) CreateAuthKey(c *gin.Context) {
	var req CreateAuthKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"detail": err.Error(),
		})
		return
	}

	// Valores por defecto y sanitización
	if req.Description == "" {
		req.Description = "Auth key created via ARK_DEPLOY"
	}
	// Sanitizar la descripción antes de enviar a Tailscale
	req.Description = SanitizeDescription(req.Description)
	
	if req.ExpirySeconds <= 0 {
		req.ExpirySeconds = 3600 // 1 hora
	}

	authKey, err := h.client.CreateAuthKey(
		req.Description,
		req.Reusable,
		req.Ephemeral,
		req.Preauthorized,
		req.ExpirySeconds,
		req.Tags,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create auth key",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"auth_key": authKey.Key,
		"id":       authKey.ID,
		"created":  authKey.Created,
		"expires":  authKey.Expires,
		"instructions": "Install Tailscale on the device and run: tailscale up --authkey=" + authKey.Key,
	})
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
