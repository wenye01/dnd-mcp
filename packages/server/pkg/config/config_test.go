package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_DefaultValues(t *testing.T) {
	// Clear all environment variables to test defaults
	envVars := []string{
		"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_USER", "POSTGRES_PASSWORD",
		"POSTGRES_DBNAME", "POSTGRES_SSLMODE", "POSTGRES_POOL_SIZE",
		"POSTGRES_MAX_CONN_LIFETIME", "POSTGRES_MAX_CONN_IDLETIME",
		"LOG_LEVEL", "LOG_FORMAT",
		"HTTP_HOST", "HTTP_PORT", "HTTP_READ_TIMEOUT", "HTTP_WRITE_TIMEOUT",
		"HTTP_SHUTDOWN_TIMEOUT", "HTTP_ENABLE_CORS",
		"RAG_ENABLED", "RAG_URL", "RAG_TIMEOUT",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify PostgreSQL defaults
	assert.Equal(t, "localhost", cfg.Postgres.Host)
	assert.Equal(t, 5432, cfg.Postgres.Port)
	assert.Equal(t, "dnd", cfg.Postgres.User)
	assert.Equal(t, "password", cfg.Postgres.Password)
	assert.Equal(t, "dnd_server", cfg.Postgres.DBName)
	assert.Equal(t, "disable", cfg.Postgres.SSLMode)
	assert.Equal(t, 10, cfg.Postgres.PoolSize)
	assert.Equal(t, 3600, cfg.Postgres.MaxConnLifetime)
	assert.Equal(t, 1800, cfg.Postgres.MaxConnIdleTime)

	// Verify Log defaults
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Format)

	// Verify HTTP defaults
	assert.Equal(t, "0.0.0.0", cfg.HTTP.Host)
	assert.Equal(t, 8081, cfg.HTTP.Port)
	assert.Equal(t, 30, cfg.HTTP.ReadTimeout)
	assert.Equal(t, 30, cfg.HTTP.WriteTimeout)
	assert.Equal(t, 10, cfg.HTTP.ShutdownTimeout)
	assert.True(t, cfg.HTTP.EnableCORS)

	// Verify RAG defaults
	assert.False(t, cfg.RAG.Enabled)
	assert.Equal(t, "", cfg.RAG.URL)
	assert.Equal(t, 30, cfg.RAG.Timeout)
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("POSTGRES_HOST", "db.example.com")
	os.Setenv("POSTGRES_PORT", "5433")
	os.Setenv("POSTGRES_USER", "admin")
	os.Setenv("POSTGRES_PASSWORD", "secret")
	os.Setenv("POSTGRES_DBNAME", "dnd_prod")
	os.Setenv("POSTGRES_SSLMODE", "require")
	os.Setenv("POSTGRES_POOL_SIZE", "20")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "text")
	os.Setenv("HTTP_PORT", "9090")
	os.Setenv("HTTP_ENABLE_CORS", "false")
	os.Setenv("RAG_ENABLED", "true")
	os.Setenv("RAG_URL", "http://rag.example.com")

	defer func() {
		os.Unsetenv("POSTGRES_HOST")
		os.Unsetenv("POSTGRES_PORT")
		os.Unsetenv("POSTGRES_USER")
		os.Unsetenv("POSTGRES_PASSWORD")
		os.Unsetenv("POSTGRES_DBNAME")
		os.Unsetenv("POSTGRES_SSLMODE")
		os.Unsetenv("POSTGRES_POOL_SIZE")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_FORMAT")
		os.Unsetenv("HTTP_PORT")
		os.Unsetenv("HTTP_ENABLE_CORS")
		os.Unsetenv("RAG_ENABLED")
		os.Unsetenv("RAG_URL")
	}()

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify PostgreSQL values from environment
	assert.Equal(t, "db.example.com", cfg.Postgres.Host)
	assert.Equal(t, 5433, cfg.Postgres.Port)
	assert.Equal(t, "admin", cfg.Postgres.User)
	assert.Equal(t, "secret", cfg.Postgres.Password)
	assert.Equal(t, "dnd_prod", cfg.Postgres.DBName)
	assert.Equal(t, "require", cfg.Postgres.SSLMode)
	assert.Equal(t, 20, cfg.Postgres.PoolSize)

	// Verify Log values from environment
	assert.Equal(t, "debug", cfg.Log.Level)
	assert.Equal(t, "text", cfg.Log.Format)

	// Verify HTTP values from environment
	assert.Equal(t, 9090, cfg.HTTP.Port)
	assert.False(t, cfg.HTTP.EnableCORS)

	// Verify RAG values from environment
	assert.True(t, cfg.RAG.Enabled)
	assert.Equal(t, "http://rag.example.com", cfg.RAG.URL)
}

