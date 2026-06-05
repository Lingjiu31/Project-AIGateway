package ratelimit

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed token_bucket.lua
var luaTokenBucket string

type TokenBucketLimiter struct {
	rdb   *redis.Client
	rate  float64
	burst int
}

func NewTokenBucketLimiter(rdb *redis.Client, rate float64, burst int) *TokenBucketLimiter {
	return &TokenBucketLimiter{rdb: rdb, rate: rate, burst: burst}
}

func (l *TokenBucketLimiter) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now().UnixMicro()

	result, err := l.rdb.Eval(ctx, luaTokenBucket,
		[]string{key},
		l.rate,
		l.burst,
		now,
	).Int()

	if err != nil {
		return false, fmt.Errorf("eval lua: %w", err)
	}
	return result == 1, nil
}
