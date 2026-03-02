//go:build ignore

package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "host=localhost port=5432 user=postgres password=070831 dbname=dnd_server sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	// List tables
	rows, _ := conn.Query(ctx, "SELECT tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename;")
	fmt.Println("Tables:")
	for rows.Next() {
		var name string
		rows.Scan(&name)
		fmt.Println(" -", name)
	}
	rows.Close()

	// Check migrations
	rows2, _ := conn.Query(ctx, "SELECT version, applied_at FROM schema_migrations ORDER BY version;")
	fmt.Println("Migrations:")
	for rows2.Next() {
		var version int64
		var time string
		rows2.Scan(&version, &time)
		fmt.Printf(" - %d: %s\n", version, time)
	}
	rows2.Close()
}
