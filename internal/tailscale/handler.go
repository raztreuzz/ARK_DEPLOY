package tailscale

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ClientAPI interface {
	ListDevices() ([]Device, error)
}

type Handler struct {
	client ClientAPI
}

func NewHandler(client ClientAPI) *Handler {
	if client == nil {
		panic("tailscale client is required")
	}
	return &Handler{client: client}
}

func (h *Handler) ListDevices(c *gin.Context) {
	devices, err := h.client.ListDevices()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":   len(devices),
		"devices": devices,
	})
}

func (h *Handler) GetDevice(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "device id is required"})
		return
	}

	devices, err := h.client.ListDevices()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"detail": err.Error()})
		return
	}

	for _, d := range devices {
		if d.ID == id {
			c.JSON(http.StatusOK, d)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"detail": "device not found"})
}