// Package tools provides MCP tool implementations
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dnd-mcp/server/internal/mcp"
	"github.com/dnd-mcp/server/internal/service"
)

// RestTools provides rest-related MCP tools
type RestTools struct {
	restService *service.RestService
}

// NewRestTools creates a new RestTools instance
func NewRestTools(restService *service.RestService) *RestTools {
	return &RestTools{
		restService: restService,
	}
}

// Register registers all rest tools with the registry
func (t *RestTools) Register(registry *mcp.Registry) {
	registry.MustRegister(t.takeShortRestTool())
	registry.MustRegister(t.takeLongRestTool())
	registry.MustRegister(t.partyLongRestTool())
}

// takeShortRestTool implements the take_short_rest tool
func (t *RestTools) takeShortRestTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"take_short_rest",
		"Take a short rest for a character. A short rest lasts at least 1 hour and allows spending hit dice to recover HP. Rules reference: PHB Chapter 8 - Short Rest.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id": mcp.StringProp("The campaign ID (required)"),
				"character_id": mcp.StringProp("The character ID (required)"),
				"hit_dice_to_spend": mcp.IntProp("Number of hit dice to spend (1-6, cannot exceed available hit dice)"),
			},
			mcp.Required("campaign_id", "character_id", "hit_dice_to_spend"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CampaignID     string `json:"campaign_id"`
			CharacterID    string `json:"character_id"`
			HitDiceToSpend int    `json:"hit_dice_to_spend"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		// Validate hit dice to spend
		if input.HitDiceToSpend < 1 {
			return mcp.NewErrorResponse(fmt.Errorf("hit_dice_to_spend must be at least 1"))
		}
		if input.HitDiceToSpend > 6 {
			return mcp.NewErrorResponse(fmt.Errorf("hit_dice_to_spend cannot exceed 6"))
		}

		restReq := &service.ShortRestRequest{
			CampaignID:     input.CampaignID,
			CharacterID:    input.CharacterID,
			HitDiceToSpend: input.HitDiceToSpend,
		}

		result, err := t.restService.ShortRest(ctx, restReq)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"result": result,
			"message": fmt.Sprintf("%s took a short rest and recovered %d HP.",
				result.CharacterName, result.HPHealed),
		})
	}

	return tool, handler
}

// takeLongRestTool implements the take_long_rest tool
func (t *RestTools) takeLongRestTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"take_long_rest",
		"Take a long rest for a character. A long rest lasts at least 8 hours and restores full HP, half hit dice, all spell slots, and features. Rules reference: PHB Chapter 8 - Long Rest.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id": mcp.StringProp("The campaign ID (required)"),
				"character_id": mcp.StringProp("The character ID (required)"),
			},
			mcp.Required("campaign_id", "character_id"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CampaignID  string `json:"campaign_id"`
			CharacterID string `json:"character_id"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		restReq := &service.LongRestRequest{
			CampaignID:  input.CampaignID,
			CharacterID: input.CharacterID,
		}

		result, err := t.restService.LongRest(ctx, restReq)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		message := fmt.Sprintf("%s took a long rest and recovered to full HP (%d/%d).",
			result.CharacterName, result.HPCurrent, result.HPMax)
		if result.HitDiceRestored > 0 {
			message += fmt.Sprintf(" Regained %d hit dice (now has %d/%d).",
				result.HitDiceRestored, result.HitDiceRemaining, result.HitDiceRestored+result.HitDiceRemaining)
		}
		if result.SpellSlotsRestored {
			message += " All spell slots restored."
		}
		if len(result.FeaturesRestored) > 0 {
			message += fmt.Sprintf(" Features restored: %v.", result.FeaturesRestored)
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"result":  result,
			"message": message,
		})
	}

	return tool, handler
}

// partyLongRestTool implements the party_long_rest tool
func (t *RestTools) partyLongRestTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"party_long_rest",
		"Take a long rest for all player characters in the campaign. Restores full HP, half hit dice, all spell slots, and features for each character. Advances game time by 8 hours. Rules reference: PHB Chapter 8 - Long Rest.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id": mcp.StringProp("The campaign ID (required)"),
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

		gameState, results, err := t.restService.PartyLongRest(ctx, input.CampaignID)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		// Build summary message
		message := fmt.Sprintf("Party completed a long rest. %d characters rested.\n", len(results))
		for _, result := range results {
			message += fmt.Sprintf("- %s: HP %d/%d, Hit Dice %d restored",
				result.CharacterName, result.HPCurrent, result.HPMax, result.HitDiceRestored)
			if result.SpellSlotsRestored {
				message += ", spell slots restored"
			}
			if len(result.FeaturesRestored) > 0 {
				message += fmt.Sprintf(", features: %v", result.FeaturesRestored)
			}
			message += "\n"
		}

		response := map[string]interface{}{
			"results": results,
			"message": message,
		}

		// Include game state if available
		if gameState != nil && gameState.GameTime != nil {
			response["game_time"] = gameState.GameTime
		}

		return mcp.NewJSONResponse(response)
	}

	return tool, handler
}

// Tool list for external registration
var RestToolNames = []string{
	"take_short_rest",
	"take_long_rest",
	"party_long_rest",
}
