package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 应用配置
type Config struct {
	// 服务器配置
	ServerHost string
	ServerPort int

	// 数据库配置
	PostgreSQLURL string

	// Redis配置
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// LLM配置
	LLMProvider string
	LLMAPIKey  string
	LLMBaseURL string
	LLMModel   string

	// MCP Server配置
	MCPServerURL string

	// CORS配置
	AllowOrigins []string
}

// Load 加载配置
func Load() (*Config, error) {
	// 加载.env文件(如果存在)
	_ = godotenv.Load()

	cfg := &Config{
		ServerHost:   getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort:   getEnvInt("SERVER_PORT", 8080),
		PostgreSQLURL: getEnv("POSTGRESQL_URL", "postgres://localhost/dnd_mcp?sslmode=disable"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),
		LLMProvider:   getEnv("LLM_PROVIDER", "openai"),
		LLMAPIKey:     getEnv("LLM_API_KEY", ""),
		LLMBaseURL:    getEnv("LLM_BASE_URL", ""),
		LLMModel:      getEnv("LLM_MODEL", "gpt-4"),
		MCPServerURL:  getEnv("MCP_SERVER_URL", "http://localhost:8081"),
		AllowOrigins:  []string{getEnv("ALLOW_ORIGIN", "*")},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
