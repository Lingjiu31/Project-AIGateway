package main

import (
	"fmt"

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
