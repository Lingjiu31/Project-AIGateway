package main

import (
	"ai-gateway/internal/auth"
	"ai-gateway/internal/config"
	"ai-gateway/internal/db"
	"ai-gateway/internal/middleware"
	"ai-gateway/internal/proxy"
	"ai-gateway/internal/user"
	"net/http"

	"github.com/gin-gonic/gin"
)

func setupRoutes(r *gin.Engine, cfg *config.Config) {
	handler := proxy.NewHandler(cfg.Upstream.Models[0])
	manager := auth.NewManager(cfg.JWT.Secret)
	gormDB, err := db.NewMySQL(cfg.MySQL)
	if err != nil {
		panic(err)
	}
	store := user.NewStore(gormDB)
	userHandler := user.NewHandler(store, manager)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.POST("/v1/chat/completions", middleware.Auth(manager), handler.Handle)
	r.POST("/register", userHandler.Register)
	r.POST("/login", userHandler.Login)
}
