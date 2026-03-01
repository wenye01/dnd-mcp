// Package service provides business logic layer implementations
package service

import (
	"context"
	"fmt"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/store"
)

// MapStore defines the interface for map data operations
type MapStore interface {
	Create(ctx context.Context, gameMap *models.Map) error
	Get(ctx context.Context, id string) (*models.Map, error)
	GetByCampaign(ctx context.Context, campaignID string) ([]*models.Map, error)
	Update(ctx context.Context, gameMap *models.Map) error
	Delete(ctx context.Context, id string) error
	GetWorldMap(ctx context.Context, campaignID string) (*models.Map, error)
	GetBattleMap(ctx context.Context, id string) (*models.Map, error)
	GetByParent(ctx context.Context, parentID string) ([]*models.Map, error)
}

// CampaignStoreForMap defines the campaign store interface needed by map service
type CampaignStoreForMap interface {
	Get(ctx context.Context, id string) (*models.Campaign, error)
}

// GameStateStoreForMap defines the game state store interface needed by map service
type GameStateStoreForMap interface {
	Get(ctx context.Context, campaignID string) (*models.GameState, error)
	Update(ctx context.Context, gameState *models.GameState) error
}

// MapService provides map business logic
// 规则参考: PHB 第8章 Travel, 第9章 Combat
type MapService struct {
	mapStore       MapStore
	campaignStore  CampaignStoreForMap
	gameStateStore GameStateStoreForMap
}

// NewMapService creates a new map service
func NewMapService(
	mapStore MapStore,
	campaignStore CampaignStoreForMap,
	gameStateStore GameStateStoreForMap,
) *MapService {
	return &MapService{
		mapStore:       mapStore,
		campaignStore:  campaignStore,
		gameStateStore: gameStateStore,
	}
}

// GetWorldMap retrieves the world map for a campaign
func (s *MapService) GetWorldMap(ctx context.Context, campaignID string) (*models.Map, error) {
	if campaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}

	// Verify campaign exists
	_, err := s.campaignStore.Get(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	// Get or create world map
	worldMap, err := s.mapStore.GetWorldMap(ctx, campaignID)
	if err != nil {
		// Create default world map if not exists
		worldMap = models.NewWorldMap(campaignID, "World Map", 50, 50)
		if err := s.mapStore.Create(ctx, worldMap); err != nil {
			return nil, fmt.Errorf("failed to create world map: %w", err)
		}
	}

	return worldMap, nil
}

// GetWorldMapWithPlayerMarker retrieves the world map with player marker for a campaign
// This method returns both the map and the player marker from game state for Image mode maps
func (s *MapService) GetWorldMapWithPlayerMarker(ctx context.Context, campaignID string) (*models.Map, *models.PlayerMarker, error) {
	if campaignID == "" {
		return nil, nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}

	// Get world map
	worldMap, err := s.GetWorldMap(ctx, campaignID)
	if err != nil {
		return nil, nil, err
	}

	// Get game state for player marker
	gameState, err := s.gameStateStore.Get(ctx, campaignID)
	if err != nil {
		// If game state doesn't exist, return map with nil marker
		return worldMap, nil, nil
	}

	return worldMap, gameState.PlayerMarker, nil
}

// GetBattleMap retrieves a battle map by ID
func (s *MapService) GetBattleMap(ctx context.Context, mapID string) (*models.Map, error) {
	if mapID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "map ID is required")
	}

	battleMap, err := s.mapStore.GetBattleMap(ctx, mapID)
	if err != nil {
		return nil, fmt.Errorf("failed to get battle map: %w", err)
	}

	return battleMap, nil
}

