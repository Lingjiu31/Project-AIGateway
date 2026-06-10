package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

func RateLimit(limiter Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get("userID")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		key := fmt.Sprintf("ratelimit:%v", userID)
		allowed, err := limiter.Allow(c.Request.Context(), key)
		if err != nil {
			// redis 挂了
			zap.L().Error("限流检查失败", zap.Error(err))
			c.Next() // 降级放行，别让限流故障影响正常业务
			return
		}
		if !allowed {
			zap.L().Warn("用户触发限流", zap.Any("userID", userID))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
