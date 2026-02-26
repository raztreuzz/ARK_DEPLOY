package sshusers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"ark_deploy/internal/storage"
)

type Handler struct {
	store *storage.SSHUserStore
}

func NewHandler(store *storage.SSHUserStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) List(c *gin.Context) {
	m, err := h.store.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total": len(m),
		"map":   m,
	})
}

type upsertReq struct {
	SSHUser string `json:"ssh_user"`
}

func (h *Handler) Upsert(c *gin.Context) {
	host := strings.TrimSpace(c.Param("host"))
	if host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "host is required"})
		return
	}

	var req upsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	user := strings.TrimSpace(req.SSHUser)
	if user == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "ssh_user is required"})
		return
	}

	if err := h.store.Set(host, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "host": host, "ssh_user": user})
}

func (h *Handler) Delete(c *gin.Context) {
	host := strings.TrimSpace(c.Param("host"))
	if host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "host is required"})
		return
	}
	if err := h.store.Delete(host); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "host": host})
}