// CreateBattleMapRequest represents a battle map creation request
type CreateBattleMapRequest struct {
	CampaignID string `json:"campaign_id"`
	Name       string `json:"name"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	CellSize   int    `json:"cell_size"`
	ParentID   string `json:"parent_id"` // Optional parent location ID
}

// CreateBattleMap creates a new battle map
func (s *MapService) CreateBattleMap(ctx context.Context, req *CreateBattleMapRequest) (*models.Map, error) {
	if req.CampaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}
	if req.Name == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "name is required")
	}
	if req.Width <= 0 || req.Width > models.MaxMapWidth {
		return nil, NewServiceError(ErrCodeInvalidInput, fmt.Sprintf("width must be between 1 and %d", models.MaxMapWidth))
	}
	if req.Height <= 0 || req.Height > models.MaxMapHeight {
		return nil, NewServiceError(ErrCodeInvalidInput, fmt.Sprintf("height must be between 1 and %d", models.MaxMapHeight))
	}
	if req.CellSize <= 0 {
		return nil, NewServiceError(ErrCodeInvalidInput, "cell size must be positive")
	}

	// Verify campaign exists
	_, err := s.campaignStore.Get(ctx, req.CampaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	// Create battle map
	// 规则参考: PHB 第9章 - Combat grid is typically 5 feet per square
	cellSize := req.CellSize
	if cellSize == 0 {
		cellSize = 5 // Default 5 feet
	}

	battleMap := models.NewBattleMap(req.CampaignID, req.Name, req.Width, req.Height, cellSize)
	battleMap.ParentID = req.ParentID

	if err := battleMap.Validate(); err != nil {
		return nil, NewServiceError(ErrCodeInvalidInput, fmt.Sprintf("invalid battle map: %v", err))
	}

	if err := s.mapStore.Create(ctx, battleMap); err != nil {
		return nil, fmt.Errorf("failed to create battle map: %w", err)
	}

	return battleMap, nil
}

// MoveToRequest represents a move request
type MoveToRequest struct {
	CampaignID string  `json:"campaign_id"`
	X          int     `json:"x"`                // Grid 模式坐标
	Y          int     `json:"y"`                // Grid 模式坐标
	Pace       string  `json:"pace"`             // Travel pace: fast, normal, slow
	TargetX    float64 `json:"target_x,omitempty"`  // Image 模式归一化坐标 (0-1)
	TargetY    float64 `json:"target_y,omitempty"`  // Image 模式归一化坐标 (0-1)
}

// MoveToResult represents the result of a move operation
type MoveToResult struct {
	GameState    *models.GameState   `json:"game_state"`
	TravelResult *TravelResult       `json:"travel"`
	NewMarker    *models.PlayerMarker `json:"new_marker,omitempty"` // 更新后的玩家标记
}

// MoveTo moves the party to a specific position on the world map
// 规则参考: PHB 第8章 - Travel Pace
func (s *MapService) MoveTo(ctx context.Context, req *MoveToRequest) (*MoveToResult, error) {
	if req.CampaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}

	// Get world map to check mode and validate
	worldMap, err := s.mapStore.GetWorldMap(ctx, req.CampaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get world map: %w", err)
	}

	// Get current game state
	gameState, err := s.gameStateStore.Get(ctx, req.CampaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game state: %w", err)
	}

	// Use provided pace or default to normal
	pace := req.Pace
	if pace == "" {
		pace = "normal"
	}

	// Check map mode and handle accordingly
	if worldMap.Mode == models.MapModeImage {
		// Image 模式: 使用归一化坐标
		if req.TargetX < 0 || req.TargetX > 1 {
			return nil, NewServiceError(ErrCodeInvalidInput, "target_x must be between 0 and 1")
		}
		if req.TargetY < 0 || req.TargetY > 1 {
			return nil, NewServiceError(ErrCodeInvalidInput, "target_y must be between 0 and 1")
		}

		// Create or update PlayerMarker
		newMarker := models.NewPlayerMarker(req.TargetX, req.TargetY)
		gameState.PlayerMarker = newMarker

		// Update game state
		if err := s.gameStateStore.Update(ctx, gameState); err != nil {
			return nil, fmt.Errorf("failed to update game state: %w", err)
		}

		return &MoveToResult{
			GameState: gameState,
			NewMarker: newMarker,
		}, nil
	}

	// Grid 模式: 使用现有逻辑（X, Y 坐标）
	// Validate position
	if req.X < 0 || req.Y < 0 {
		return nil, NewServiceError(ErrCodeInvalidInput, "position cannot be negative")
	}

	// Check if position is within bounds
	if req.X >= worldMap.Grid.Width || req.Y >= worldMap.Grid.Height {
		return nil, NewServiceError(ErrCodeInvalidInput, "position is out of map bounds")
	}

	// Calculate travel time and update game time
	oldX, oldY := 0, 0
	if gameState.PartyPosition != nil {
		oldX = gameState.PartyPosition.X
		oldY = gameState.PartyPosition.Y
	}

	travelResult := s.CalculateTravelTime(oldX, oldY, req.X, req.Y, pace)

	// Update position
	newPos := &models.Position{X: req.X, Y: req.Y}
	if err := gameState.SetPartyPosition(newPos); err != nil {
		return nil, fmt.Errorf("failed to set party position: %w", err)
	}

	// Advance game time by travel duration
	// 规则参考: PHB 第8章 - Travel Pace
	gameState.AdvanceTime(travelResult.Hours)

	// Update game state
	if err := s.gameStateStore.Update(ctx, gameState); err != nil {
		return nil, fmt.Errorf("failed to update game state: %w", err)
	}

	return &MoveToResult{
		GameState:    gameState,
		TravelResult: travelResult,
	}, nil
}

// MoveToLocationRequest represents a move to location request
type MoveToLocationRequest struct {
	CampaignID string `json:"campaign_id"`
	LocationID string `json:"location_id"`
}

// MoveToLocation moves the party to a specific location
func (s *MapService) MoveToLocation(ctx context.Context, req *MoveToLocationRequest) (*models.GameState, *models.Location, error) {
	if req.CampaignID == "" {
		return nil, nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}
	if req.LocationID == "" {
		return nil, nil, NewServiceError(ErrCodeInvalidInput, "location ID is required")
	}

	// Get world map
	worldMap, err := s.mapStore.GetWorldMap(ctx, req.CampaignID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get world map: %w", err)
	}

	// Find location
	location := worldMap.GetLocation(req.LocationID)
	if location == nil {
		return nil, nil, NewServiceError(ErrCodeNotFound, "location not found")
	}

	// Get current game state
	gameState, err := s.gameStateStore.Get(ctx, req.CampaignID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get game state: %w", err)
	}

	// Calculate travel time
	oldX, oldY := 0, 0
	if gameState.PartyPosition != nil {
		oldX = gameState.PartyPosition.X
		oldY = gameState.PartyPosition.Y
	}

	travelResult := s.CalculateTravelTime(oldX, oldY, location.Position.X, location.Position.Y, "normal")

	// Update position
	if err := gameState.SetPartyPosition(&location.Position); err != nil {
		return nil, nil, fmt.Errorf("failed to set party position: %w", err)
	}

	// Advance game time
	gameState.AdvanceTime(travelResult.Hours)

	// Update game state
	if err := s.gameStateStore.Update(ctx, gameState); err != nil {
		return nil, nil, fmt.Errorf("failed to update game state: %w", err)
	}

	return gameState, location, nil
}

// TravelResult represents the result of travel time calculation
type TravelResult struct {
	Distance    int     `json:"distance"`     // Distance in miles
	Hours       int     `json:"hours"`        // Travel time in hours
	Days        float64 `json:"days"`         // Travel time in days
	Pace        string  `json:"pace"`         // Travel pace used
	Description string  `json:"description"`  // Human-readable description
}

// TravelMode represents the travel pace
// 规则参考: PHB 第8章 - Travel Pace
type TravelMode string

const (
	// TravelModeFast Fast pace: 4 miles/hour, 30 miles/day, disadvantage on Perception
	TravelModeFast TravelMode = "fast"
	// TravelModeNormal Normal pace: 3 miles/hour, 24 miles/day
	TravelModeNormal TravelMode = "normal"
	// TravelModeSlow Slow pace: 2 miles/hour, 12 miles/day, can use Stealth
	TravelModeSlow TravelMode = "slow"
)

// CalculateTravelTime calculates travel time between two positions
// 规则参考: PHB 第8章 - Travel Pace
// - Fast: 4 mph / 30 miles per day
// - Normal: 3 mph / 24 miles per day
// - Slow: 2 mph / 12 miles per day
// World map: 1 grid = 1 mile
func (s *MapService) CalculateTravelTime(fromX, fromY, toX, toY int, travelMode string) *TravelResult {
	// Calculate distance using Manhattan distance for overland travel
	distance := abs(toX-fromX) + abs(toY-fromY)

	// Normalize travel mode
	mode := TravelModeNormal
	if travelMode != "" {
		mode = TravelMode(travelMode)
	}

	var mph int
	var milesPerDay int
	var description string

	switch mode {
	case TravelModeFast:
		mph = 4
		milesPerDay = 30
		description = "Fast pace"
	case TravelModeSlow:
		mph = 2
		milesPerDay = 12
		description = "Slow pace"
	default: // Normal
		mph = 3
		milesPerDay = 24
		description = "Normal pace"
	}

	// Calculate hours (at least 1 hour if distance > 0)
	hours := distance / mph
	if distance > 0 && hours == 0 {
		hours = 1
	}

	// Calculate days
	days := float64(distance) / float64(milesPerDay)

	return &TravelResult{
		Distance:    distance,
		Hours:       hours,
		Days:        days,
		Pace:        string(mode),
		Description: fmt.Sprintf("%s: %d miles at %d mph = %d hours", description, distance, mph, hours),
	}
}

// AddLocationRequest represents a location addition request
type AddLocationRequest struct {
	CampaignID  string `json:"campaign_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	BattleMapID string `json:"battle_map_id"` // Optional associated battle map
}

