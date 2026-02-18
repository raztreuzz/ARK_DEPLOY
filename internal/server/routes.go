package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ark_deploy/internal/config"
	"ark_deploy/internal/deployments"
	"ark_deploy/internal/products"
	"ark_deploy/internal/storage"
	"ark_deploy/internal/tailscale"
)

func RegisterRoutes(r *gin.Engine, cfg config.Config, store *storage.ProductStore) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Products
	ph := products.NewHandler(store)
	r.POST("/products", ph.Create)
	r.GET("/products", ph.List)
	r.GET("/products/:id", ph.Get)
	r.PUT("/products/:id", ph.Update)
	r.DELETE("/products/:id", ph.Delete)

	// Deployments
	h := deployments.NewHandler(cfg, store)

	r.GET("/jobs", h.ListJobs)

	r.POST("/deployments", h.Create)

	r.GET("/deployments/pending", h.PendingJobs)
	r.GET("/deployments/queue", h.QueueToBuild)
	r.GET("/deployments/job/:job/build/:build/status", h.BuildStatus)
	r.GET("/deployments/job/:job/build/:build/logs", h.BuildLogs)

	// Tailscale
	tsClient := tailscale.NewClient(cfg.TailscaleAPIKey, cfg.TailscaleTailnet)
	tsHandler := tailscale.NewHandler(tsClient)

	r.GET("/tailscale/devices", tsHandler.ListDevices)
	r.GET("/tailscale/devices/:id", tsHandler.GetDevice)
	r.POST("/tailscale/auth-keys", tsHandler.CreateAuthKey)
	r.DELETE("/tailscale/devices/:id", tsHandler.DeleteDevice)
}
