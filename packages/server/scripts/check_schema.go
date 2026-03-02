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

	// Check schema_migrations table schema
	var columnInfo string
	err = conn.QueryRow(ctx, "SELECT column_name, data_type FROM information_schema.columns WHERE table_name='schema_migrations' AND column_name='version';").Scan(&columnInfo, &columnInfo)
	if err != nil {
		fmt.Printf("Query error: %v\n", err)
		// Try alternative query
		rows, _ := conn.Query(ctx, "SELECT version, applied_at FROM schema_migrations;")
		defer rows.Close()
		fmt.Println("Rows from schema_migrations:")
		for rows.Next() {
			var version int64
			var appliedAt string
			if err := rows.Scan(&version, &appliedAt); err != nil {
				fmt.Printf("Scan error: %v\n", err)
			} else {
				fmt.Printf(" version=%d (%T), applied_at=%s (%T)\n", version, version, appliedAt, appliedAt)
			}
		}
		return
	}
	fmt.Println("Version column type:", columnInfo)
}
