//go:build ignore

package main

import (
	"context"
	"fmt"

	"github.com/dnd-mcp/server/pkg/config"
	"github.com/dnd-mcp/server/internal/store/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	client, err := postgres.NewClient(cfg.Postgres)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ctx := context.Background()

	// Check current database
	var dbName string
	client.Pool().QueryRow(ctx, "SELECT current_database();").Scan(&dbName)
	fmt.Println("Current database via Pool:", dbName)

	// Check migrations
	rows, _ := client.Pool().Query(ctx, "SELECT version, applied_at FROM schema_migrations ORDER BY version;")
	fmt.Println("Migrations via Pool:")
	hasMigrations := false
	for rows.Next() {
		var version int64
		var time string
		rows.Scan(&version, &time)
		fmt.Printf(" - %d: %s\n", version, time)
		hasMigrations = true
	}
	rows.Close()
	if !hasMigrations {
		fmt.Println(" (no migrations)")
	}

	// Run migrations
	fmt.Println("Running migrations...")
	migrator := postgres.NewMigrator(client)
	if err := migrator.Up(ctx); err != nil {
		fmt.Println("Migration error:", err)
	}
}
