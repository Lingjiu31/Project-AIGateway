package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"ai-gateway/internal/config"
)

func NewRedis(cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close() // Ping 失败，释放连接池资源
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return rdb, nil
}
