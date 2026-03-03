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
	"github.com/dnd-mcp/server/internal/importer"
	"github.com/dnd-mcp/server/internal/importer/converter"
	"github.com/dnd-mcp/server/internal/importer/format"
	importer_parser "github.com/dnd-mcp/server/internal/importer/parser"
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
	campaignStore := postgres.NewCampaignStore(dbClient)
	gameStateStore := postgres.NewGameStateStore(dbClient)
	characterStore := postgres.NewCharacterStore(dbClient)
	combatStore := postgres.NewCombatStore(dbClient)
	mapStore := postgres.NewMapStore(dbClient)
	messageStore := postgres.NewMessageStore(dbClient) // M7: Context Management

	// Step 6: Initialize services
	campaignService := service.NewCampaignService(campaignStore, gameStateStore)
	characterService := service.NewCharacterService(characterStore)
	diceService := service.NewDiceService(characterStore)
	combatService := service.NewCombatService(combatStore, characterStore, campaignStore, gameStateStore, diceService)
	mapService := service.NewMapServiceWithCharacters(mapStore, campaignStore, gameStateStore, characterStore)
	contextService := service.NewContextService(messageStore, characterStore, gameStateStore, combatStore, mapStore) // M7: Context Management
	restService := service.NewRestService(characterStore, gameStateStore)                                           // M7.5: Rest System
	conditionService := service.NewConditionService(characterStore)                                                 // M7.5: Condition System

	// Step 6.5: Initialize import service
	importService := importer.NewImportService(mapStore)
	importService.RegisterParser(importer_parser.NewUVTTParser())
	importService.RegisterParser(importer_parser.NewFVTTSceneParser())
	mapConverter := converter.NewMapConverter()
	importService.RegisterConverterForFormat(mapConverter, format.FormatUVTT)
	importService.RegisterConverterForFormat(mapConverter, format.FormatFVTTScene)
	importService.RegisterConverterForFormat(mapConverter, format.FormatFVTTModule)

	// Step 7: Register Tools
	campaignTools := tools.NewCampaignTools(campaignService)
	campaignTools.Register(server.Registry())
	fmt.Println("Campaign tools registered: create_campaign, get_campaign, list_campaigns, delete_campaign, get_campaign_summary")

	characterTools := tools.NewCharacterTools(characterService)
	characterTools.Register(server.Registry())
	fmt.Println("Character tools registered: create_character, get_character, update_character, list_characters, delete_character")

	diceTools := tools.NewDiceTools(diceService)
	diceTools.Register(server.Registry())
	fmt.Println("Dice tools registered: roll_dice, roll_check, roll_save")

	combatTools := tools.NewCombatTools(combatService)
	combatTools.Register(server.Registry())
	fmt.Println("Combat tools registered: start_combat, get_combat_state, attack, cast_spell, end_turn, end_combat")

	mapTools := tools.NewMapToolsWithCharacters(mapService)
	mapTools.Register(server.Registry())
	fmt.Println("Map tools registered: get_world_map, move_to, move_token, enter_battle_map, get_battle_map, exit_battle_map, create_visual_location, update_location")

	// Step 7.5: Register Import Tools
	importTools := tools.NewImportTools(importService)
	importTools.Register(server.Registry())
	fmt.Println("Import tools registered: import_map, import_map_from_module")

	// Step 7.6: Register Context Tools (M7)
	contextTools := tools.NewContextTools(contextService)
	contextTools.Register(server.Registry())
	fmt.Println("Context tools registered: get_context, get_raw_context, save_message")

	// Step 7.7: Register Rest Tools (M7.5)
	restTools := tools.NewRestTools(restService)
	restTools.Register(server.Registry())
	fmt.Println("Rest tools registered: take_short_rest, take_long_rest, party_long_rest")

	// Step 7.8: Register Condition Tools (M7.5)
	conditionTools := tools.NewConditionTools(conditionService)
	conditionTools.Register(server.Registry())
	fmt.Println("Condition tools registered: apply_condition, remove_condition, get_conditions, has_condition")

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
