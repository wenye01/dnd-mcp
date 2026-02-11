// Package config 提供应用配置管理
// 支持环境变量和配置文件,环境变量优先级更高
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config 应用配置
type Config struct {
	Redis    RedisConfig    `mapstructure:"redis"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Log      LogConfig      `mapstructure:"log"`
	HTTP     HTTPConfig     `mapstructure:"http"`
	LLM      LLMConfig      `mapstructure:"llm"`
	MCP      MCPConfig      `mapstructure:"mcp"`
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

// HTTPConfig HTTP 服务器配置
type HTTPConfig struct {
	Host            string `mapstructure:"host" env:"HTTP_HOST" default:"0.0.0.0"`
	Port            int    `mapstructure:"port" env:"HTTP_PORT" default:"8080"`
	ReadTimeout     int    `mapstructure:"read_timeout" env:"HTTP_READ_TIMEOUT" default:"30"`         // seconds
	WriteTimeout    int    `mapstructure:"write_timeout" env:"HTTP_WRITE_TIMEOUT" default:"30"`       // seconds
	ShutdownTimeout int    `mapstructure:"shutdown_timeout" env:"HTTP_SHUTDOWN_TIMEOUT" default:"10"` // seconds
	EnableCORS      bool   `mapstructure:"enable_cors" env:"HTTP_ENABLE_CORS" default:"true"`
}

// PostgresConfig PostgreSQL 配置
type PostgresConfig struct {
	Host            string `mapstructure:"host" env:"POSTGRES_HOST" default:"localhost"`
	Port            int    `mapstructure:"port" env:"POSTGRES_PORT" default:"5432"`
	User            string `mapstructure:"user" env:"POSTGRES_USER" default:"dnd"`
	Password        string `mapstructure:"password" env:"POSTGRES_PASSWORD" default:"password"`
	DBName          string `mapstructure:"dbname" env:"POSTGRES_DBNAME" default:"dnd_client"`
	SSLMode         string `mapstructure:"sslmode" env:"POSTGRES_SSLMODE" default:"disable"`
	PoolSize        int    `mapstructure:"pool_size" env:"POSTGRES_POOL_SIZE" default:"5"`
	MaxConnLifetime int    `mapstructure:"max_conn_lifetime" env:"POSTGRES_MAX_CONN_LIFETIME" default:"3600"` // seconds
	MaxConnIdleTime int    `mapstructure:"max_conn_idletime" env:"POSTGRES_MAX_CONN_IDLETIME" default:"1800"` // seconds
}

// LLMConfig LLM 配置
type LLMConfig struct {
	Provider    string  `mapstructure:"provider" env:"LLM_PROVIDER" default:"mock"`
	APIKey      string  `mapstructure:"api_key" env:"LLM_API_KEY" default:""`
	Model       string  `mapstructure:"model" env:"LLM_MODEL" default:"gpt-4"`
	MaxTokens   int     `mapstructure:"max_tokens" env:"LLM_MAX_TOKENS" default:"4096"`
	Temperature float64 `mapstructure:"temperature" env:"LLM_TEMPERATURE" default:"0.7"`
	Timeout     int     `mapstructure:"timeout" env:"LLM_TIMEOUT" default:"30"` // seconds
}

// MCPConfig MCP 配置
type MCPConfig struct {
	ServerURL string `mapstructure:"server_url" env:"MCP_SERVER_URL" default:"mock://"` // mock:// or http://...
	Timeout   int    `mapstructure:"timeout" env:"MCP_TIMEOUT" default:"30"`            // seconds
}

// Load 从环境变量和.env文件加载配置
// 优先级: 环境变量 > .env文件 > 默认值
func Load() (*Config, error) {
	// 尝试加载 .env 文件（如果存在）
	_ = godotenv.Load() // 忽略错误，.env 文件是可选的

	cfg := &Config{
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost:6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvInt("REDIS_DB", 0),
			PoolSize:     getEnvInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: getEnvInt("REDIS_MIN_IDLE_CONNS", 5),
		},
		Postgres: PostgresConfig{
			Host:            getEnv("POSTGRES_HOST", "localhost"),
			Port:            getEnvInt("POSTGRES_PORT", 5432),
			User:            getEnv("POSTGRES_USER", "dnd"),
			Password:        getEnv("POSTGRES_PASSWORD", "password"),
			DBName:          getEnv("POSTGRES_DBNAME", "dnd_client"),
			SSLMode:         getEnv("POSTGRES_SSLMODE", "disable"),
			PoolSize:        getEnvInt("POSTGRES_POOL_SIZE", 5),
			MaxConnLifetime: getEnvInt("POSTGRES_MAX_CONN_LIFETIME", 3600),
			MaxConnIdleTime: getEnvInt("POSTGRES_MAX_CONN_IDLETIME", 1800),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		HTTP: HTTPConfig{
			Host:            getEnv("HTTP_HOST", "0.0.0.0"),
			Port:            getEnvInt("HTTP_PORT", 8080),
			ReadTimeout:     getEnvInt("HTTP_READ_TIMEOUT", 30),
			WriteTimeout:    getEnvInt("HTTP_WRITE_TIMEOUT", 30),
			ShutdownTimeout: getEnvInt("HTTP_SHUTDOWN_TIMEOUT", 10),
			EnableCORS:      getEnvBool("HTTP_ENABLE_CORS", true),
		},
		LLM: LLMConfig{
			Provider:    getEnv("LLM_PROVIDER", "mock"),
			APIKey:      getEnv("LLM_API_KEY", ""),
			Model:       getEnv("LLM_MODEL", "gpt-4"),
			MaxTokens:   getEnvInt("LLM_MAX_TOKENS", 4096),
			Temperature: getEnvFloat64("LLM_TEMPERATURE", 0.7),
			Timeout:     getEnvInt("LLM_TIMEOUT", 30),
		},
		MCP: MCPConfig{
			ServerURL: getEnv("MCP_SERVER_URL", "mock://"),
			Timeout:   getEnvInt("MCP_TIMEOUT", 30),
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
	if c.HTTP.Port <= 0 || c.HTTP.Port > 65535 {
		return fmt.Errorf("HTTP port 必须在 1-65535 范围内")
	}
	if c.HTTP.ReadTimeout <= 0 {
		return fmt.Errorf("HTTP read timeout 必须大于 0")
	}
	if c.HTTP.WriteTimeout <= 0 {
		return fmt.Errorf("HTTP write timeout 必须大于 0")
	}
	if c.HTTP.ShutdownTimeout <= 0 {
		return fmt.Errorf("HTTP shutdown timeout 必须大于 0")
	}

	// 验证 PostgreSQL 配置
	if c.Postgres.Host == "" {
		return fmt.Errorf("postgres host 不能为空")
	}
	if c.Postgres.User == "" {
		return fmt.Errorf("postgres user 不能为空")
	}
	if c.Postgres.DBName == "" {
		return fmt.Errorf("postgres dbname 不能为空")
	}
	if c.Postgres.PoolSize <= 0 {
		return fmt.Errorf("postgres pool size 必须大于 0")
	}
	if c.Postgres.MaxConnLifetime <= 0 {
		return fmt.Errorf("postgres max conn lifetime 必须大于 0")
	}
	if c.Postgres.MaxConnIdleTime <= 0 {
		return fmt.Errorf("postgres max conn idletime 必须大于 0")
	}

	// 验证 LLM 配置
	if c.LLM.Provider != "mock" && c.LLM.Provider != "openai" {
		return fmt.Errorf("无效的 LLM provider: %s", c.LLM.Provider)
	}

	if c.LLM.Provider == "openai" && c.LLM.APIKey == "" {
		return fmt.Errorf("OpenAI API key 不能为空")
	}

	if c.LLM.MaxTokens <= 0 {
		return fmt.Errorf("LLM max tokens 必须大于 0")
	}

	if c.LLM.Temperature < 0 || c.LLM.Temperature > 2 {
		return fmt.Errorf("LLM temperature 必须在 0-2 范围内")
	}

	if c.LLM.Timeout <= 0 {
		return fmt.Errorf("LLM timeout 必须大于 0")
	}

	// 验证 MCP 配置
	if c.MCP.ServerURL == "" {
		return fmt.Errorf("MCP server URL 不能为空")
	}

	if c.MCP.Timeout <= 0 {
		return fmt.Errorf("MCP timeout 必须大于 0")
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

// getEnvBool 获取布尔类型的环境变量,如果不存在或转换失败则返回默认值
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

// getEnvFloat64 获取浮点数类型的环境变量,如果不存在或转换失败则返回默认值
func getEnvFloat64(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}