// AddLocation adds a new location to the world map
func (s *MapService) AddLocation(ctx context.Context, req *AddLocationRequest) (*models.Location, error) {
	if req.CampaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}
	if req.Name == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "name is required")
	}

	// Get world map
	worldMap, err := s.mapStore.GetWorldMap(ctx, req.CampaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get world map: %w", err)
	}

	// Validate position is within bounds
	if req.X < 0 || req.X >= worldMap.Grid.Width || req.Y < 0 || req.Y >= worldMap.Grid.Height {
		return nil, NewServiceError(ErrCodeInvalidInput, "position is out of map bounds")
	}

	// Check if location already exists at this position
	if existing := worldMap.GetLocationAtPosition(req.X, req.Y); existing != nil {
		return nil, NewServiceError(ErrCodeInvalidState, "a location already exists at this position")
	}

	// Create new location
	location := models.NewLocation(req.Name, req.Description, req.X, req.Y)
	location.BattleMapID = req.BattleMapID

	// Add to map
	if err := worldMap.AddLocation(*location); err != nil {
		return nil, fmt.Errorf("failed to add location: %w", err)
	}

	// Update map
	if err := s.mapStore.Update(ctx, worldMap); err != nil {
		return nil, fmt.Errorf("failed to update map: %w", err)
	}

	return location, nil
}

// LocationUpdate represents location update fields
type LocationUpdate struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	X           *int    `json:"x,omitempty"`
	Y           *int    `json:"y,omitempty"`
	BattleMapID *string `json:"battle_map_id,omitempty"`
}

