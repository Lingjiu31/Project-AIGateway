package main

import (
	"ai-gateway/internal/auth"
	"ai-gateway/internal/proxy"
	"fmt"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/config"
	"ai-gateway/internal/db"
	"ai-gateway/internal/user"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		panic(err)
	}

	gormDB, err := db.NewMySQL(cfg.MySQL)
	if err != nil {
		panic(err)
	}

	if err := db.Migrate(gormDB, &user.User{}); err != nil {
		panic(err)
	}

	handler := proxy.NewHandler(cfg.Upstream.Models[0])
	manager := auth.NewManager(cfg.JWT.Secret)
	store := user.NewStore(gormDB)
	userHandler := user.NewHandler(store, manager)

	r := gin.New()
	setupRoutes(r, manager, handler, userHandler)

	if err := r.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		panic(err)
	}
}
