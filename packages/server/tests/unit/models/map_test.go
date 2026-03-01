package models_test

import (
	"testing"

	"github.com/dnd-mcp/server/internal/models"
)

func TestGetTokenSizeInFeet(t *testing.T) {
	tests := []struct {
		name     string
		size     models.TokenSize
		expected int
	}{
		{"tiny token", models.TokenSizeTiny, 2},
		{"small token", models.TokenSizeSmall, 5},
		{"medium token", models.TokenSizeMedium, 5},
		{"large token", models.TokenSizeLarge, 10},
		{"huge token", models.TokenSizeHuge, 15},
		{"gargantuan token", models.TokenSizeGargantuan, 20},
		{"unknown token defaults to medium", models.TokenSize("unknown"), 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.GetTokenSizeInFeet(tt.size)
			if result != tt.expected {
				t.Errorf("GetTokenSizeInFeet() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestGetTokenSizeInGrids(t *testing.T) {
	tests := []struct {
		name     string
		size     models.TokenSize
		expected int
	}{
		{"tiny token", models.TokenSizeTiny, 0}, // 2英尺/5 = 0 (向下取整)
		{"small token", models.TokenSizeSmall, 1},
		{"medium token", models.TokenSizeMedium, 1},
		{"large token", models.TokenSizeLarge, 2},
		{"huge token", models.TokenSizeHuge, 3},
		{"gargantuan token", models.TokenSizeGargantuan, 4},
		{"unknown token defaults to medium", models.TokenSize("unknown"), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.GetTokenSizeInGrids(tt.size)
			if result != tt.expected {
				t.Errorf("GetTokenSizeInGrids() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestNewGrid(t *testing.T) {
	width, height, cellSize := 10, 15, 5
	grid := models.NewGrid(width, height, cellSize)

	if grid.Width != width {
		t.Errorf("expected Width %d, got %d", width, grid.Width)
	}
	if grid.Height != height {
		t.Errorf("expected Height %d, got %d", height, grid.Height)
	}
	if grid.CellSize != cellSize {
		t.Errorf("expected CellSize %d, got %d", cellSize, grid.CellSize)
	}
	if len(grid.Cells) != height {
		t.Errorf("expected %d rows, got %d", height, len(grid.Cells))
	}
	for i, row := range grid.Cells {
		if len(row) != width {
			t.Errorf("expected row %d to have %d columns, got %d", i, width, len(row))
		}
		for j, cell := range row {
			if cell != models.CellTypeEmpty {
				t.Errorf("expected cell (%d,%d) to be empty, got %s", i, j, cell)
			}
		}
	}
}

func TestGrid_Validate(t *testing.T) {
	tests := []struct {
		name     string
		grid     *models.Grid
		wantErr  bool
		errField string
	}{
		{
			name:     "valid grid",
			grid:     models.NewGrid(10, 10, 5),
			wantErr:  false,
		},
		{
			name: "width too low",
			grid: &models.Grid{
				Width:  0,
				Height: 10,
				Cells:  make([][]models.CellType, 10),
			},
			wantErr:  true,
			errField: "grid.width",
		},
		{
			name: "width too high",
			grid: &models.Grid{
				Width:  201,
				Height: 10,
				Cells:  make([][]models.CellType, 10),
			},
			wantErr:  true,
			errField: "grid.width",
		},
		{
			name: "height too low",
			grid: &models.Grid{
				Width:  10,
				Height: -1,
				Cells:  make([][]models.CellType, 0),
			},
			wantErr:  true,
			errField: "grid.height",
		},
		{
			name: "cell size zero",
			grid: &models.Grid{
				Width:    10,
				Height:   10,
				CellSize: 0,
				Cells:    make([][]models.CellType, 10),
			},
			wantErr:  true,
			errField: "grid.cell_size",
		},
		{
			name: "cells height mismatch",
			grid: &models.Grid{
				Width:    10,
				Height:   10,
				CellSize: 5,
				Cells:    make([][]models.CellType, 5),
			},
			wantErr:  true,
			errField: "grid.cells",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize cells with proper width if needed
			if tt.grid.Cells != nil && len(tt.grid.Cells) > 0 && len(tt.grid.Cells[0]) == 0 {
				for i := range tt.grid.Cells {
					tt.grid.Cells[i] = make([]models.CellType, tt.grid.Width)
				}
			}

			err := tt.grid.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if verr, ok := err.(*models.ValidationError); ok {
					if tt.errField != "" && verr.Field != tt.errField {
						t.Errorf("expected error field '%s', got '%s'", tt.errField, verr.Field)
					}
				} else {
					t.Errorf("expected ValidationError, got %T", err)
				}
			}
		})
	}
}

func TestGrid_GetCell(t *testing.T) {
	grid := models.NewGrid(10, 10, 5)

	// Test getting cell
	cell := grid.GetCell(5, 5)
	if cell != models.CellTypeEmpty {
		t.Errorf("expected empty cell, got %s", cell)
	}

	// Test out of bounds
	cell = grid.GetCell(-1, 0)
	if cell != models.CellTypeEmpty {
		t.Errorf("expected empty cell for out of bounds, got %s", cell)
	}

	cell = grid.GetCell(10, 0)
	if cell != models.CellTypeEmpty {
		t.Errorf("expected empty cell for out of bounds, got %s", cell)
	}
}

func TestGrid_SetCell(t *testing.T) {
	grid := models.NewGrid(10, 10, 5)

	// Test setting cell
	success := grid.SetCell(5, 5, models.CellTypeWall)
	if !success {
		t.Error("expected SetCell to succeed")
	}

	cell := grid.GetCell(5, 5)
	if cell != models.CellTypeWall {
		t.Errorf("expected wall cell, got %s", cell)
	}

	// Test out of bounds
	success = grid.SetCell(-1, 0, models.CellTypeWall)
	if success {
		t.Error("expected SetCell to fail for out of bounds")
	}

	success = grid.SetCell(10, 0, models.CellTypeWall)
	if success {
		t.Error("expected SetCell to fail for out of bounds")
	}
}

func TestGrid_IsWalkable(t *testing.T) {
	grid := models.NewGrid(10, 10, 5)

	// Empty cell is walkable
	if !grid.IsWalkable(0, 0) {
		t.Error("expected empty cell to be walkable")
	}

	// Set wall
	grid.SetCell(5, 5, models.CellTypeWall)
	if grid.IsWalkable(5, 5) {
		t.Error("expected wall to not be walkable")
	}

	// Difficult terrain is still walkable
	grid.SetCell(3, 3, models.CellTypeDifficult)
	if !grid.IsWalkable(3, 3) {
		t.Error("expected difficult terrain to be walkable")
	}
}

func TestGrid_IsDifficultTerrain(t *testing.T) {
	grid := models.NewGrid(10, 10, 5)

	// Empty cell is not difficult terrain
	if grid.IsDifficultTerrain(0, 0) {
		t.Error("expected empty cell to not be difficult terrain")
	}

	// Difficult terrain
	grid.SetCell(1, 1, models.CellTypeDifficult)
	if !grid.IsDifficultTerrain(1, 1) {
		t.Error("expected difficult terrain")
	}

	// Water is difficult terrain
	grid.SetCell(2, 2, models.CellTypeWater)
	if !grid.IsDifficultTerrain(2, 2) {
		t.Error("expected water to be difficult terrain")
	}

	// Forest is difficult terrain
	grid.SetCell(3, 3, models.CellTypeForest)
	if !grid.IsDifficultTerrain(3, 3) {
		t.Error("expected forest to be difficult terrain")
	}

	// Mountain is difficult terrain
	grid.SetCell(4, 4, models.CellTypeMountain)
	if !grid.IsDifficultTerrain(4, 4) {
		t.Error("expected mountain to be difficult terrain")
	}

	// Wall is not difficult terrain (it's not walkable at all)
	grid.SetCell(5, 5, models.CellTypeWall)
	if grid.IsDifficultTerrain(5, 5) {
		t.Error("expected wall to not be difficult terrain")
	}
}

func TestNewLocation(t *testing.T) {
	loc := models.NewLocation("Test Location", "A test location", 5, 10)

	if loc.Name != "Test Location" {
		t.Errorf("expected Name 'Test Location', got %s", loc.Name)
	}
	if loc.Description != "A test location" {
		t.Errorf("expected Description 'A test location', got %s", loc.Description)
	}
	if loc.Position.X != 5 {
		t.Errorf("expected X 5, got %d", loc.Position.X)
	}
	if loc.Position.Y != 10 {
		t.Errorf("expected Y 10, got %d", loc.Position.Y)
	}
	if loc.ID == "" {
		t.Error("expected ID to be generated")
	}
}

func TestLocation_Validate(t *testing.T) {
	tests := []struct {
		name     string
		location *models.Location
		wantErr  bool
		errField string
	}{
		{
			name:     "valid location",
			location: models.NewLocation("Test", "Description", 5, 10),
			wantErr:  false,
		},
		{
			name: "empty name",
			location: &models.Location{
				Name:     "",
				Position: models.Position{X: 0, Y: 0},
			},
			wantErr:  true,
			errField: "location.name",
		},
		{
			name: "invalid position",
			location: &models.Location{
				Name:     "Test",
				Position: models.Position{X: -1, Y: 0},
			},
			wantErr:  true,
			errField: "x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.location.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if verr, ok := err.(*models.ValidationError); ok {
					if tt.errField != "" && verr.Field != tt.errField {
						t.Errorf("expected error field '%s', got '%s'", tt.errField, verr.Field)
					}
				} else {
					t.Errorf("expected ValidationError, got %T", err)
				}
			}
		})
	}
}

func TestNewToken(t *testing.T) {
	token := models.NewToken("char-001", 5, 10, models.TokenSizeMedium)

	if token.CharacterID != "char-001" {
		t.Errorf("expected CharacterID 'char-001', got %s", token.CharacterID)
	}
	if token.Position.X != 5 {
		t.Errorf("expected X 5, got %d", token.Position.X)
	}
	if token.Position.Y != 10 {
		t.Errorf("expected Y 10, got %d", token.Position.Y)
	}
	if token.Size != models.TokenSizeMedium {
		t.Errorf("expected Size medium, got %s", token.Size)
	}
	if !token.Visible {
		t.Error("expected Visible to be true by default")
	}
	if token.ID == "" {
		t.Error("expected ID to be generated")
	}
}

func TestToken_Validate(t *testing.T) {
	tests := []struct {
		name     string
		token    *models.Token
		wantErr  bool
		errField string
	}{
		{
			name:     "valid token",
			token:    models.NewToken("char-001", 5, 10, models.TokenSizeMedium),
			wantErr:  false,
		},
		{
			name: "empty character id",
			token: &models.Token{
				CharacterID: "",
				Position:    models.Position{X: 0, Y: 0},
				Size:        models.TokenSizeMedium,
			},
			wantErr:  true,
			errField: "token.character_id",
		},
		{
			name: "invalid position",
			token: &models.Token{
				CharacterID: "char-001",
				Position:    models.Position{X: -1, Y: 0},
				Size:        models.TokenSizeMedium,
			},
			wantErr:  true,
			errField: "x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.token.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if verr, ok := err.(*models.ValidationError); ok {
					if tt.errField != "" && verr.Field != tt.errField {
						t.Errorf("expected error field '%s', got '%s'", tt.errField, verr.Field)
					}
				} else {
					t.Errorf("expected ValidationError, got %T", err)
				}
			}
		})
	}
}

