//go:build ignore

package main

import (
	"context"
	"fmt"

	"github.com/dnd-mcp/server/pkg/config"
	"github.com/jackc/pgx/v5"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DBName, cfg.Postgres.SSLMode)
	fmt.Println("Connection string:", connStr)

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	fmt.Println("Connected!")

	// Check current database
	var dbName string
	conn.QueryRow(ctx, "SELECT current_database();").Scan(&dbName)
	fmt.Println("Current database:", dbName)

	// Check tables
	rows, _ := conn.Query(ctx, "SELECT tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename;")
	fmt.Println("Tables:")
	hasTables := false
	for rows.Next() {
		var name string
		rows.Scan(&name)
		fmt.Println(" -", name)
		hasTables = true
	}
	rows.Close()
	if !hasTables {
		fmt.Println(" (no tables)")
	}

	// Check migrations
	rows2, _ := conn.Query(ctx, "SELECT version, applied_at FROM schema_migrations ORDER BY version;")
	fmt.Println("Migrations:")
	hasMigrations := false
	for rows2.Next() {
		var version int64
		var time string
		rows2.Scan(&version, &time)
		fmt.Printf(" - %d: %s\n", version, time)
		hasMigrations = true
	}
	rows2.Close()
	if !hasMigrations {
		fmt.Println(" (no migrations)")
	}
}