// UpdateLocation updates an existing location
func (s *MapService) UpdateLocation(ctx context.Context, campaignID, locationID string, updates LocationUpdate) (*models.Location, error) {
	if campaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}
	if locationID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "location ID is required")
	}

	// Get world map
	worldMap, err := s.mapStore.GetWorldMap(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get world map: %w", err)
	}

	// Find location
	location := worldMap.GetLocation(locationID)
	if location == nil {
		return nil, NewServiceError(ErrCodeNotFound, "location not found")
	}

	// Apply updates
	if updates.Name != nil {
		location.Name = *updates.Name
	}
	if updates.Description != nil {
		location.Description = *updates.Description
	}
	if updates.X != nil {
		// Validate new position
		if *updates.X < 0 || *updates.X >= worldMap.Grid.Width {
			return nil, NewServiceError(ErrCodeInvalidInput, "new X position is out of bounds")
		}
		location.Position.X = *updates.X
	}
	if updates.Y != nil {
		if *updates.Y < 0 || *updates.Y >= worldMap.Grid.Height {
			return nil, NewServiceError(ErrCodeInvalidInput, "new Y position is out of bounds")
		}
		location.Position.Y = *updates.Y
	}
	if updates.BattleMapID != nil {
		location.BattleMapID = *updates.BattleMapID
	}

	// Validate
	if err := location.Validate(); err != nil {
		return nil, NewServiceError(ErrCodeInvalidInput, fmt.Sprintf("invalid location: %v", err))
	}

	// Update map
	if err := s.mapStore.Update(ctx, worldMap); err != nil {
		return nil, fmt.Errorf("failed to update map: %w", err)
	}

	return location, nil
}

// TokenMoveRequest represents a token move request
type TokenMoveRequest struct {
	CampaignID string `json:"campaign_id"`
	MapID      string `json:"map_id"`
	TokenID    string `json:"token_id"`
	ToX        int    `json:"to_x"`
	ToY        int    `json:"to_y"`
	// Speed is the available movement speed in feet (optional, defaults to character speed)
	Speed *int `json:"speed,omitempty"`
}

// TokenMoveResult represents the result of a token move operation
type TokenMoveResult struct {
	Token                  *models.Token  `json:"token"`
	MovementUsed           int            `json:"movement_used"`
	RemainingSpeed         int            `json:"remaining_speed"`
	Path                   []models.Position `json:"path"`
	DifficultTerrainCount  int            `json:"difficult_terrain_count"`
}

// MoveToken moves a token on a battle map
// 规则参考: PHB 第9章 - Movement and Position
func (s *MapService) MoveToken(ctx context.Context, req *TokenMoveRequest) (*TokenMoveResult, error) {
	if req.CampaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}
	if req.MapID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "map ID is required")
	}
	if req.TokenID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "token ID is required")
	}

	// Get the battle map
	battleMap, err := s.mapStore.Get(ctx, req.MapID)
	if err != nil {
		return nil, fmt.Errorf("failed to get map: %w", err)
	}

	if !battleMap.IsBattleMap() {
		return nil, NewServiceError(ErrCodeInvalidInput, "token movement is only supported on battle maps")
	}

	// Verify campaign matches
	if battleMap.CampaignID != req.CampaignID {
		return nil, NewServiceError(ErrCodeInvalidInput, "map does not belong to the specified campaign")
	}

	// Get the token
	token := battleMap.GetToken(req.TokenID)
	if token == nil {
		return nil, NewServiceError(ErrCodeNotFound, "token not found on this map")
	}

	// Validate destination position BEFORE calculating movement cost
	// This allows early exit for invalid positions
	sizeInGrids := token.GetSizeInGrids()
	if req.ToX < 0 || req.ToY < 0 || req.ToX+sizeInGrids > battleMap.Grid.Width || req.ToY+sizeInGrids > battleMap.Grid.Height {
		return nil, NewServiceError(ErrCodeInvalidInput, "destination position is out of map bounds")
	}

	// Check if no actual movement
	fromX, fromY := token.Position.X, token.Position.Y
	if fromX == req.ToX && fromY == req.ToY {
		// No movement needed, return early
		return &TokenMoveResult{
			Token:                 token,
			MovementUsed:          0,
			RemainingSpeed:        30, // Default or use req.Speed
			Path:                  []models.Position{{X: fromX, Y: fromY}},
			DifficultTerrainCount: 0,
		}, nil
	}

	// Calculate movement cost using Manhattan distance with terrain consideration
	movementCost, difficultCount := s.calculateMovementCost(battleMap, token, fromX, fromY, req.ToX, req.ToY)

	// Determine available speed
	availableSpeed := 30 // Default 30 feet
	if req.Speed != nil {
		availableSpeed = *req.Speed
	} else {
		// Try to get speed from character (if we had character store access)
		// For now, use default speed
	}

	// Check if movement is possible
	if movementCost > availableSpeed {
		return nil, NewServiceError(ErrCodeInvalidState, fmt.Sprintf("insufficient movement: need %d feet, have %d feet", movementCost, availableSpeed))
	}

	// Check for blocking walls (simple line check)
	if s.isPathBlocked(battleMap, token, fromX, fromY, req.ToX, req.ToY) {
		return nil, NewServiceError(ErrCodeInvalidState, "movement path is blocked by walls")
	}

	// Check for other tokens blocking destination
	otherTokens := battleMap.GetTokensAtPosition(req.ToX, req.ToY)
	for _, other := range otherTokens {
		if other.ID != token.ID {
			// Check if can move through (size difference)
			if !canTokenMoveThrough(token, &other) {
				return nil, NewServiceError(ErrCodeInvalidState, "destination space is occupied by another creature")
			}
		}
	}

	// Generate simple path (straight line, can be improved with A* later)
	path := s.generatePath(fromX, fromY, req.ToX, req.ToY)

	// Update token position
	token.SetPosition(req.ToX, req.ToY)

	// Update map
	if err := s.mapStore.Update(ctx, battleMap); err != nil {
		return nil, fmt.Errorf("failed to update map: %w", err)
	}

	remainingSpeed := availableSpeed - movementCost

	return &TokenMoveResult{
		Token:                 token,
		MovementUsed:          movementCost,
		RemainingSpeed:        remainingSpeed,
		Path:                  path,
		DifficultTerrainCount: difficultCount,
	}, nil
}

