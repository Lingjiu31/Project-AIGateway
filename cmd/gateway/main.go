package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ai-gateway/internal/auth"
	"ai-gateway/internal/config"
	"ai-gateway/internal/db"
	"ai-gateway/internal/logger"
	"ai-gateway/internal/proxy"
	"ai-gateway/internal/user"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		panic(err)
	}

	log, err := logger.New(gin.Mode() != gin.ReleaseMode)
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(log)
	defer log.Sync()

	gormDB, err := db.NewMySQL(cfg.MySQL)
	if err != nil {
		zap.L().Fatal("连接数据库失败", zap.Error(err))
	}

	if err := db.Migrate(gormDB, &user.User{}); err != nil {
		zap.L().Fatal("建表失败", zap.Error(err))
	}

	handler := proxy.NewHandler(cfg.Upstream.Models[0])
	manager := auth.NewManager(cfg.JWT.Secret)
	store := user.NewStore(gormDB)
	userHandler := user.NewHandler(store, manager)

	r := gin.New()
	setupRoutes(r, manager, handler, userHandler)

	zap.L().Info("服务启动", zap.Int("端口", cfg.Server.Port))
	if err := r.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		zap.L().Fatal("服务异常退出", zap.Error(err))
	}
}
