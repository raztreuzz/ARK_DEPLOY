package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"ark_deploy/internal/config"
	"ark_deploy/internal/server"
	"ark_deploy/internal/storage"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	store := storage.NewProductStore()

	r := gin.Default()
	server.RegisterRoutes(r, cfg, store)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