// calculateMovementCost calculates the movement cost for a token
func (s *MapService) calculateMovementCost(battleMap *models.Map, token *models.Token, fromX, fromY, toX, toY int) (int, int) {
	dx := abs(toX - fromX)
	dy := abs(toY - fromY)

	// Manhattan distance for base cost
	baseSquares := dx + dy
	cellSize := battleMap.Grid.CellSize
	if cellSize == 0 {
		cellSize = 5 // Default 5 feet
	}

	movementCost := baseSquares * cellSize
	difficultCount := 0

	// Diagonal movement penalty (every 2nd diagonal = 2 squares, simplified to 1.5x)
	isDiagonal := (dx > 0 && dy > 0)
	if isDiagonal {
		movementCost = (movementCost * 3) / 2
	}

	// Check for difficult terrain along the path
	// For simplicity, we check if any square along the direct path is difficult
	if s.pathHasDifficultTerrain(battleMap, fromX, fromY, toX, toY) {
		// Double the cost for difficult terrain
		movementCost *= 2
		difficultCount = baseSquares
	}

	return movementCost, difficultCount
}

// pathHasDifficultTerrain checks if the path passes through difficult terrain
func (s *MapService) pathHasDifficultTerrain(battleMap *models.Map, fromX, fromY, toX, toY int) bool {
	// Simple check: sample a few points along the path
	dx := toX - fromX
	dy := toY - fromY
	steps := max(abs(dx), abs(dy))

	for i := 0; i <= steps; i++ {
		if steps == 0 {
			break
		}
		t := float64(i) / float64(steps)
		x := fromX + int(float64(dx)*t + 0.5)
		y := fromY + int(float64(dy)*t + 0.5)

		if battleMap.Grid.IsDifficultTerrain(x, y) {
			return true
		}
	}

	return false
}

// isPathBlocked checks if the path is blocked by walls
func (s *MapService) isPathBlocked(battleMap *models.Map, token *models.Token, fromX, fromY, toX, toY int) bool {
	// Simple check: sample points along the path for walls
	dx := toX - fromX
	dy := toY - fromY
	steps := max(abs(dx), abs(dy))

	// Check all points along the path
	for i := 1; i < steps; i++ {
		t := float64(i) / float64(steps)
		x := fromX + int(float64(dx)*t + 0.5)
		y := fromY + int(float64(dy)*t + 0.5)

		if !battleMap.Grid.IsWalkable(x, y) {
			return true
		}
	}

	return false
}

// generatePath generates a simple path between two positions
func (s *MapService) generatePath(fromX, fromY, toX, toY int) []models.Position {
	dx := toX - fromX
	dy := toY - fromY
	steps := max(abs(dx), abs(dy))

	path := make([]models.Position, 0, steps+1)
	path = append(path, models.Position{X: fromX, Y: fromY})

	for i := 1; i < steps; i++ {
		t := float64(i) / float64(steps)
		x := fromX + int(float64(dx)*t + 0.5)
		y := fromY + int(float64(dy)*t + 0.5)
		path = append(path, models.Position{X: x, Y: y})
	}

	path = append(path, models.Position{X: toX, Y: toY})
	return path
}

// canTokenMoveThrough checks if a token can move through another token's space
// 规则参考: PHB 第9章 - Size and Space
// "A creature can move through a space occupied by a creature 2 or more sizes smaller"
func canTokenMoveThrough(movingToken, otherToken *models.Token) bool {
	movingSize := getSizeRank(movingToken.Size)
	otherSize := getSizeRank(otherToken.Size)
	return movingSize >= otherSize+2
}