func TestToken_GetSizeInFeet(t *testing.T) {
	tests := []struct {
		size     models.TokenSize
		expected int
	}{
		{models.TokenSizeTiny, 2},
		{models.TokenSizeSmall, 5},
		{models.TokenSizeMedium, 5},
		{models.TokenSizeLarge, 10},
		{models.TokenSizeHuge, 15},
		{models.TokenSizeGargantuan, 20},
	}

	for _, tt := range tests {
		t.Run(string(tt.size), func(t *testing.T) {
			token := models.NewToken("char-001", 0, 0, tt.size)
			result := token.GetSizeInFeet()
			if result != tt.expected {
				t.Errorf("GetSizeInFeet() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestToken_GetSizeInGrids(t *testing.T) {
	tests := []struct {
		size     models.TokenSize
		expected int
	}{
		{models.TokenSizeTiny, 0},
		{models.TokenSizeSmall, 1},
		{models.TokenSizeMedium, 1},
		{models.TokenSizeLarge, 2},
		{models.TokenSizeHuge, 3},
		{models.TokenSizeGargantuan, 4},
	}

	for _, tt := range tests {
		t.Run(string(tt.size), func(t *testing.T) {
			token := models.NewToken("char-001", 0, 0, tt.size)
			result := token.GetSizeInGrids()
			if result != tt.expected {
				t.Errorf("GetSizeInGrids() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestToken_SetPosition(t *testing.T) {
	token := models.NewToken("char-001", 5, 10, models.TokenSizeMedium)

	token.SetPosition(15, 20)

	if token.Position.X != 15 {
		t.Errorf("expected X 15, got %d", token.Position.X)
	}
	if token.Position.Y != 20 {
		t.Errorf("expected Y 20, got %d", token.Position.Y)
	}
}

func TestNewMap(t *testing.T) {
	campaignID := "campaign-001"
	name := "Test Map"
	mapType := models.MapTypeWorld

	gameMap := models.NewMap(campaignID, name, mapType, 20, 30, 1)

	if gameMap.Name != name {
		t.Errorf("expected Name '%s', got %s", name, gameMap.Name)
	}
	if gameMap.CampaignID != campaignID {
		t.Errorf("expected CampaignID '%s', got %s", campaignID, gameMap.CampaignID)
	}
	if gameMap.Type != mapType {
		t.Errorf("expected Type '%s', got %s", mapType, gameMap.Type)
	}
	if gameMap.Grid == nil {
		t.Error("expected Grid to be initialized")
	}
	if gameMap.Grid.Width != 20 {
		t.Errorf("expected Grid Width 20, got %d", gameMap.Grid.Width)
	}
	if gameMap.Grid.Height != 30 {
		t.Errorf("expected Grid Height 30, got %d", gameMap.Grid.Height)
	}
	if gameMap.ID == "" {
		t.Error("expected ID to be generated")
	}
	if len(gameMap.Locations) != 0 {
		t.Error("expected Locations to be empty")
	}
	if len(gameMap.Tokens) != 0 {
		t.Error("expected Tokens to be empty")
	}
}

func TestNewWorldMap(t *testing.T) {
	worldMap := models.NewWorldMap("campaign-001", "World", 50, 50)

	if worldMap.Type != models.MapTypeWorld {
		t.Errorf("expected Type '%s', got %s", models.MapTypeWorld, worldMap.Type)
	}
	if worldMap.Grid.CellSize != 1 {
		t.Errorf("expected CellSize 1 for world map, got %d", worldMap.Grid.CellSize)
	}
}

func TestNewBattleMap(t *testing.T) {
	battleMap := models.NewBattleMap("campaign-001", "Battle", 20, 20, 5)

	if battleMap.Type != models.MapTypeBattle {
		t.Errorf("expected Type '%s', got %s", models.MapTypeBattle, battleMap.Type)
	}
	if battleMap.Grid.CellSize != 5 {
		t.Errorf("expected CellSize 5 for battle map, got %d", battleMap.Grid.CellSize)
	}
}

func TestMap_Validate(t *testing.T) {
	tests := []struct {
		name     string
		gameMap  *models.Map
		wantErr  bool
		errField string
	}{
		{
			name:     "valid world map",
			gameMap:  models.NewWorldMap("campaign-001", "World", 20, 20),
			wantErr:  false,
		},
		{
			name:     "valid battle map",
			gameMap:  models.NewBattleMap("campaign-001", "Battle", 20, 20, 5),
			wantErr:  false,
		},
		{
			name: "empty name",
			gameMap: &models.Map{
				Name:       "",
				CampaignID: "campaign-001",
				Type:       models.MapTypeWorld,
				Grid:       models.NewGrid(10, 10, 1),
				Locations:  make([]models.Location, 0),
				Tokens:     make([]models.Token, 0),
			},
			wantErr:  true,
			errField: "name",
		},
		{
			name: "empty campaign id",
			gameMap: &models.Map{
				Name:       "Test",
				CampaignID: "",
				Type:       models.MapTypeWorld,
				Grid:       models.NewGrid(10, 10, 1),
				Locations:  make([]models.Location, 0),
				Tokens:     make([]models.Token, 0),
			},
			wantErr:  true,
			errField: "campaign_id",
		},
		{
			name: "invalid grid",
			gameMap: &models.Map{
				Name:       "Test",
				CampaignID: "campaign-001",
				Type:       models.MapTypeWorld,
				Grid: &models.Grid{
					Width:  -1,
					Height: 10,
					Cells:  make([][]models.CellType, 10),
				},
				Locations: make([]models.Location, 0),
				Tokens:    make([]models.Token, 0),
			},
			wantErr:  true,
			errField: "grid.width",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize grid cells if needed
			if tt.gameMap.Grid != nil && tt.gameMap.Grid.Cells != nil && len(tt.gameMap.Grid.Cells) > 0 {
				for i := range tt.gameMap.Grid.Cells {
					if tt.gameMap.Grid.Cells[i] == nil && tt.gameMap.Grid.Width > 0 {
						tt.gameMap.Grid.Cells[i] = make([]models.CellType, tt.gameMap.Grid.Width)
					}
				}
			}

			err := tt.gameMap.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if verr, ok := err.(*models.ValidationError); ok {
					if tt.errField != "" && verr.Field != tt.errField {
						t.Errorf("expected error field '%s', got '%s'", tt.errField, verr.Field)
					}
				} else {
					t.Errorf("expected ValidationError, got %T", err)
				}
			}
		})
	}
}

func TestMap_IsWorldMap(t *testing.T) {
	worldMap := models.NewWorldMap("campaign-001", "World", 20, 20)
	battleMap := models.NewBattleMap("campaign-001", "Battle", 20, 20, 5)

	if !worldMap.IsWorldMap() {
		t.Error("expected world map to be identified as world map")
	}
	if worldMap.IsBattleMap() {
		t.Error("expected world map to not be identified as battle map")
	}

	if !battleMap.IsBattleMap() {
		t.Error("expected battle map to be identified as battle map")
	}
	if battleMap.IsWorldMap() {
		t.Error("expected battle map to not be identified as world map")
	}
}

func TestMap_AddLocation(t *testing.T) {
	gameMap := models.NewWorldMap("campaign-001", "World", 20, 20)
	loc := models.NewLocation("Town", "A town", 5, 10)

	err := gameMap.AddLocation(*loc)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(gameMap.Locations) != 1 {
		t.Errorf("expected 1 location, got %d", len(gameMap.Locations))
	}

	// Add invalid location
	invalidLoc := models.Location{
		Name:     "", // invalid
		Position: models.Position{X: 0, Y: 0},
	}
	err = gameMap.AddLocation(invalidLoc)
	if err == nil {
		t.Error("expected error for invalid location")
	}
}

func TestMap_RemoveLocation(t *testing.T) {
	gameMap := models.NewWorldMap("campaign-001", "World", 20, 20)
	loc := models.NewLocation("Town", "A town", 5, 10)

	err := gameMap.AddLocation(*loc)
	if err != nil {
		t.Fatalf("failed to add location: %v", err)
	}

	// Remove existing location
	success := gameMap.RemoveLocation(loc.ID)
	if !success {
		t.Error("expected RemoveLocation to succeed")
	}

	if len(gameMap.Locations) != 0 {
		t.Errorf("expected 0 locations, got %d", len(gameMap.Locations))
	}

	// Remove non-existing location
	success = gameMap.RemoveLocation("non-existing")
	if success {
		t.Error("expected RemoveLocation to fail for non-existing location")
	}
}

func TestMap_GetLocation(t *testing.T) {
	gameMap := models.NewWorldMap("campaign-001", "World", 20, 20)
	loc := models.NewLocation("Town", "A town", 5, 10)

	err := gameMap.AddLocation(*loc)
	if err != nil {
		t.Fatalf("failed to add location: %v", err)
	}

	// Get existing location
	found := gameMap.GetLocation(loc.ID)
	if found == nil {
		t.Error("expected to find location")
	}
	if found.Name != "Town" {
		t.Errorf("expected location name 'Town', got %s", found.Name)
	}

	// Get non-existing location
	found = gameMap.GetLocation("non-existing")
	if found != nil {
		t.Error("expected not to find non-existing location")
	}
}

func TestMap_GetLocationAtPosition(t *testing.T) {
	gameMap := models.NewWorldMap("campaign-001", "World", 20, 20)
	loc := models.NewLocation("Town", "A town", 5, 10)

	err := gameMap.AddLocation(*loc)
	if err != nil {
		t.Fatalf("failed to add location: %v", err)
	}

	// Get location at position
	found := gameMap.GetLocationAtPosition(5, 10)
	if found == nil {
		t.Error("expected to find location at (5, 10)")
	}
	if found.Name != "Town" {
		t.Errorf("expected location name 'Town', got %s", found.Name)
	}

	// Get location at different position
	found = gameMap.GetLocationAtPosition(0, 0)
	if found != nil {
		t.Error("expected not to find location at (0, 0)")
	}
}

func TestMap_AddToken(t *testing.T) {
	gameMap := models.NewBattleMap("campaign-001", "Battle", 20, 20, 5)
	token := models.NewToken("char-001", 5, 10, models.TokenSizeMedium)

	err := gameMap.AddToken(*token)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(gameMap.Tokens) != 1 {
		t.Errorf("expected 1 token, got %d", len(gameMap.Tokens))
	}

	// Add invalid token
	invalidToken := models.Token{
		CharacterID: "", // invalid
		Position:    models.Position{X: 0, Y: 0},
		Size:        models.TokenSizeMedium,
	}
	err = gameMap.AddToken(invalidToken)
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestMap_RemoveToken(t *testing.T) {
	gameMap := models.NewBattleMap("campaign-001", "Battle", 20, 20, 5)
	token := models.NewToken("char-001", 5, 10, models.TokenSizeMedium)

	err := gameMap.AddToken(*token)
	if err != nil {
		t.Fatalf("failed to add token: %v", err)
	}

	// Remove existing token
	success := gameMap.RemoveToken(token.ID)
	if !success {
		t.Error("expected RemoveToken to succeed")
	}

	if len(gameMap.Tokens) != 0 {
		t.Errorf("expected 0 tokens, got %d", len(gameMap.Tokens))
	}

	// Remove non-existing token
	success = gameMap.RemoveToken("non-existing")
	if success {
		t.Error("expected RemoveToken to fail for non-existing token")
	}
}

func TestMap_GetToken(t *testing.T) {
	gameMap := models.NewBattleMap("campaign-001", "Battle", 20, 20, 5)
	token := models.NewToken("char-001", 5, 10, models.TokenSizeMedium)

	err := gameMap.AddToken(*token)
	if err != nil {
		t.Fatalf("failed to add token: %v", err)
	}

	// Get existing token
	found := gameMap.GetToken(token.ID)
	if found == nil {
		t.Error("expected to find token")
	}
	if found.CharacterID != "char-001" {
		t.Errorf("expected character ID 'char-001', got %s", found.CharacterID)
	}

	// Get non-existing token
	found = gameMap.GetToken("non-existing")
	if found != nil {
		t.Error("expected not to find non-existing token")
	}
}

func TestMap_GetTokenByCharacterID(t *testing.T) {
	gameMap := models.NewBattleMap("campaign-001", "Battle", 20, 20, 5)
	token := models.NewToken("char-001", 5, 10, models.TokenSizeMedium)

	err := gameMap.AddToken(*token)
	if err != nil {
		t.Fatalf("failed to add token: %v", err)
	}

	// Get token by character ID
	found := gameMap.GetTokenByCharacterID("char-001")
	if found == nil {
		t.Error("expected to find token by character ID")
	}
	if found.CharacterID != "char-001" {
		t.Errorf("expected character ID 'char-001', got %s", found.CharacterID)
	}

	// Get non-existing character
	found = gameMap.GetTokenByCharacterID("non-existing")
	if found != nil {
		t.Error("expected not to find non-existing character")
	}
}

func TestMap_GetTokensAtPosition(t *testing.T) {
	gameMap := models.NewBattleMap("campaign-001", "Battle", 20, 20, 5)

	// Add medium token (1x1 grid)
	mediumToken := models.NewToken("char-001", 5, 5, models.TokenSizeMedium)
	gameMap.AddToken(*mediumToken)

	// Add large token (2x2 grids)
	largeToken := models.NewToken("char-002", 10, 10, models.TokenSizeLarge)
	gameMap.AddToken(*largeToken)

	// Test getting token at medium token position
	tokens := gameMap.GetTokensAtPosition(5, 5)
	if len(tokens) != 1 {
		t.Errorf("expected 1 token at (5, 5), got %d", len(tokens))
	}

	// Test getting token at large token position (all 4 cells)
	for x := 10; x <= 11; x++ {
		for y := 10; y <= 11; y++ {
			tokens = gameMap.GetTokensAtPosition(x, y)
			if len(tokens) != 1 {
				t.Errorf("expected 1 token at (%d, %d), got %d", x, y, len(tokens))
			}
		}
	}

	// Test empty position
	tokens = gameMap.GetTokensAtPosition(0, 0)
	if len(tokens) != 0 {
		t.Errorf("expected 0 tokens at (0, 0), got %d", len(tokens))
	}
}

// Visual Location and Image Mode Tests

func TestMap_AddVisualLocation(t *testing.T) {
	gameMap := models.NewWorldMap("campaign-001", "World", 20, 20)
	gameMap.Mode = models.MapModeImage

	vloc := models.NewVisualLocation("Town", "A town", "town", 0.5, 0.3)

	err := gameMap.AddVisualLocation(*vloc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(gameMap.VisualLocations) != 1 {
		t.Errorf("expected 1 visual location, got %d", len(gameMap.VisualLocations))
	}

	// Add invalid visual location
	invalidVloc := models.VisualLocation{
		Name:      "", // invalid
		PositionX: 0.5,
		PositionY: 0.3,
	}
	err = gameMap.AddVisualLocation(invalidVloc)
	if err == nil {
		t.Error("expected error for invalid visual location")
	}
}

func TestMap_RemoveVisualLocation(t *testing.T) {
	gameMap := models.NewWorldMap("campaign-001", "World", 20, 20)
	gameMap.Mode = models.MapModeImage

	vloc := models.NewVisualLocation("Town", "A town", "town", 0.5, 0.3)
	gameMap.AddVisualLocation(*vloc)

	// Remove existing visual location
	success := gameMap.RemoveVisualLocation(vloc.ID)
	if !success {
		t.Error("expected RemoveVisualLocation to succeed")
	}

	if len(gameMap.VisualLocations) != 0 {
		t.Errorf("expected 0 visual locations, got %d", len(gameMap.VisualLocations))
	}

	// Remove non-existing visual location
	success = gameMap.RemoveVisualLocation("non-existing")
	if success {
		t.Error("expected RemoveVisualLocation to fail for non-existing location")
	}
}

func TestMap_GetVisualLocation(t *testing.T) {
	gameMap := models.NewWorldMap("campaign-001", "World", 20, 20)
	gameMap.Mode = models.MapModeImage

	vloc := models.NewVisualLocation("Town", "A town", "town", 0.5, 0.3)
	gameMap.AddVisualLocation(*vloc)

	// Get existing visual location
	found := gameMap.GetVisualLocation(vloc.ID)
	if found == nil {
		t.Error("expected to find visual location")
	}
	if found.Name != "Town" {
		t.Errorf("expected location name 'Town', got %s", found.Name)
	}

	// Get non-existing visual location
	found = gameMap.GetVisualLocation("non-existing")
	if found != nil {
		t.Error("expected not to find non-existing visual location")
	}
}

func TestMap_GetVisualLocationAtPosition(t *testing.T) {
	gameMap := models.NewWorldMap("campaign-001", "World", 20, 20)
	gameMap.Mode = models.MapModeImage

	vloc := models.NewVisualLocation("Town", "A town", "town", 0.5, 0.3)
	gameMap.AddVisualLocation(*vloc)

	// Get location at exact position with small tolerance
	found := gameMap.GetVisualLocationAtPosition(0.5, 0.3, 0.01)
	if found == nil {
		t.Error("expected to find visual location at (0.5, 0.3)")
	}
	if found.Name != "Town" {
		t.Errorf("expected location name 'Town', got %s", found.Name)
	}

	// Get location at nearby position within tolerance
	found = gameMap.GetVisualLocationAtPosition(0.51, 0.31, 0.02)
	if found == nil {
		t.Error("expected to find visual location at nearby position")
	}

	// Get location outside tolerance
	found = gameMap.GetVisualLocationAtPosition(0.6, 0.6, 0.01)
	if found != nil {
		t.Error("expected not to find visual location far away")
	}
}

func TestMap_GetConfirmedVisualLocations(t *testing.T) {
	gameMap := models.NewWorldMap("campaign-001", "World", 20, 20)
	gameMap.Mode = models.MapModeImage

	vloc1 := models.NewVisualLocation("Town", "A town", "town", 0.5, 0.3)
	vloc1.Confirm()
	gameMap.AddVisualLocation(*vloc1)

	vloc2 := models.NewVisualLocation("Forest", "A forest", "forest", 0.6, 0.7)
	gameMap.AddVisualLocation(*vloc2)

	confirmed := gameMap.GetConfirmedVisualLocations()
	if len(confirmed) != 1 {
		t.Errorf("expected 1 confirmed location, got %d", len(confirmed))
	}
	if confirmed[0].Name != "Town" {
		t.Errorf("expected 'Town', got %s", confirmed[0].Name)
	}
}

func TestMap_GetUnconfirmedVisualLocations(t *testing.T) {
	gameMap := models.NewWorldMap("campaign-001", "World", 20, 20)
	gameMap.Mode = models.MapModeImage

	vloc1 := models.NewVisualLocation("Town", "A town", "town", 0.5, 0.3)
	vloc1.Confirm()
	gameMap.AddVisualLocation(*vloc1)

	vloc2 := models.NewVisualLocation("Forest", "A forest", "forest", 0.6, 0.7)
	gameMap.AddVisualLocation(*vloc2)

	unconfirmed := gameMap.GetUnconfirmedVisualLocations()
	if len(unconfirmed) != 1 {
		t.Errorf("expected 1 unconfirmed location, got %d", len(unconfirmed))
	}
	if unconfirmed[0].Name != "Forest" {
		t.Errorf("expected 'Forest', got %s", unconfirmed[0].Name)
	}
}

func TestMap_ValidateWithVisualLocations(t *testing.T) {
	gameMap := models.NewWorldMap("campaign-001", "World", 20, 20)
	gameMap.Mode = models.MapModeImage

	// Add valid visual location
	vloc1 := models.NewVisualLocation("Town", "A town", "town", 0.5, 0.3)
	gameMap.AddVisualLocation(*vloc1)

	err := gameMap.Validate()
	if err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}

	// Note: Map.Validate() doesn't validate individual VisualLocations
	// This is by design - visual locations are validated when added via AddVisualLocation
	// The test below documents the current behavior
	vloc2 := models.VisualLocation{
		Name:      "", // invalid - empty name
		PositionX: 0.6,
		PositionY: 0.7,
	}
	gameMap.VisualLocations = append(gameMap.VisualLocations, vloc2)

	// Map validation doesn't check VisualLocation validity
	err = gameMap.Validate()
	if err != nil {
		t.Errorf("Map validation shouldn't fail for invalid visual location: %v", err)
	}

	// But the individual VisualLocation validation should fail
	err = vloc2.Validate()
	if err == nil {
		t.Error("expected VisualLocation validation error for invalid location")
	}
}

func TestMapMode_String(t *testing.T) {
	tests := []struct {
		mode     models.MapMode
		expected string
	}{
		{models.MapModeGrid, "grid"},
		{models.MapModeImage, "image"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.mode) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(tt.mode))
			}
		})
	}
}

func TestMapWithImage(t *testing.T) {
	gameMap := models.NewWorldMap("campaign-001", "World", 20, 20)
	gameMap.Mode = models.MapModeImage
	gameMap.Image = models.NewMapImage("https://example.com/map.jpg")
	gameMap.Image.Width = 1920
	gameMap.Image.Height = 1080

	if gameMap.Image == nil {
		t.Error("expected Image to be set")
	}
	if gameMap.Image.URL != "https://example.com/map.jpg" {
		t.Errorf("expected URL 'https://example.com/map.jpg', got %s", gameMap.Image.URL)
	}
	if gameMap.Image.Width != 1920 {
		t.Errorf("expected Width 1920, got %d", gameMap.Image.Width)
	}
	if gameMap.Image.Height != 1080 {
		t.Errorf("expected Height 1080, got %d", gameMap.Image.Height)
	}

	// Validate the map
	err := gameMap.Validate()
	if err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}
