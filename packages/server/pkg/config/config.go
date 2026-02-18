// Package config provides application configuration management
// Supports environment variables and .env files, with environment variables taking priority
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config application configuration
type Config struct {
	Postgres PostgresConfig `json:"postgres"`
	HTTP     HTTPConfig     `json:"http"`
	Log      LogConfig      `json:"log"`
	RAG      RAGConfig      `json:"rag"`
}

// PostgresConfig PostgreSQL configuration
type PostgresConfig struct {
	Host            string `json:"host" env:"POSTGRES_HOST"`
	Port            int    `json:"port" env:"POSTGRES_PORT"`
	User            string `json:"user" env:"POSTGRES_USER"`
	Password        string `json:"-" env:"POSTGRES_PASSWORD"` // hide in JSON output
	DBName          string `json:"dbname" env:"POSTGRES_DBNAME"`
	SSLMode         string `json:"sslmode" env:"POSTGRES_SSLMODE"`
	PoolSize        int    `json:"pool_size" env:"POSTGRES_POOL_SIZE"`
	MaxConnLifetime int    `json:"max_conn_lifetime" env:"POSTGRES_MAX_CONN_LIFETIME"` // seconds
	MaxConnIdleTime int    `json:"max_conn_idletime" env:"POSTGRES_MAX_CONN_IDLETIME"` // seconds
}

// DSN returns the PostgreSQL connection string
func (c *PostgresConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// LogConfig logging configuration
type LogConfig struct {
	Level  string `json:"level" env:"LOG_LEVEL"`
	Format string `json:"format" env:"LOG_FORMAT"` // json or text
}

// HTTPConfig HTTP server configuration
type HTTPConfig struct {
	Host            string `json:"host" env:"HTTP_HOST"`
	Port            int    `json:"port" env:"HTTP_PORT"`
	ReadTimeout     int    `json:"read_timeout" env:"HTTP_READ_TIMEOUT"`           // seconds
	WriteTimeout    int    `json:"write_timeout" env:"HTTP_WRITE_TIMEOUT"`         // seconds
	ShutdownTimeout int    `json:"shutdown_timeout" env:"HTTP_SHUTDOWN_TIMEOUT"`   // seconds
	EnableCORS      bool   `json:"enable_cors" env:"HTTP_ENABLE_CORS"`
}

// RAGConfig RAG (Retrieval-Augmented Generation) configuration
type RAGConfig struct {
	Enabled bool   `json:"enabled" env:"RAG_ENABLED"`
	URL     string `json:"url" env:"RAG_URL"`
	Timeout int    `json:"timeout" env:"RAG_TIMEOUT"` // seconds
}

// Load loads configuration from environment variables and .env file
// Priority: environment variables > .env file > default values
func Load() (*Config, error) {
	// Try to load .env file (if exists)
	_ = godotenv.Load() // ignore error, .env file is optional

	cfg := &Config{
		Postgres: PostgresConfig{
			Host:            getEnv("POSTGRES_HOST", "localhost"),
			Port:            getEnvInt("POSTGRES_PORT", 5432),
			User:            getEnv("POSTGRES_USER", "dnd"),
			Password:        getEnv("POSTGRES_PASSWORD", "password"),
			DBName:          getEnv("POSTGRES_DBNAME", "dnd_server"),
			SSLMode:         getEnv("POSTGRES_SSLMODE", "disable"),
			PoolSize:        getEnvInt("POSTGRES_POOL_SIZE", 10),
			MaxConnLifetime: getEnvInt("POSTGRES_MAX_CONN_LIFETIME", 3600),
			MaxConnIdleTime: getEnvInt("POSTGRES_MAX_CONN_IDLETIME", 1800),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		HTTP: HTTPConfig{
			Host:            getEnv("HTTP_HOST", "0.0.0.0"),
			Port:            getEnvInt("HTTP_PORT", 8081),
			ReadTimeout:     getEnvInt("HTTP_READ_TIMEOUT", 30),
			WriteTimeout:    getEnvInt("HTTP_WRITE_TIMEOUT", 30),
			ShutdownTimeout: getEnvInt("HTTP_SHUTDOWN_TIMEOUT", 10),
			EnableCORS:      getEnvBool("HTTP_ENABLE_CORS", true),
		},
		RAG: RAGConfig{
			Enabled: getEnvBool("RAG_ENABLED", false),
			URL:     getEnv("RAG_URL", ""),
			Timeout: getEnvInt("RAG_TIMEOUT", 30),
		},
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate PostgreSQL configuration
	if c.Postgres.Host == "" {
		return fmt.Errorf("postgres host cannot be empty")
	}
	if c.Postgres.User == "" {
		return fmt.Errorf("postgres user cannot be empty")
	}
	if c.Postgres.DBName == "" {
		return fmt.Errorf("postgres dbname cannot be empty")
	}
	if c.Postgres.PoolSize <= 0 {
		return fmt.Errorf("postgres pool size must be greater than 0")
	}
	if c.Postgres.MaxConnLifetime <= 0 {
		return fmt.Errorf("postgres max conn lifetime must be greater than 0")
	}
	if c.Postgres.MaxConnIdleTime <= 0 {
		return fmt.Errorf("postgres max conn idletime must be greater than 0")
	}

	// Validate log configuration
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}
	if !validLogLevels[strings.ToLower(c.Log.Level)] {
		return fmt.Errorf("invalid log level: %s", c.Log.Level)
	}
	if c.Log.Format != "json" && c.Log.Format != "text" {
		return fmt.Errorf("invalid log format: %s (must be json or text)", c.Log.Format)
	}

	// Validate HTTP configuration
	if c.HTTP.Port <= 0 || c.HTTP.Port > 65535 {
		return fmt.Errorf("HTTP port must be in range 1-65535")
	}
	if c.HTTP.ReadTimeout <= 0 {
		return fmt.Errorf("HTTP read timeout must be greater than 0")
	}
	if c.HTTP.WriteTimeout <= 0 {
		return fmt.Errorf("HTTP write timeout must be greater than 0")
	}
	if c.HTTP.ShutdownTimeout <= 0 {
		return fmt.Errorf("HTTP shutdown timeout must be greater than 0")
	}

	// Validate RAG configuration (if enabled)
	if c.RAG.Enabled && c.RAG.URL == "" {
		return fmt.Errorf("RAG URL cannot be empty when RAG is enabled")
	}
	if c.RAG.Timeout <= 0 {
		return fmt.Errorf("RAG timeout must be greater than 0")
	}

	return nil
}

// getEnv gets an environment variable, returns default value if not exists
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an integer environment variable, returns default value if not exists or conversion fails
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvBool gets a boolean environment variable, returns default value if not exists or conversion fails
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
