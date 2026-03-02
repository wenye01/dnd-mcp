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

	// Clear schema_migrations and re-insert with correct types
	_, err = conn.Exec(ctx, "DELETE FROM schema_migrations;")
	if err != nil {
		fmt.Printf("Delete error: %v\n", err)
	}

	// Insert all migrations as applied
	for _, v := range []int64{1, 2, 3, 5, 6} {
		_, err = conn.Exec(ctx, "INSERT INTO schema_migrations (version, applied_at) VALUES ($1, NOW());", v)
		if err != nil {
			fmt.Printf("Insert error for version %d: %v\n", v, err)
		}
	}

	fmt.Println("Migrations fixed!")

	// Verify
	rows, _ := conn.Query(ctx, "SELECT version, applied_at FROM schema_migrations ORDER BY version;")
	fmt.Println("Current migrations:")
	for rows.Next() {
		var version int64
		var appliedAt string
		rows.Scan(&version, &appliedAt)
		fmt.Printf(" - %d: %s\n", version, appliedAt)
	}
	rows.Close()
}
