// Package tools provides MCP tool implementations
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dnd-mcp/server/internal/mcp"
	"github.com/dnd-mcp/server/internal/service"
)

// ConditionTools provides condition-related MCP tools
type ConditionTools struct {
	conditionService *service.ConditionService
}

// NewConditionTools creates a new ConditionTools instance
func NewConditionTools(conditionService *service.ConditionService) *ConditionTools {
	return &ConditionTools{
		conditionService: conditionService,
	}
}

// Register registers all condition tools with the registry
func (t *ConditionTools) Register(registry *mcp.Registry) {
	registry.MustRegister(t.applyConditionTool())
	registry.MustRegister(t.removeConditionTool())
	registry.MustRegister(t.getConditionsTool())
	registry.MustRegister(t.hasConditionTool())
}

// applyConditionTool implements the apply_condition tool
func (t *ConditionTools) applyConditionTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"apply_condition",
		"Apply a condition effect to a character. Conditions follow D&D 5e rules from PHB Appendix A. Duration is in rounds (-1 for permanent).",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id": mcp.StringProp("The ID of the campaign (required)"),
				"character_id": mcp.StringProp("The ID of the character to apply the condition to (required)"),
				"condition_type": mcp.StringProp("The type of condition (required). Valid types: blinded, charmed, deafened, frightened, grappled, incapacitated, invisible, paralyzed, petrified, poisoned, prone, restrained, stunned, unconscious, exhaustion"),
				"duration": mcp.IntProp("Duration in rounds (-1 for permanent, default is -1)"),
				"source": mcp.StringProp("The source of the condition (e.g., 'spider poison', 'spell')"),
				"exhaustion_level": mcp.IntProp("The exhaustion level (1-6, only for exhaustion condition)"),
			},
			mcp.Required("campaign_id", "character_id", "condition_type"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CampaignID      string `json:"campaign_id"`
			CharacterID     string `json:"character_id"`
			ConditionType   string `json:"condition_type"`
			Duration        int    `json:"duration"`
			Source          string `json:"source"`
			ExhaustionLevel int    `json:"exhaustion_level"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		applyReq := &service.ApplyConditionRequest{
			CampaignID:      input.CampaignID,
			CharacterID:     input.CharacterID,
			ConditionType:   input.ConditionType,
			Duration:        input.Duration,
			Source:          input.Source,
			ExhaustionLevel: input.ExhaustionLevel,
		}

		resp, err := t.conditionService.ApplyCondition(ctx, applyReq)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		// Build conditions summary
		conditions := make([]map[string]interface{}, len(resp.Conditions))
		for i, c := range resp.Conditions {
			conditions[i] = map[string]interface{}{
				"type":     c.Type,
				"duration": c.Duration,
				"source":   c.Source,
			}
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"character": map[string]interface{}{
				"id":         resp.Character.ID,
				"name":       resp.Character.Name,
				"conditions": conditions,
			},
			"applied": resp.Applied,
			"message": resp.Message,
		})
	}

	return tool, handler
}

// removeConditionTool implements the remove_condition tool
func (t *ConditionTools) removeConditionTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"remove_condition",
		"Remove a condition from a character. Can remove a specific condition or all conditions.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id": mcp.StringProp("The ID of the campaign (required)"),
				"character_id": mcp.StringProp("The ID of the character to remove the condition from (required)"),
				"condition_type": mcp.StringProp("The type of condition to remove"),
				"remove_all": mcp.BoolProp("Remove all conditions from the character"),
			},
			mcp.Required("campaign_id", "character_id"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CampaignID    string `json:"campaign_id"`
			CharacterID   string `json:"character_id"`
			ConditionType string `json:"condition_type"`
			RemoveAll     bool   `json:"remove_all"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		removeReq := &service.RemoveConditionRequest{
			CampaignID:    input.CampaignID,
			CharacterID:   input.CharacterID,
			ConditionType: input.ConditionType,
			RemoveAll:     input.RemoveAll,
		}

		resp, err := t.conditionService.RemoveCondition(ctx, removeReq)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		// Build conditions summary
		conditions := make([]map[string]interface{}, len(resp.Conditions))
		for i, c := range resp.Conditions {
			conditions[i] = map[string]interface{}{
				"type":     c.Type,
				"duration": c.Duration,
				"source":   c.Source,
			}
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"character": map[string]interface{}{
				"id":         resp.Character.ID,
				"name":       resp.Character.Name,
				"conditions": conditions,
			},
			"removed": resp.Removed,
			"message": resp.Message,
		})
	}

	return tool, handler
}

// getConditionsTool implements the get_conditions tool
func (t *ConditionTools) getConditionsTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"get_conditions",
		"Get all conditions on a character and their effects. Returns detailed information about how conditions affect the character including advantages, disadvantages, and special effects.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id": mcp.StringProp("The ID of the campaign (required)"),
				"character_id": mcp.StringProp("The ID of the character (required)"),
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

		getReq := &service.GetConditionEffectsRequest{
			CampaignID:  input.CampaignID,
			CharacterID: input.CharacterID,
		}

		resp, err := t.conditionService.GetConditionEffects(ctx, getReq)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"character": map[string]interface{}{
				"id":   resp.CharacterID,
				"name": resp.CharacterName,
			},
			"conditions":         resp.Conditions,
			"total_disadvantage": resp.TotalDisadvantage,
			"total_advantage":    resp.TotalAdvantage,
			"cannot_act":         resp.CannotAct,
			"incapped":           resp.Incapped,
		})
	}

	return tool, handler
}

// hasConditionTool implements the has_condition tool
func (t *ConditionTools) hasConditionTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"has_condition",
		"Check if a character has a specific condition.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"character_id": mcp.StringProp("The ID of the character (required)"),
				"condition_type": mcp.StringProp("The type of condition to check (required)"),
			},
			mcp.Required("character_id", "condition_type"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CharacterID   string `json:"character_id"`
			ConditionType string `json:"condition_type"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		hasReq := &service.HasConditionRequest{
			CharacterID:   input.CharacterID,
			ConditionType: input.ConditionType,
		}

		resp, err := t.conditionService.HasCondition(ctx, hasReq)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"has_condition": resp.HasCondition,
			"message":       resp.Message,
		})
	}

	return tool, handler
}

// Tool list for external registration
var ConditionToolNames = []string{
	"apply_condition",
	"remove_condition",
	"get_conditions",
	"has_condition",
}
