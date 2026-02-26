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
	ih := instances.NewHandler(routeStore)
	ih.RegisterRoutes(r)

	api := r.Group("/api")

	ph := products.NewHandler(productStore)
	api.POST("/products", ph.Create)
	api.GET("/products", ph.List)
	api.GET("/products/:id", ph.Get)
	api.PUT("/products/:id", ph.Update)
	api.DELETE("/products/:id", ph.Delete)

	dh := deployments.NewHandler(cfg, productStore, instanceStore)
	api.GET("/deployments", dh.List)
	api.POST("/deployments", dh.Create)
	api.GET("/deployments/:id/logs", dh.GetLogs)
	api.DELETE("/deployments/:id", dh.Delete)

	api.GET("/deployments/pending", dh.PendingJobs)
	api.GET("/deployments/job/:job/build/:build/status", dh.BuildStatus)
	api.GET("/deployments/job/:job/build/:build/logs", dh.BuildLogs)

	tsClient := tailscale.NewClient(cfg.TailscaleAPIKey, cfg.TailscaleTailnet)
	tsHandler := tailscale.NewHandler(tsClient)
	api.GET("/tailscale/devices", tsHandler.ListDevices)
	api.GET("/tailscale/devices/:id", tsHandler.GetDevice)
}
