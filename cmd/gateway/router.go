package main

import (
	"net/http"
	"time"

	"ai-gateway/internal/auth"
	"ai-gateway/internal/middleware"
	"ai-gateway/internal/proxy"
	"ai-gateway/internal/user"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func setupRoutes(r *gin.Engine, manager *auth.Manager,
	handler *proxy.Handler, userHandler *user.Handler) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:  []string{"Content-Type", "Authorization"},
		ExposeHeaders: []string{"Content-Type"},
		MaxAge:        12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.POST("/v1/chat/completions", middleware.Auth(manager), handler.Handle)
	r.POST("/register", userHandler.Register)
	r.POST("/login", userHandler.Login)
}
