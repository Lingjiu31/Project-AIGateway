package main

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"ai-gateway/internal/auth"
	"ai-gateway/internal/config"
	"ai-gateway/internal/db"
	"ai-gateway/internal/middleware"
	"ai-gateway/internal/proxy"
	"ai-gateway/internal/user"
)

func setupRoutes(r *gin.Engine, cfg *config.Config) {
	// CORS：允许本地测试页跨域访问
	r.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:  []string{"Content-Type", "Authorization"},
		ExposeHeaders: []string{"Content-Type"},
		MaxAge:        12 * time.Hour,
	}))

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
