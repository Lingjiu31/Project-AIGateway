package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// New 根据运行模式返回对应 logger
// dev=true  → 彩色可读格式（开发用）
// dev=false → JSON 结构化格式（生产用）
func New(dev bool) (*zap.Logger, error) {
	if dev {
		log, err := zap.NewDevelopment()
		if err != nil {
			return nil, fmt.Errorf("new development logger: %w", err)
		}
		return log, nil
	}
	log, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("new production logger: %w", err)
	}
	return log, nil
}