// getSizeRank returns a numeric rank for token size comparison
func getSizeRank(size models.TokenSize) int {
	switch size {
	case models.TokenSizeTiny:
		return 0
	case models.TokenSizeSmall:
		return 1
	case models.TokenSizeMedium:
		return 2
	case models.TokenSizeLarge:
		return 3
	case models.TokenSizeHuge:
		return 4
	case models.TokenSizeGargantuan:
		return 5
	default:
		return 2 // Default to medium
	}
}

// abs returns the absolute value of an integer
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ============ Map Switching Methods ============

// CharacterStoreForMap defines the character store interface needed for token placement
type CharacterStoreForMap interface {
	GetByCampaignAndID(ctx context.Context, campaignID, characterID string) (*models.Character, error)
	List(ctx context.Context, filter *store.CharacterFilter) ([]*models.Character, error)
}

// MapServiceWithCharacters extends MapService with character store for token operations
type MapServiceWithCharacters struct {
	*MapService
	characterStore CharacterStoreForMap
}

// NewMapServiceWithCharacters creates a new map service with character support
func NewMapServiceWithCharacters(
	mapStore MapStore,
	campaignStore CampaignStoreForMap,
	gameStateStore GameStateStoreForMap,
	characterStore CharacterStoreForMap,
) *MapServiceWithCharacters {
	baseService := NewMapService(mapStore, campaignStore, gameStateStore)
	return &MapServiceWithCharacters{
		MapService:     baseService,
		characterStore: characterStore,
	}
}

// EnterBattleMapRequest represents a request to enter a battle map
type EnterBattleMapRequest struct {
	CampaignID      string `json:"campaign_id"`
	LocationID      string `json:"location_id"`
	BattleMapID     string `json:"battle_map_id,omitempty"`
	CreateIfMissing bool   `json:"create_if_missing"`
}

// EnterBattleMapResult represents the result of entering a battle map
type EnterBattleMapResult struct {
	BattleMap    *models.Map        `json:"battle_map"`
	GameState    *models.GameState  `json:"game_state"`
	TokensPlaced int                `json:"tokens_placed"`
}

// EnterBattleMap enters a battle map from the world map
// 规则参考: PHB 第9章 - Combat starts when characters roll initiative
func (s *MapServiceWithCharacters) EnterBattleMap(ctx context.Context, req *EnterBattleMapRequest) (*EnterBattleMapResult, error) {
	if req.CampaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}
	if req.LocationID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "location ID is required")
	}

	// Verify campaign exists
	_, err := s.campaignStore.Get(ctx, req.CampaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	// Get current game state
	gameState, err := s.gameStateStore.Get(ctx, req.CampaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game state: %w", err)
	}

	// Check if already in battle map
	if gameState.IsInBattleMap() {
		return nil, NewServiceError(ErrCodeInvalidState, "already in a battle map")
	}

	// Get world map
	worldMap, err := s.mapStore.GetWorldMap(ctx, req.CampaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get world map: %w", err)
	}

	// Find location
	location := worldMap.GetLocation(req.LocationID)
	if location == nil {
		return nil, NewServiceError(ErrCodeNotFound, "location not found on world map")
	}

	var battleMap *models.Map

	// Check if location has an associated battle map or use provided one
	battleMapID := req.BattleMapID
	if battleMapID == "" && location.BattleMapID != "" {
		battleMapID = location.BattleMapID
	}

	if battleMapID != "" {
		// Get existing battle map
		battleMap, err = s.mapStore.GetBattleMap(ctx, battleMapID)
		if err != nil {
			return nil, fmt.Errorf("failed to get battle map: %w", err)
		}
		if battleMap.CampaignID != req.CampaignID {
			return nil, NewServiceError(ErrCodeInvalidInput, "battle map does not belong to this campaign")
		}
	} else if req.CreateIfMissing {
		// Create a new battle map
		createReq := &CreateBattleMapRequest{
			CampaignID: req.CampaignID,
			Name:       location.Name + " Battle Map",
			Width:      20, // Default 20x20 grid
			Height:     20,
			CellSize:   5,  // 5 feet per square
			ParentID:   req.LocationID,
		}
		battleMap, err = s.CreateBattleMap(ctx, createReq)
		if err != nil {
			return nil, fmt.Errorf("failed to create battle map: %w", err)
		}
		// Update location with battle map reference
		location.BattleMapID = battleMap.ID
		if err := s.mapStore.Update(ctx, worldMap); err != nil {
			return nil, fmt.Errorf("failed to update location: %w", err)
		}
	} else {
		return nil, NewServiceError(ErrCodeNotFound, "no battle map found for this location")
	}

	// Place party tokens at the edge of the battle map
	tokensPlaced, err := s.placePartyTokens(ctx, req.CampaignID, battleMap)
	if err != nil {
		return nil, fmt.Errorf("failed to place tokens: %w", err)
	}

	// Update game state to current battle map
	gameState.SetCurrentMap(battleMap.ID, models.MapTypeBattle)

	// Save updated battle map and game state
	if err := s.mapStore.Update(ctx, battleMap); err != nil {
		return nil, fmt.Errorf("failed to update battle map: %w", err)
	}
	if err := s.gameStateStore.Update(ctx, gameState); err != nil {
		return nil, fmt.Errorf("failed to update game state: %w", err)
	}

	return &EnterBattleMapResult{
		BattleMap:    battleMap,
		GameState:    gameState,
		TokensPlaced: tokensPlaced,
	}, nil
}

