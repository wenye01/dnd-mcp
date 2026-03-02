//go:build ignore

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dnd-mcp/server/pkg/config"
	"github.com/dnd-mcp/server/internal/store/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Config DB: %s@%s:%d/%s\n", cfg.Postgres.User, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DBName)

	client, err := postgres.NewClient(cfg.Postgres)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Client error: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	ctx := context.Background()

	// Check current database
	var dbName string
	err = client.Pool().QueryRow(ctx, "SELECT current_database();").Scan(&dbName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Connected to database: %s\n", dbName)

	// Check migrations
	fmt.Println("Migrations:")
	rows, err := client.Pool().Query(ctx, "SELECT version, applied_at FROM schema_migrations ORDER BY version;")
	if err != nil {
		fmt.Printf("Query error: %v\n", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version int64
		var appliedAt string
		if err := rows.Scan(&version, &appliedAt); err != nil {
			fmt.Printf("Scan error: %v\n", err)
			continue
		}
		fmt.Printf(" - %d: %s\n", version, appliedAt)
	}
}
