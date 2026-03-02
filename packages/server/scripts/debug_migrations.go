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

	// Clear and check
	_, err = conn.Exec(ctx, "DELETE FROM schema_migrations;")
	if err != nil {
		fmt.Printf("Delete error: %v\n", err)
	}

	// Try insert one by one with detailed output
	for _, v := range []int64{1, 2, 3, 5, 6} {
		result, err := conn.Exec(ctx, "INSERT INTO schema_migrations (version, applied_at) VALUES ($1, NOW());", v)
		if err != nil {
			fmt.Printf("Insert error for version %d: %v\n", v, err)
		} else {
			fmt.Printf("Inserted version %d, rows affected: %d\n", v, result.RowsAffected())
		}
	}

	// Verify
	rows, _ := conn.Query(ctx, "SELECT version, applied_at FROM schema_migrations ORDER BY version;")
	fmt.Println("Current migrations:")
	count := 0
	for rows.Next() {
		var version int64
		var appliedAt string
		err := rows.Scan(&version, &appliedAt)
		if err != nil {
			fmt.Printf("Scan error at row %d: %v\n", count+1, err)
		} else {
			fmt.Printf(" - %d: %s\n", version, appliedAt)
		}
		count++
	}
	if rows.Err() != nil {
		fmt.Printf("Rows error: %v\n", rows.Err())
	}
	rows.Close()
	fmt.Printf("Total migrations scanned: %d\n", count)
}