// GetBattleMap retrieves the current battle map for a campaign
func (s *MapService) GetBattleMapByCampaign(ctx context.Context, campaignID string) (*models.Map, error) {
	if campaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}

	// Get game state to find current battle map
	gameState, err := s.gameStateStore.Get(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game state: %w", err)
	}

	// Check if in battle map
	if !gameState.IsInBattleMap() {
		return nil, NewServiceError(ErrCodeInvalidState, "not currently in a battle map")
	}

	// Get the battle map
	battleMap, err := s.mapStore.GetBattleMap(ctx, gameState.CurrentMapID)
	if err != nil {
		return nil, fmt.Errorf("failed to get battle map: %w", err)
	}

	return battleMap, nil
}

// ExitBattleMapRequest represents a request to exit a battle map
type ExitBattleMapRequest struct {
	CampaignID    string `json:"campaign_id"`
	KeepBattleMap bool   `json:"keep_battle_map"`
}

// ExitBattleMapResult represents the result of exiting a battle map
type ExitBattleMapResult struct {
	GameState *models.GameState `json:"game_state"`
	Location  *models.Location  `json:"location"`
}

// ExitBattleMap exits the current battle map and returns to the world map
func (s *MapService) ExitBattleMap(ctx context.Context, req *ExitBattleMapRequest) (*ExitBattleMapResult, error) {
	if req.CampaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}

	// Get current game state
	gameState, err := s.gameStateStore.Get(ctx, req.CampaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game state: %w", err)
	}

	// Check if in battle map
	if !gameState.IsInBattleMap() {
		return nil, NewServiceError(ErrCodeInvalidState, "not currently in a battle map")
	}

	// Get the battle map to find parent location
	battleMap, err := s.mapStore.Get(ctx, gameState.CurrentMapID)
	if err != nil {
		return nil, fmt.Errorf("failed to get battle map: %w", err)
	}

	// Get world map
	worldMap, err := s.mapStore.GetWorldMap(ctx, req.CampaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get world map: %w", err)
	}

	// Find the parent location
	var location *models.Location
	if battleMap.ParentID != "" {
		location = worldMap.GetLocation(battleMap.ParentID)
	}

	// Optionally delete the battle map
	if !req.KeepBattleMap {
		if err := s.mapStore.Delete(ctx, gameState.CurrentMapID); err != nil {
			return nil, fmt.Errorf("failed to delete battle map: %w", err)
		}
	}

	// Update game state to return to world map
	// Set current map to world map and type to world
	gameState.SetCurrentMap(worldMap.ID, models.MapTypeWorld)

	// Update party position to the location if found
	if location != nil {
		if err := gameState.SetPartyPosition(&location.Position); err != nil {
			return nil, fmt.Errorf("failed to set party position: %w", err)
		}
	}

	// Save updated game state
	if err := s.gameStateStore.Update(ctx, gameState); err != nil {
		return nil, fmt.Errorf("failed to update game state: %w", err)
	}

	return &ExitBattleMapResult{
		GameState: gameState,
		Location:  location,
	}, nil
}

// placePartyTokens places party character tokens on the battle map
// Places tokens at the bottom edge of the map for players
func (s *MapServiceWithCharacters) placePartyTokens(ctx context.Context, campaignID string, battleMap *models.Map) (int, error) {
	// Get player characters in the campaign
	isNPC := false
	filter := &store.CharacterFilter{
		CampaignID: campaignID,
		IsNPC:      &isNPC,
	}

	// Get characters using the character store's List method
	type lister interface {
		List(ctx context.Context, filter *store.CharacterFilter) ([]*models.Character, error)
	}

	characterLister, ok := s.characterStore.(lister)
	if !ok {
		return 0, fmt.Errorf("character store does not support List method")
	}

	characters, err := characterLister.List(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to list characters: %w", err)
	}

	tokensPlaced := 0
	startX := 1 // Start from left edge
	startY := battleMap.Grid.Height - 2 // Near bottom edge
	spacing := 2 // Space between tokens

	for _, char := range characters {
		// Check if token already exists
		if battleMap.GetTokenByCharacterID(char.ID) != nil {
			continue
		}

		// Calculate position
		x := startX + (tokensPlaced * spacing)
		y := startY

		// Wrap to next row if needed
		if x >= battleMap.Grid.Width-2 {
			x = startX
			y -= spacing
		}

		// Create token
		token := models.NewToken(char.ID, x, y, models.TokenSizeMedium)

		// Add to battle map
		if err := battleMap.AddToken(*token); err != nil {
			return tokensPlaced, fmt.Errorf("failed to add token for character %s: %w", char.Name, err)
		}

		tokensPlaced++
	}

	return tokensPlaced, nil
}

// ============ Visual Location Methods ============

