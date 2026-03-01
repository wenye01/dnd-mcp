// Package tools provides MCP tool implementations
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dnd-mcp/server/internal/mcp"
	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/service"
)

// MapTools provides map-related MCP tools
type MapTools struct {
	mapService          *service.MapService
	mapServiceWithChars *service.MapServiceWithCharacters
}

// NewMapTools creates a new MapTools instance
func NewMapTools(mapService *service.MapService) *MapTools {
	return &MapTools{
		mapService: mapService,
	}
}

// NewMapToolsWithCharacters creates a new MapTools instance with character support
func NewMapToolsWithCharacters(mapService *service.MapServiceWithCharacters) *MapTools {
	return &MapTools{
		mapServiceWithChars: mapService,
	}
}

// Register registers all map tools with the registry
func (t *MapTools) Register(registry *mcp.Registry) {
	registry.MustRegister(t.getWorldMapTool())
	registry.MustRegister(t.getMoveToTool())
	registry.MustRegister(t.getMoveTokenTool())
	registry.MustRegister(t.getEnterBattleMapTool())
	registry.MustRegister(t.getGetBattleMapTool())
	registry.MustRegister(t.getExitBattleMapTool())
	registry.MustRegister(t.getCreateVisualLocationTool())
	registry.MustRegister(t.getUpdateVisualLocationTool())
}

// Tool definitions

