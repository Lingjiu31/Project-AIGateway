package router

import (
	"fmt"
	"time"

	"ai-gateway/internal/breaker"
	"ai-gateway/internal/config"
)

type Router struct {
	models   map[string]config.ModelConfig
	breakers map[string]*breaker.CircuitBreaker
}

func NewRouter(models []config.ModelConfig, cbCfg config.CircuitBreakerConfig) *Router {
	r := &Router{
		models:   make(map[string]config.ModelConfig),
		breakers: make(map[string]*breaker.CircuitBreaker),
	}
	for _, model := range models {
		r.models[model.Name] = model
		r.breakers[model.Name] = breaker.New(cbCfg.MaxFailures, time.Duration(cbCfg.TimeoutSeconds)*time.Second)
	}
	return r
}

// Get 根据 model 名字返回配置和熔断器，找不到返回 error
func (r *Router) Get(name string) (config.ModelConfig, *breaker.CircuitBreaker, error) {
	model, ok := r.models[name]
	if !ok {
		return config.ModelConfig{}, nil, fmt.Errorf("未知模型: %s", name)
	}
	return model, r.breakers[name], nil
}