// CreateVisualLocationRequest represents a visual location creation request
type CreateVisualLocationRequest struct {
	CampaignID  string  `json:"campaign_id"`
	MapID       string  `json:"map_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Type        string  `json:"type"`        // town, dungeon, forest, mountain, etc.
	PositionX   float64 `json:"position_x"`  // 0-1 归一化坐标
	PositionY   float64 `json:"position_y"`  // 0-1 归一化坐标
}

// CreateVisualLocation creates a visual location on an image mode map
func (s *MapService) CreateVisualLocation(ctx context.Context, req *CreateVisualLocationRequest) (*models.VisualLocation, error) {
	// 1. 验证参数
	if req.CampaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}
	if req.MapID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "map ID is required")
	}
	if req.Name == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "name is required")
	}
	if req.Type == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "type is required")
	}

	// 2. 验证坐标范围 (0-1)
	if req.PositionX < 0 || req.PositionX > 1 {
		return nil, NewServiceError(ErrCodeInvalidInput, "position_x must be between 0 and 1")
	}
	if req.PositionY < 0 || req.PositionY > 1 {
		return nil, NewServiceError(ErrCodeInvalidInput, "position_y must be between 0 and 1")
	}

	// 3. 获取地图并验证是 Image 模式
	gameMap, err := s.mapStore.Get(ctx, req.MapID)
	if err != nil {
		return nil, fmt.Errorf("failed to get map: %w", err)
	}

	// 验证地图属于指定战役
	if gameMap.CampaignID != req.CampaignID {
		return nil, NewServiceError(ErrCodeInvalidInput, "map does not belong to the specified campaign")
	}

	// 验证地图是 Image 模式
	if gameMap.Mode != models.MapModeImage {
		return nil, NewServiceError(ErrCodeInvalidInput, "visual locations can only be created on image mode maps")
	}

	// 4. 创建 VisualLocation
	visualLocation := models.NewVisualLocation(
		req.Name,
		req.Description,
		req.Type,
		req.PositionX,
		req.PositionY,
	)

	// 验证 VisualLocation
	if err := visualLocation.Validate(); err != nil {
		return nil, NewServiceError(ErrCodeInvalidInput, fmt.Sprintf("invalid visual location: %v", err))
	}

	// 5. 添加到地图
	if err := gameMap.AddVisualLocation(*visualLocation); err != nil {
		return nil, fmt.Errorf("failed to add visual location: %w", err)
	}

	// 6. 更新地图存储
	if err := s.mapStore.Update(ctx, gameMap); err != nil {
		return nil, fmt.Errorf("failed to update map: %w", err)
	}

	// 7. 返回创建的地点
	return visualLocation, nil
}

// UpdateVisualLocationRequest represents a visual location update request
type UpdateVisualLocationRequest struct {
	CampaignID   string  `json:"campaign_id"`
	MapID        string  `json:"map_id"`
	LocationID   string  `json:"location_id"`
	CustomName   string  `json:"custom_name,omitempty"`   // DM 自定义名称
	Description  string  `json:"description,omitempty"`  // 更新描述
	IsConfirmed  *bool   `json:"is_confirmed,omitempty"` // 确认状态（使用指针支持部分更新）
	BattleMapID  string  `json:"battle_map_id,omitempty"` // 关联战斗地图
}

// UpdateVisualLocation updates an existing visual location on an image mode map
func (s *MapService) UpdateVisualLocation(ctx context.Context, req *UpdateVisualLocationRequest) (*models.VisualLocation, error) {
	// 1. 验证参数
	if req.CampaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}
	if req.MapID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "map ID is required")
	}
	if req.LocationID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "location ID is required")
	}

	// 2. 获取地图
	gameMap, err := s.mapStore.Get(ctx, req.MapID)
	if err != nil {
		return nil, fmt.Errorf("failed to get map: %w", err)
	}

	// 验证地图属于指定战役
	if gameMap.CampaignID != req.CampaignID {
		return nil, NewServiceError(ErrCodeInvalidInput, "map does not belong to the specified campaign")
	}

	// 3. 查找 VisualLocation
	visualLocation := gameMap.GetVisualLocation(req.LocationID)
	if visualLocation == nil {
		return nil, NewServiceError(ErrCodeNotFound, "visual location not found")
	}

	// 4. 更新字段（只更新非零值）
	if req.CustomName != "" {
		visualLocation.SetCustomName(req.CustomName)
	}
	if req.Description != "" {
		visualLocation.Description = req.Description
	}
	if req.IsConfirmed != nil {
		if *req.IsConfirmed {
			visualLocation.Confirm()
		} else {
			visualLocation.IsConfirmed = false
		}
	}
	if req.BattleMapID != "" {
		visualLocation.SetBattleMapID(req.BattleMapID)
	}

	// 5. 验证更新后的地点
	if err := visualLocation.Validate(); err != nil {
		return nil, NewServiceError(ErrCodeInvalidInput, fmt.Sprintf("invalid visual location: %v", err))
	}

	// 6. 保存地图
	if err := s.mapStore.Update(ctx, gameMap); err != nil {
		return nil, fmt.Errorf("failed to update map: %w", err)
	}

	// 7. 返回更新后的地点
	return visualLocation, nil
}
