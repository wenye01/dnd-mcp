// Package store_test contains integration tests for the store package
package store_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/dnd-mcp/server/internal/store/postgres"
	"github.com/dnd-mcp/server/pkg/config"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getModuleRoot returns the module root directory
func getModuleRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)
	return filepath.Dir(filepath.Dir(filepath.Dir(testDir)))
}

// getMigrationsPath returns the absolute path to migrations directory
func getMigrationsPath() string {
	return filepath.Join(getModuleRoot(), "internal", "store", "postgres", "migrations")
}

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Load .env file from module root
	envPath := filepath.Join(getModuleRoot(), ".env")
	if _, err := os.Stat(envPath); err == nil {
		godotenv.Load(envPath)
	}

	// Set test database name if not specified
	if os.Getenv("POSTGRES_TEST_DBNAME") != "" {
		os.Setenv("POSTGRES_DBNAME", os.Getenv("POSTGRES_TEST_DBNAME"))
	} else if os.Getenv("POSTGRES_DBNAME") == "" {
		os.Setenv("POSTGRES_DBNAME", "dnd_server_test")
	}

	// Run tests
	code := m.Run()
	os.Exit(code)
}

// getTestConfig returns a test configuration
func getTestConfig() config.PostgresConfig {
	return config.PostgresConfig{
		Host:            getEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:            5432,
		User:            getEnvOrDefault("POSTGRES_USER", "dnd"),
		Password:        getEnvOrDefault("POSTGRES_PASSWORD", "password"),
		DBName:          getEnvOrDefault("POSTGRES_DBNAME", "dnd_server_test"),
		SSLMode:         "disable",
		PoolSize:        5,
		MaxConnLifetime: 3600,
		MaxConnIdleTime: 1800,
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// skipIfNoDatabase skips the test if database is not available
func skipIfNoDatabase(t *testing.T, client *postgres.Client) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		t.Skipf("Database not available: %v", err)
	}
}

func TestNewClient(t *testing.T) {
	cfg := getTestConfig()

	client, err := postgres.NewClient(cfg)
	require.NoError(t, err, "Failed to create client")
	require.NotNil(t, client, "Client should not be nil")

	defer client.Close()

	skipIfNoDatabase(t, client)

	// Test Ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Ping(ctx)
	assert.NoError(t, err, "Ping should succeed")
}

func TestClientPool(t *testing.T) {
	cfg := getTestConfig()

	client, err := postgres.NewClient(cfg)
	require.NoError(t, err, "Failed to create client")
	defer client.Close()

	skipIfNoDatabase(t, client)

	pool := client.Pool()
	assert.NotNil(t, pool, "Pool should not be nil")
}

func TestMigrator_NewMigrator(t *testing.T) {
	cfg := getTestConfig()

	client, err := postgres.NewClient(cfg)
	require.NoError(t, err, "Failed to create client")
	defer client.Close()

	skipIfNoDatabase(t, client)

	migrator := postgres.NewMigrator(client)
	assert.NotNil(t, migrator, "Migrator should not be nil")
}

func TestMigrator_EnsureSchemaMigrationsTable(t *testing.T) {
	cfg := getTestConfig()

	client, err := postgres.NewClient(cfg)
	require.NoError(t, err, "Failed to create client")
	defer client.Close()

	skipIfNoDatabase(t, client)

	// Use absolute path to migrations directory
	migrator := postgres.NewMigratorWithPath(client, getMigrationsPath())
	ctx := context.Background()

	// This is tested implicitly through Status()
	migrations, err := migrator.Status(ctx)
	require.NoError(t, err, "Status should succeed")
	assert.NotNil(t, migrations, "Migrations should not be nil")

	// Clean up schema_migrations table for other tests
	_, _ = client.Pool().Exec(ctx, "DROP TABLE IF EXISTS schema_migrations")
}

func TestBuildConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		cfg      config.PostgresConfig
		expected string
	}{
		{
			name: "standard config",
			cfg: config.PostgresConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "dnd",
				Password: "secret",
				DBName:   "dnd_server",
				SSLMode:  "disable",
			},
			expected: "host=localhost port=5432 user=dnd password=secret dbname=dnd_server sslmode=disable",
		},
		{
			name: "with SSL",
			cfg: config.PostgresConfig{
				Host:     "db.example.com",
				Port:     5433,
				User:     "admin",
				Password: "pass123",
				DBName:   "production",
				SSLMode:  "require",
			},
			expected: "host=db.example.com port=5433 user=admin password=pass123 dbname=production sslmode=require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: buildConnectionString is not exported, so we test via DSN method on config
			dsn := tt.cfg.DSN()
			assert.Equal(t, tt.expected, dsn)
		})
	}
}

func TestClient_ConnectionPoolSettings(t *testing.T) {
	cfg := getTestConfig()
	cfg.PoolSize = 3
	cfg.MaxConnLifetime = 600
	cfg.MaxConnIdleTime = 300

	client, err := postgres.NewClient(cfg)
	require.NoError(t, err, "Failed to create client")
	defer client.Close()

	skipIfNoDatabase(t, client)

	// Test that we can acquire and release connections
	ctx := context.Background()

	// Run multiple concurrent queries to test pool
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(id int) {
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			var result int
			err := client.Pool().QueryRow(ctx, "SELECT 1").Scan(&result)
			assert.NoError(t, err, "Query %d should succeed", id)
			assert.Equal(t, 1, result, "Query %d should return 1", id)
			done <- true
		}(i)
	}

	// Wait for all queries to complete
	for i := 0; i < 5; i++ {
		select {
		case <-done:
			// Good
		case <-time.After(15 * time.Second):
			t.Fatal("Timeout waiting for concurrent queries")
		}
	}
}

// Example of how to run integration tests:
// go test -v ./tests/integration/store/... -tags=integration
//
// To set up a test database:
// 1. Ensure PostgreSQL is running
// 2. Create test database: CREATE DATABASE dnd_server_test;
// 3. Grant permissions: GRANT ALL PRIVILEGES ON DATABASE dnd_server_test TO dnd;
//
// Note: Example test is commented out because it requires a database.
// Uncomment and configure for your environment to test manually.
/*
func ExampleNewClient() {
	cfg := config.PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "dnd",
		Password: "password",
		DBName:   "dnd_server",
		SSLMode:  "disable",
		PoolSize: 10,
	}

	client, err := postgres.NewClient(cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		fmt.Printf("Ping failed: %v\n", err)
		return
	}

	fmt.Println("Connected successfully!")
	// Output: Connected successfully!
}
*/
