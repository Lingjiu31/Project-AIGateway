package main

import (
	"ai-gateway/internal/auth"
	"ai-gateway/internal/middleware"
	"ai-gateway/internal/proxy"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/config"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		panic(err)
	}

	r := gin.New()
	setupRoutes(r, cfg)

	if err := r.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		panic(err)
	}
}

func setupRoutes(r *gin.Engine, cfg *config.Config) {
	handler := proxy.NewHandler(cfg.Upstream.Models[0])
	manager := auth.NewManager(cfg.JWT.Secret)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.POST("/v1/chat/completions", middleware.Auth(manager), handler.Handle)
}