func (t *MapTools) getWorldMapTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"get_world_map",
		"Get the world map for a campaign, including locations and terrain information. The world map shows the overall geography where the adventure takes place. Returns different data based on map mode (grid or image).",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id": mcp.StringProp("The ID of the campaign to get the world map for (required)"),
			},
			mcp.Required("campaign_id"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CampaignID string `json:"campaign_id"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		// Get world map - use mapServiceWithChars if available, otherwise mapService
		var worldMap *models.Map
		var err error
		if t.mapServiceWithChars != nil {
			worldMap, err = t.mapServiceWithChars.GetWorldMap(ctx, input.CampaignID)
		} else if t.mapService != nil {
			worldMap, err = t.mapService.GetWorldMap(ctx, input.CampaignID)
		} else {
			return mcp.NewErrorResponse(fmt.Errorf("map service not configured"))
		}

		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		// Try to get player marker from game state (for Image mode)
		var playerMarker *models.PlayerMarker
		if worldMap.Mode == models.MapModeImage && t.mapServiceWithChars != nil {
			// MapServiceWithCharacters embeds MapService, so we can call GetWorldMapWithPlayerMarker on it
			_, playerMarker, _ = t.mapServiceWithChars.GetWorldMapWithPlayerMarker(ctx, input.CampaignID)
		}

		// Build response based on map mode
		response := map[string]interface{}{
			"map":     worldMap,
			"mode":    string(worldMap.Mode),
			"message": fmt.Sprintf("Retrieved world map '%s' for campaign", worldMap.Name),
		}

		// Grid mode: return grid data and locations
		if worldMap.Mode == models.MapModeGrid {
			response["locations"] = worldMap.Locations
			if worldMap.Grid != nil {
				response["grid"] = map[string]interface{}{
					"width":  worldMap.Grid.Width,
					"height": worldMap.Grid.Height,
					"cells":  worldMap.Grid.Cells,
				}
			}
		}

		// Image mode: return image data, visual locations, and player marker
		if worldMap.Mode == models.MapModeImage {
			if worldMap.Image != nil {
				response["image"] = map[string]interface{}{
					"url":    worldMap.Image.URL,
					"width":  worldMap.Image.Width,
					"height": worldMap.Image.Height,
				}
			}
			response["visual_locations"] = worldMap.VisualLocations

			// Add player marker if available
			if playerMarker != nil {
				markerData := map[string]interface{}{
					"position_x": playerMarker.PositionX,
					"position_y": playerMarker.PositionY,
				}
				if playerMarker.CurrentScene != "" {
					markerData["current_scene"] = playerMarker.CurrentScene
				}
				response["player_marker"] = markerData
			}
		}

		return mcp.NewJSONResponse(response)
	}

	return tool, handler
}

func (t *MapTools) getMoveToTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"move_to",
		"Move the party to a specific position on the world map. This updates the party's location and advances game time based on travel distance. Uses normal travel pace by default (3 mph / 24 miles per day). For Grid mode maps, use x/y coordinates. For Image mode maps, use target_x/target_y coordinates (normalized 0-1).",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id": mcp.StringProp("The ID of the campaign (required)"),
				"x":           mcp.IntProp("The X coordinate on the world map grid (for Grid mode, required with y)"),
				"y":           mcp.IntProp("The Y coordinate on the world map grid (for Grid mode, required with x)"),
				"pace": mcp.PropWithEnum(
					"Travel pace: 'fast' (4 mph, -5 passive Perception), 'normal' (3 mph), or 'slow' (2 mph, can use Stealth)",
					"fast", "normal", "slow",
				),
				"target_x": mcp.Prop("number", "The X coordinate as normalized value between 0 and 1 (for Image mode, required with target_y)"),
				"target_y": mcp.Prop("number", "The Y coordinate as normalized value between 0 and 1 (for Image mode, required with target_x)"),
			},
			mcp.Required("campaign_id"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CampaignID string  `json:"campaign_id"`
			X          int     `json:"x"`
			Y          int     `json:"y"`
			Pace       string  `json:"pace"`
			TargetX    float64 `json:"target_x"`
			TargetY    float64 `json:"target_y"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		// Set default pace if not provided
		if input.Pace == "" {
			input.Pace = "normal"
		}

		moveReq := &service.MoveToRequest{
			CampaignID: input.CampaignID,
			X:          input.X,
			Y:          input.Y,
			Pace:       input.Pace,
			TargetX:    input.TargetX,
			TargetY:    input.TargetY,
		}

		// Perform the move
		result, err := t.mapService.MoveTo(ctx, moveReq)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		// Build response based on mode
		response := map[string]interface{}{
			"game_state": map[string]interface{}{
				"game_time": result.GameState.GameTime,
			},
		}

		// Grid mode response
		if result.TravelResult != nil {
			response["message"] = fmt.Sprintf("Party moved to position (%d, %d)", input.X, input.Y)
			response["position"] = map[string]interface{}{
				"x": input.X,
				"y": input.Y,
			}
			response["travel"] = map[string]interface{}{
				"distance":    result.TravelResult.Distance,
				"hours":       result.TravelResult.Hours,
				"days":        result.TravelResult.Days,
				"pace":        result.TravelResult.Pace,
				"description": result.TravelResult.Description,
			}
		}

		// Image mode response
		if result.NewMarker != nil {
			response["message"] = fmt.Sprintf("Party moved to position (%.2f, %.2f)", input.TargetX, input.TargetY)
			response["new_marker"] = map[string]interface{}{
				"position_x": result.NewMarker.PositionX,
				"position_y": result.NewMarker.PositionY,
			}
			if result.NewMarker.CurrentScene != "" {
				response["new_marker"].(map[string]interface{})["current_scene"] = result.NewMarker.CurrentScene
			}
		}

		return mcp.NewJSONResponse(response)
	}

	return tool, handler
}

