//go:build ignore

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dnd-mcp/server/pkg/config"
	"github.com/jackc/pgx/v5"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	// Connect to postgres database first to drop/create dnd_server
	connString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.SSLMode)
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	// Drop and recreate database
	dbName := cfg.Postgres.DBName
	// Force disconnect all users
	_, err = conn.Exec(ctx, fmt.Sprintf("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '%s' AND pid <> pg_backend_pid();", dbName))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to terminate connections: %v\n", err)
	}
	_, err = conn.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s;", dbName))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to drop database: %v\n", err)
		os.Exit(1)
	}
	_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s;", dbName))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create database: %v\n", err)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to drop schema: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Database reset successfully. Run migrations again.")
}
