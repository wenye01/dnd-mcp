// Package postgres provides PostgreSQL storage implementations
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/dnd-mcp/server/pkg/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

// Client PostgreSQL client with connection pool
type Client struct {
	pool *pgxpool.Pool
}

// NewClient creates a PostgreSQL client
func NewClient(cfg config.PostgresConfig) (*Client, error) {
	// Build connection string
	connStr := buildConnectionString(cfg)

	// Configure connection pool
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Set pool parameters
	poolConfig.MaxConns = int32(cfg.PoolSize)
	poolConfig.MinConns = 1
	poolConfig.MaxConnLifetime = time.Duration(cfg.MaxConnLifetime) * time.Second
	poolConfig.MaxConnIdleTime = time.Duration(cfg.MaxConnIdleTime) * time.Second
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	// Set default query exec mode for proper type scanning
	// Use DescribeExec mode for binary protocol and proper type handling
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeDescribeExec

	// Configure logging (optional, for debugging)
	poolConfig.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   &pgxLogger{},
		LogLevel: tracelog.LogLevelWarn, // Use Warn in production
	}

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return &Client{pool: pool}, nil
}

// Ping checks connection status
func (c *Client) Ping(ctx context.Context) error {
	return c.pool.Ping(ctx)
}

// Close closes the connection
func (c *Client) Close() error {
	c.pool.Close()
	return nil
}

// Pool returns the connection pool
func (c *Client) Pool() *pgxpool.Pool {
	return c.pool
}

// buildConnectionString builds the connection string
func buildConnectionString(cfg config.PostgresConfig) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)
}

// pgxLogger implements pgx tracelog.Logger interface
type pgxLogger struct{}

func (l *pgxLogger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	// Simple logging implementation
	// In production, use a proper logging library
	_ = ctx
	_ = level
	_ = msg
	_ = data
}