func (t *MapTools) getMoveTokenTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"move_token",
		"Move a token on a battle map. Calculates movement cost considering terrain and obstacles. Rules: 1 square = 5 feet, difficult terrain costs double, diagonal movement has +50% cost. Size-based movement: can move through creatures 2+ sizes smaller.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id": mcp.StringProp("The ID of the campaign (required)"),
				"map_id":      mcp.StringProp("The ID of the battle map (required)"),
				"token_id":    mcp.StringProp("The ID of the token to move (required)"),
				"to_x":        mcp.IntProp("The destination X coordinate on the grid (required, non-negative)"),
				"to_y":        mcp.IntProp("The destination Y coordinate on the grid (required, non-negative)"),
				"speed":       mcp.IntProp("Available movement speed in feet for this turn (optional, defaults to character speed)"),
			},
			mcp.Required("campaign_id", "map_id", "token_id", "to_x", "to_y"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CampaignID string `json:"campaign_id"`
			MapID      string `json:"map_id"`
			TokenID    string `json:"token_id"`
			ToX        int    `json:"to_x"`
			ToY        int    `json:"to_y"`
			Speed      *int   `json:"speed"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		moveReq := &service.TokenMoveRequest{
			CampaignID: input.CampaignID,
			MapID:      input.MapID,
			TokenID:    input.TokenID,
			ToX:        input.ToX,
			ToY:        input.ToY,
			Speed:      input.Speed,
		}

		// Perform the move
		result, err := t.mapService.MoveToken(ctx, moveReq)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"message": fmt.Sprintf("Token moved to position (%d, %d), using %d feet of movement", input.ToX, input.ToY, result.MovementUsed),
			"token": map[string]interface{}{
				"id":           result.Token.ID,
				"character_id": result.Token.CharacterID,
				"position": map[string]interface{}{
					"x": result.Token.Position.X,
					"y": result.Token.Position.Y,
				},
				"size": result.Token.Size,
			},
			"movement": map[string]interface{}{
				"used":              result.MovementUsed,
				"remaining":         result.RemainingSpeed,
				"difficult_terrain": result.DifficultTerrainCount,
			},
			"path": result.Path,
		})
	}

	return tool, handler
}

func (t *MapTools) getEnterBattleMapTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"enter_battle_map",
		"Enter a battle map from the world map. This transitions the party from world exploration to tactical combat mode. If no battle map exists for the location, one can be created automatically. Party character tokens will be placed at the edge of the battle map.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id":       mcp.StringProp("The ID of the campaign (required)"),
				"location_id":       mcp.StringProp("The ID of the location on the world map to enter (required)"),
				"battle_map_id":     mcp.StringProp("The ID of an existing battle map to enter (optional)"),
				"create_if_missing": mcp.BoolProp("If true, creates a new battle map if none exists for the location (optional, default false)"),
			},
			mcp.Required("campaign_id", "location_id"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CampaignID      string `json:"campaign_id"`
			LocationID      string `json:"location_id"`
			BattleMapID     string `json:"battle_map_id"`
			CreateIfMissing bool   `json:"create_if_missing"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		// Check if we have the extended map service
		if t.mapServiceWithChars == nil {
			return mcp.NewErrorResponse(fmt.Errorf("map switching not available - character store not configured"))
		}

		enterReq := &service.EnterBattleMapRequest{
			CampaignID:      input.CampaignID,
			LocationID:      input.LocationID,
			BattleMapID:     input.BattleMapID,
			CreateIfMissing: input.CreateIfMissing,
		}

		result, err := t.mapServiceWithChars.EnterBattleMap(ctx, enterReq)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"message": fmt.Sprintf("Entered battle map '%s' with %d party tokens placed", result.BattleMap.Name, result.TokensPlaced),
			"battle_map": map[string]interface{}{
				"id":     result.BattleMap.ID,
				"name":   result.BattleMap.Name,
				"type":   result.BattleMap.Type,
				"width":  result.BattleMap.Grid.Width,
				"height": result.BattleMap.Grid.Height,
				"tokens": result.BattleMap.Tokens,
			},
			"game_state": map[string]interface{}{
				"current_map_id":   result.GameState.CurrentMapID,
				"current_map_type": result.GameState.CurrentMapType,
			},
			"tokens_placed": result.TokensPlaced,
		})
	}

	return tool, handler
}

