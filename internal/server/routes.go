package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ark_deploy/internal/config"
	"ark_deploy/internal/deployments"
	"ark_deploy/internal/instances"
	"ark_deploy/internal/products"
	"ark_deploy/internal/storage"
	"ark_deploy/internal/tailscale"
)

func RegisterRoutes(r *gin.Engine, cfg config.Config, productStore *storage.ProductStore, instanceStore *storage.InstanceStore) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	routeStore := storage.NewRouteStore()
	instances.RegisterRoutes(r, routeStore)

	// API group
	api := r.Group("/api")

	// Products
	ph := products.NewHandler(productStore)
	api.POST("/products", ph.Create)
	api.GET("/products", ph.List)
	api.GET("/products/:id", ph.Get)
	api.PUT("/products/:id", ph.Update)
	api.DELETE("/products/:id", ph.Delete)

	// Deployments
	h := deployments.NewHandler(cfg, productStore, instanceStore)

	api.GET("/jobs", h.ListJobs)
	api.GET("/deployments", h.List)
	api.POST("/deployments", h.Create)
	api.GET("/deployments/:id/logs", h.GetLogs)
	api.DELETE("/deployments/:id", h.Delete)

	api.GET("/deployments/pending", h.PendingJobs)
	api.GET("/deployments/queue", h.QueueToBuild)
	api.GET("/deployments/job/:job/build/:build/status", h.BuildStatus)
	api.GET("/deployments/job/:job/build/:build/logs", h.BuildLogs)

	// Tailscale
	tsClient := tailscale.NewClient(cfg.TailscaleAPIKey, cfg.TailscaleTailnet)
	tsHandler := tailscale.NewHandler(tsClient)

	api.GET("/tailscale/devices", tsHandler.ListDevices)
	api.GET("/tailscale/devices/:id", tsHandler.GetDevice)
	api.DELETE("/tailscale/devices/:id", tsHandler.DeleteDevice)
}
