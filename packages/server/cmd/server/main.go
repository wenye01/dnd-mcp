// Package main is the entry point for the DND MCP Server
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dnd-mcp/server/internal/api/tools"
	"github.com/dnd-mcp/server/internal/mcp"
	"github.com/dnd-mcp/server/internal/service"
	"github.com/dnd-mcp/server/internal/store/postgres"
	"github.com/dnd-mcp/server/pkg/config"
)

// Build information (set via ldflags)
var (
	version   = "dev"
	gitCommit = "unknown"
	buildTime = "unknown"
)

func main() {
	// Parse command line flags
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("DND MCP Server v%s\n", version)
		fmt.Printf("Git Commit: %s\n", gitCommit)
		fmt.Printf("Build Time: %s\n", buildTime)
		os.Exit(0)
	}

	// Step 1: Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting DND MCP Server v%s...\n", version)
	fmt.Printf("Log level: %s\n", cfg.Log.Level)

	// Step 2: Connect to database
	fmt.Println("Connecting to database...")
	dbClient, err := postgres.NewClient(cfg.Postgres)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := dbClient.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close database connection: %v\n", err)
		}
	}()

	// Verify database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := dbClient.Ping(ctx); err != nil {
		cancel()
		fmt.Fprintf(os.Stderr, "Failed to ping database: %v\n", err)
		os.Exit(1)
	}
	cancel()
	fmt.Println("Database connection established")

	// Step 3: Run database migrations
	fmt.Println("Running database migrations...")
	migrator := postgres.NewMigrator(dbClient)
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	if err := migrator.Up(ctx); err != nil {
		cancel()
		fmt.Fprintf(os.Stderr, "Failed to run migrations: %v\n", err)
		os.Exit(1)
	}
	cancel()

	// Step 4: Create MCP Server
	server := mcp.NewServer(cfg)

	// Step 5: Initialize stores
	characterStore := postgres.NewCharacterStore(dbClient)

	// Step 6: Initialize services
	diceService := service.NewDiceService(characterStore)

	// Step 7: Register Tools
	diceTools := tools.NewDiceTools(diceService)
	diceTools.Register(server.Registry())
	fmt.Println("Dice tools registered: roll_dice, roll_check, roll_save")

	// Step 8: Start HTTP server in goroutine
	go func() {
		fmt.Printf("HTTP server listening on %s:%d\n", cfg.HTTP.Host, cfg.HTTP.Port)
		if err := server.Start(context.Background()); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
			os.Exit(1)
		}
	}()

	fmt.Println("Server initialized successfully")

	// Step 9: Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	// Give outstanding requests 10 seconds to complete
	ctx, cancel = context.WithTimeout(context.Background(), time.Duration(cfg.HTTP.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Server shutdown error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server stopped")
}