func (t *MapTools) getGetBattleMapTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"get_battle_map",
		"Get the current battle map for the campaign, including all token positions. This can only be used when the party is currently in a battle map.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id": mcp.StringProp("The ID of the campaign (required)"),
			},
			mcp.Required("campaign_id"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CampaignID string `json:"campaign_id"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		var battleMap *models.Map
		var err error

		// Try using mapServiceWithChars first (has embedded MapService)
		if t.mapServiceWithChars != nil {
			battleMap, err = t.mapServiceWithChars.GetBattleMapByCampaign(ctx, input.CampaignID)
		} else if t.mapService != nil {
			battleMap, err = t.mapService.GetBattleMapByCampaign(ctx, input.CampaignID)
		} else {
			err = fmt.Errorf("map service not configured")
		}

		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"message": fmt.Sprintf("Retrieved battle map '%s'", battleMap.Name),
			"battle_map": map[string]interface{}{
				"id":        battleMap.ID,
				"name":      battleMap.Name,
				"type":      battleMap.Type,
				"width":     battleMap.Grid.Width,
				"height":    battleMap.Grid.Height,
				"cell_size": battleMap.Grid.CellSize,
				"tokens":    battleMap.Tokens,
			},
		})
	}

	return tool, handler
}

func (t *MapTools) getExitBattleMapTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"exit_battle_map",
		"Exit the current battle map and return to the world map. This ends the tactical combat mode and returns to world exploration. The battle map can be kept for later use or deleted.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id":     mcp.StringProp("The ID of the campaign (required)"),
				"keep_battle_map": mcp.BoolProp("If true, the battle map is saved for future use; if false, it is deleted (optional, default false)"),
			},
			mcp.Required("campaign_id"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CampaignID    string `json:"campaign_id"`
			KeepBattleMap bool   `json:"keep_battle_map"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		exitReq := &service.ExitBattleMapRequest{
			CampaignID:    input.CampaignID,
			KeepBattleMap: input.KeepBattleMap,
		}

		var result *service.ExitBattleMapResult
		var err error

		// Try using mapServiceWithChars first (has embedded MapService)
		if t.mapServiceWithChars != nil {
			result, err = t.mapServiceWithChars.ExitBattleMap(ctx, exitReq)
		} else if t.mapService != nil {
			result, err = t.mapService.ExitBattleMap(ctx, exitReq)
		} else {
			err = fmt.Errorf("map service not configured")
		}

		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		response := map[string]interface{}{
			"message": "Exited battle map and returned to world map",
			"game_state": map[string]interface{}{
				"current_map_id":   result.GameState.CurrentMapID,
				"current_map_type": result.GameState.CurrentMapType,
				"party_position":   result.GameState.PartyPosition,
			},
		}

		if result.Location != nil {
			response["location"] = map[string]interface{}{
				"id":          result.Location.ID,
				"name":        result.Location.Name,
				"description": result.Location.Description,
				"position":    result.Location.Position,
			}
		}

		return mcp.NewJSONResponse(response)
	}

	return tool, handler
}