func TestValidate_InvalidLogLevel(t *testing.T) {
	cfg := &Config{
		Postgres: PostgresConfig{
			Host:            "localhost",
			User:            "dnd",
			DBName:          "dnd",
			PoolSize:        10,
			MaxConnLifetime: 3600,
			MaxConnIdleTime: 1800,
		},
		Log: LogConfig{
			Level:  "invalid",
			Format: "json",
		},
		HTTP: HTTPConfig{
			Port:            8080,
			ReadTimeout:     30,
			WriteTimeout:    30,
			ShutdownTimeout: 10,
		},
		RAG: RAGConfig{
			Timeout: 30,
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log level")
}

func TestValidate_InvalidLogFormat(t *testing.T) {
	cfg := &Config{
		Postgres: PostgresConfig{
			Host:            "localhost",
			User:            "dnd",
			DBName:          "dnd",
			PoolSize:        10,
			MaxConnLifetime: 3600,
			MaxConnIdleTime: 1800,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "yaml",
		},
		HTTP: HTTPConfig{
			Port:            8080,
			ReadTimeout:     30,
			WriteTimeout:    30,
			ShutdownTimeout: 10,
		},
		RAG: RAGConfig{
			Timeout: 30,
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log format")
}

func TestValidate_InvalidHTTPPort(t *testing.T) {
	cfg := &Config{
		Postgres: PostgresConfig{
			Host:            "localhost",
			User:            "dnd",
			DBName:          "dnd",
			PoolSize:        10,
			MaxConnLifetime: 3600,
			MaxConnIdleTime: 1800,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
		HTTP: HTTPConfig{
			Port:            70000, // invalid port
			ReadTimeout:     30,
			WriteTimeout:    30,
			ShutdownTimeout: 10,
		},
		RAG: RAGConfig{
			Timeout: 30,
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP port")
}

func TestValidate_RAGEnabledWithoutURL(t *testing.T) {
	cfg := &Config{
		Postgres: PostgresConfig{
			Host:            "localhost",
			User:            "dnd",
			DBName:          "dnd",
			PoolSize:        10,
			MaxConnLifetime: 3600,
			MaxConnIdleTime: 1800,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
		HTTP: HTTPConfig{
			Port:            8080,
			ReadTimeout:     30,
			WriteTimeout:    30,
			ShutdownTimeout: 10,
		},
		RAG: RAGConfig{
			Enabled: true,
			URL:     "", // missing URL
			Timeout: 30,
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RAG URL cannot be empty")
}

func TestValidate_MissingPostgresHost(t *testing.T) {
	cfg := &Config{
		Postgres: PostgresConfig{
			Host:            "", // missing host
			User:            "dnd",
			DBName:          "dnd",
			PoolSize:        10,
			MaxConnLifetime: 3600,
			MaxConnIdleTime: 1800,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
		HTTP: HTTPConfig{
			Port:            8080,
			ReadTimeout:     30,
			WriteTimeout:    30,
			ShutdownTimeout: 10,
		},
		RAG: RAGConfig{
			Timeout: 30,
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "postgres host")
}

func TestPostgresConfig_DSN(t *testing.T) {
	cfg := &PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "dnd",
		Password: "secret",
		DBName:   "dnd_server",
		SSLMode:  "disable",
	}

	dsn := cfg.DSN()
	expected := "host=localhost port=5432 user=dnd password=secret dbname=dnd_server sslmode=disable"
	assert.Equal(t, expected, dsn)
}

func TestGetEnv(t *testing.T) {
	// Test with existing env var
	os.Setenv("TEST_KEY", "test_value")
	assert.Equal(t, "test_value", getEnv("TEST_KEY", "default"))
	os.Unsetenv("TEST_KEY")

	// Test with missing env var
	assert.Equal(t, "default", getEnv("NONEXISTENT_KEY", "default"))
}

func TestGetEnvInt(t *testing.T) {
	// Test with valid int
	os.Setenv("TEST_INT", "42")
	assert.Equal(t, 42, getEnvInt("TEST_INT", 0))
	os.Unsetenv("TEST_INT")

	// Test with invalid int
	os.Setenv("TEST_INT", "not_a_number")
	assert.Equal(t, 100, getEnvInt("TEST_INT", 100))
	os.Unsetenv("TEST_INT")

	// Test with missing env var
	assert.Equal(t, 50, getEnvInt("NONEXISTENT_INT", 50))
}

func TestGetEnvBool(t *testing.T) {
	// Use unique env var name to avoid conflicts
	envKey := "DND_SERVER_TEST_BOOL_12345"

	// Test with true values
	for _, val := range []string{"true", "TRUE", "1"} {
		os.Setenv(envKey, val)
		assert.True(t, getEnvBool(envKey, false), "expected true for value: %s", val)
		os.Unsetenv(envKey)
	}

	// Test with false values
	for _, val := range []string{"false", "FALSE", "0"} {
		os.Setenv(envKey, val)
		assert.False(t, getEnvBool(envKey, true), "expected false for value: %s", val)
		os.Unsetenv(envKey)
	}

	// Test with missing env var
	assert.True(t, getEnvBool("NONEXISTENT_BOOL_"+t.Name(), true))
	assert.False(t, getEnvBool("NONEXISTENT_BOOL_"+t.Name(), false))
}
