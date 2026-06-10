package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server         ServerConfig         `mapstructure:"server"`
	Upstream       UpstreamConfig       `mapstructure:"upstream"`
	Redis          RedisConfig          `mapstructure:"redis"`
	MySQL          MySQLConfig          `mapstructure:"mysql"`
	JWT            JWTConfig            `mapstructure:"jwt"`
	RateLimit      RateLimitConfig      `mapstructure:"ratelimit"`
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuit_breaker"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type UpstreamConfig struct {
	Models []ModelConfig `mapstructure:"models"`
}

type ModelConfig struct {
	Name    string `mapstructure:"name"`
	BaseURL string `mapstructure:"base_url"`
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type MySQLConfig struct {
	DSN string `mapstructure:"dsn"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

type RateLimitConfig struct {
	Rate  float64 `mapstructure:"rate"`
	Burst int     `mapstructure:"burst"`
}

type CircuitBreakerConfig struct {
	MaxFailures    int `mapstructure:"max_failures"`    // 连续失败几次触发熔断
	TimeoutSeconds int `mapstructure:"timeout_seconds"` // 熔断后等多少秒进入 HalfOpen
}

// Load 从指定路径加载配置文件，支持环境变量覆盖
func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}