func (t *MapTools) getCreateVisualLocationTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"create_visual_location",
		"Create a visual location on an image mode world map. Visual locations use normalized coordinates (0-1) to mark points of interest identified through image analysis, such as towns, dungeons, forests, or mountains.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id": mcp.StringProp("The ID of the campaign (required)"),
				"map_id":      mcp.StringProp("The ID of the image mode map (required)"),
				"name":        mcp.StringProp("The name of the location (required)"),
				"description": mcp.StringProp("A description of the location (optional)"),
				"type":        mcp.StringProp("The type of location: town, dungeon, forest, mountain, etc. (required)"),
				"position_x":  mcp.Prop("number", "X coordinate as normalized value between 0 and 1 (required)"),
				"position_y":  mcp.Prop("number", "Y coordinate as normalized value between 0 and 1 (required)"),
			},
			mcp.Required("campaign_id", "map_id", "name", "type", "position_x", "position_y"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CampaignID  string  `json:"campaign_id"`
			MapID       string  `json:"map_id"`
			Name        string  `json:"name"`
			Description string  `json:"description"`
			Type        string  `json:"type"`
			PositionX   float64 `json:"position_x"`
			PositionY   float64 `json:"position_y"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		createReq := &service.CreateVisualLocationRequest{
			CampaignID:  input.CampaignID,
			MapID:       input.MapID,
			Name:        input.Name,
			Description: input.Description,
			Type:        input.Type,
			PositionX:   input.PositionX,
			PositionY:   input.PositionY,
		}

		visualLocation, err := t.mapService.CreateVisualLocation(ctx, createReq)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"message": fmt.Sprintf("Created visual location '%s' at position (%.2f, %.2f)", visualLocation.Name, visualLocation.PositionX, visualLocation.PositionY),
			"visual_location": map[string]interface{}{
				"id":           visualLocation.ID,
				"name":         visualLocation.Name,
				"description":  visualLocation.Description,
				"type":         visualLocation.Type,
				"position_x":   visualLocation.PositionX,
				"position_y":   visualLocation.PositionY,
				"is_confirmed": visualLocation.IsConfirmed,
			},
		})
	}

	return tool, handler
}

func (t *MapTools) getUpdateVisualLocationTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"update_location",
		"Update a visual location on an image mode world map. Allows the DM to modify location details such as custom name, description, confirmation status, or associated battle map.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id":   mcp.StringProp("The ID of the campaign (required)"),
				"map_id":        mcp.StringProp("The ID of the image mode map (required)"),
				"location_id":   mcp.StringProp("The ID of the visual location to update (required)"),
				"custom_name":   mcp.StringProp("Custom name for the location set by the DM (optional)"),
				"description":   mcp.StringProp("Updated description of the location (optional)"),
				"is_confirmed":  mcp.BoolProp("Whether the DM has confirmed this location (optional)"),
				"battle_map_id": mcp.StringProp("ID of the battle map to associate with this location (optional)"),
			},
			mcp.Required("campaign_id", "map_id", "location_id"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CampaignID  string `json:"campaign_id"`
			MapID       string `json:"map_id"`
			LocationID  string `json:"location_id"`
			CustomName  string `json:"custom_name,omitempty"`
			Description string `json:"description,omitempty"`
			IsConfirmed *bool  `json:"is_confirmed,omitempty"`
			BattleMapID string `json:"battle_map_id,omitempty"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		updateReq := &service.UpdateVisualLocationRequest{
			CampaignID:  input.CampaignID,
			MapID:       input.MapID,
			LocationID:  input.LocationID,
			CustomName:  input.CustomName,
			Description: input.Description,
			IsConfirmed: input.IsConfirmed,
			BattleMapID: input.BattleMapID,
		}

		visualLocation, err := t.mapService.UpdateVisualLocation(ctx, updateReq)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		displayName := visualLocation.GetDisplayName()
		return mcp.NewJSONResponse(map[string]interface{}{
			"message": fmt.Sprintf("Updated visual location '%s'", displayName),
			"visual_location": map[string]interface{}{
				"id":            visualLocation.ID,
				"name":          visualLocation.Name,
				"custom_name":   visualLocation.CustomName,
				"display_name":  displayName,
				"description":   visualLocation.Description,
				"type":          visualLocation.Type,
				"position_x":    visualLocation.PositionX,
				"position_y":    visualLocation.PositionY,
				"is_confirmed":  visualLocation.IsConfirmed,
				"battle_map_id": visualLocation.BattleMapID,
			},
		})
	}

	return tool, handler
}

// Tool list for external registration
var MapToolNames = []string{
	"get_world_map",
	"move_to",
	"move_token",
	"enter_battle_map",
	"get_battle_map",
	"exit_battle_map",
	"create_visual_location",
	"update_location",
}
