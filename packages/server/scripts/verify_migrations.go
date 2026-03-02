//go:build ignore

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "host=localhost port=5432 user=postgres password=070831 dbname=dnd_server sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, "SELECT version, applied_at FROM schema_migrations ORDER BY version;")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	fmt.Println("Schema migrations:")
	for rows.Next() {
		var version int64
		var appliedAt time.Time
		if err := rows.Scan(&version, &appliedAt); err != nil {
			fmt.Printf("Scan error: %v\n", err)
		} else {
			fmt.Printf(" - %d: %s\n", version, appliedAt.Format("2006-01-02 15:04:05"))
		}
	}
}
