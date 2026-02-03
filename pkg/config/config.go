// Package config 提供应用配置管理
// 支持环境变量和配置文件,环境变量优先级更高
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config 应用配置
type Config struct {
	Redis  RedisConfig  `mapstructure:"redis"`
	Log    LogConfig    `mapstructure:"log"`
	Server ServerConfig `mapstructure:"server"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host         string `mapstructure:"host" env:"REDIS_HOST" default:"localhost:6379"`
	Password     string `mapstructure:"password" env:"REDIS_PASSWORD" default:""`
	DB           int    `mapstructure:"db" env:"REDIS_DB" default:"0"`
	PoolSize     int    `mapstructure:"pool_size" env:"REDIS_POOL_SIZE" default:"10"`
	MinIdleConns int    `mapstructure:"min_idle_conns" env:"REDIS_MIN_IDLE_CONNS" default:"5"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level" env:"LOG_LEVEL" default:"info"`
	Format string `mapstructure:"format" env:"LOG_FORMAT" default:"json"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `mapstructure:"host" env:"SERVER_HOST" default:"0.0.0.0"`
	Port int    `mapstructure:"port" env:"SERVER_PORT" default:"8080"`
}

// Load 从环境变量加载配置
// 环境变量优先级最高,未设置时使用默认值
func Load() (*Config, error) {
	cfg := &Config{
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost:6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvInt("REDIS_DB", 0),
			PoolSize:     getEnvInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: getEnvInt("REDIS_MIN_IDLE_CONNS", 5),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnvInt("SERVER_PORT", 8080),
		},
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return cfg, nil
}

// Validate 验证配置是否有效
func (c *Config) Validate() error {
	// 验证 Redis 配置
	if c.Redis.Host == "" {
		return fmt.Errorf("redis host 不能为空")
	}
	if c.Redis.PoolSize <= 0 {
		return fmt.Errorf("redis pool size 必须大于 0")
	}
	if c.Redis.MinIdleConns < 0 {
		return fmt.Errorf("redis min idle conns 不能为负数")
	}
	if c.Redis.DB < 0 {
		return fmt.Errorf("redis db 不能为负数")
	}

	// 验证日志配置
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}
	if !validLogLevels[strings.ToLower(c.Log.Level)] {
		return fmt.Errorf("无效的 log level: %s", c.Log.Level)
	}

	// 验证服务器配置
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server port 必须在 1-65535 范围内")
	}

	return nil
}

// getEnv 获取环境变量,如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取整数类型的环境变量,如果不存在或转换失败则返回默认值
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
