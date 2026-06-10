package main

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"ai-gateway/internal/auth"
	"ai-gateway/internal/middleware"
	"ai-gateway/internal/proxy"
	"ai-gateway/internal/router"
	"ai-gateway/internal/user"
)

func setupRoutes(r *gin.Engine, manager *auth.Manager,
	handler *proxy.Handler, userHandler *user.Handler, limiter middleware.Limiter, rt *router.Router) {

	r.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:  []string{"Content-Type", "Authorization"},
		ExposeHeaders: []string{"Content-Type"},
		MaxAge:        12 * time.Hour,
	}))
	r.Use(middleware.Logger())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/models", func(c *gin.Context) {
		c.JSON(http.StatusOK, rt.States())
	})
	r.POST("/v1/chat/completions",
		middleware.Auth(manager),
		middleware.RateLimit(limiter),
		handler.Handle,
	)
	r.POST("/register", userHandler.Register)
	r.POST("/login", userHandler.Login)
}
