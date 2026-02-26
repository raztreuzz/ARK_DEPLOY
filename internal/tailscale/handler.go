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

func (h *Handler) CurrentDevice(c *gin.Context) {
	devices, err := h.client.ListDevices()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"detail": err.Error()})
		return
	}

	clientIP := strings.TrimSpace(c.ClientIP())
	normalize := func(v string) string {
		return strings.TrimSpace(strings.Split(strings.TrimSpace(v), "/")[0])
	}

	match := func(d Device) bool {
		for _, a := range d.Addresses {
			if normalize(a) == clientIP {
				return true
			}
		}
		return false
	}

	for _, d := range devices {
		if match(d) {
			targetHost := ""
			for _, a := range d.Addresses {
				n := normalize(a)
				if strings.HasPrefix(n, "100.") {
					targetHost = n
					break
				}
			}
			if targetHost == "" && len(d.Addresses) > 0 {
				targetHost = normalize(d.Addresses[0])
			}
			c.JSON(http.StatusOK, gin.H{
				"found":          true,
				"client_ip":      clientIP,
				"target_host":    targetHost,
				"tailscale_user": resolveDeviceUser(d),
				"device":         d,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"found":     false,
		"client_ip": clientIP,
		"detail":    "device not matched by client ip",
	})
}

func resolveDeviceUser(d Device) string {
	tryAny := func(v any) string {
		switch t := v.(type) {
		case string:
			return strings.TrimSpace(t)
		case map[string]any:
			for _, k := range []string{"login", "loginName", "name", "email", "displayName"} {
				if raw, ok := t[k]; ok {
					if s, ok := raw.(string); ok && strings.TrimSpace(s) != "" {
						return strings.TrimSpace(s)
					}
				}
			}
		}
		return ""
	}

	if s := tryAny(d.User); s != "" {
		return s
	}
	if s := tryAny(d.Owner); s != "" {
		return s
	}
	if strings.TrimSpace(d.Name) != "" {
		return strings.TrimSpace(d.Name)
	}
	return strings.TrimSpace(d.Hostname)
}
