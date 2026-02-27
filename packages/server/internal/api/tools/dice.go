// Package tools provides MCP tool implementations
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dnd-mcp/server/internal/mcp"
	"github.com/dnd-mcp/server/internal/service"
)

// DiceTools provides dice-related MCP tools
type DiceTools struct {
	diceService *service.DiceService
}

// NewDiceTools creates a new DiceTools instance
func NewDiceTools(diceService *service.DiceService) *DiceTools {
	return &DiceTools{
		diceService: diceService,
	}
}

// Register registers all dice tools with the registry
func (t *DiceTools) Register(registry *mcp.Registry) {
	registry.MustRegister(t.rollDiceTool())
	registry.MustRegister(t.rollCheckTool())
	registry.MustRegister(t.rollSaveTool())
}

// rollDiceTool implements the roll_dice tool
func (t *DiceTools) rollDiceTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"roll_dice",
		"Roll dice using standard D&D notation. Supports formulas like '1d20+5', '2d6', '4d6kh3' (keep highest 3), '2d20kh1' (advantage), '2d20kl1' (disadvantage).",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"formula": mcp.StringProp("The dice formula to roll (e.g., '1d20+5', '2d6', '4d6kh3') (required)"),
			},
			mcp.Required("formula"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			Formula string `json:"formula"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		resp, err := t.diceService.RollDice(ctx, &service.RollDiceRequest{
			Formula: input.Formula,
		})
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"result": resp.Result,
		})
	}

	return tool, handler
}

// rollCheckTool implements the roll_check tool
func (t *DiceTools) rollCheckTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"roll_check",
		"Perform an ability or skill check for a character. Rolls d20, adds ability modifier and proficiency bonus if applicable. Supports advantage/disadvantage.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"character_id":  mcp.StringProp("The ID of the character making the check (required)"),
				"ability":       mcp.StringProp("The ability to use (strength, dexterity, constitution, intelligence, wisdom, charisma) (required)"),
				"skill":         mcp.StringProp("Optional skill to apply proficiency bonus (e.g., athletics, stealth, perception)"),
				"dc":            mcp.IntProp("Optional difficulty class (DC) to compare against"),
				"advantage":     mcp.BoolProp("Roll with advantage (roll 2d20, take higher)"),
				"disadvantage":  mcp.BoolProp("Roll with disadvantage (roll 2d20, take lower)"),
			},
			mcp.Required("character_id", "ability"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CharacterID  string `json:"character_id"`
			Ability      string `json:"ability"`
			Skill        string `json:"skill"`
			DC           int    `json:"dc"`
			Advantage    bool   `json:"advantage"`
			Disadvantage bool   `json:"disadvantage"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		resp, err := t.diceService.RollCheck(ctx, &service.RollCheckRequest{
			CharacterID:  input.CharacterID,
			Ability:      input.Ability,
			Skill:        input.Skill,
			DC:           input.DC,
			Advantage:    input.Advantage,
			Disadvantage: input.Disadvantage,
		})
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		// Build response message
		result := map[string]interface{}{
			"result": resp.Result,
		}

		// Add human-readable message
		msg := fmt.Sprintf("Rolled %s check", input.Ability)
		if input.Skill != "" {
			msg = fmt.Sprintf("Rolled %s (%s) check", input.Skill, input.Ability)
		}
		msg += fmt.Sprintf(": %d", resp.Result.DiceResult.Total)

		if input.DC > 0 {
			if resp.Result.Success {
				msg += fmt.Sprintf(" vs DC %d - SUCCESS", input.DC)
			} else {
				msg += fmt.Sprintf(" vs DC %d - FAILURE", input.DC)
			}
		}

		if resp.Result.DiceResult.IsCritical() {
			msg += " [CRITICAL!]"
		} else if resp.Result.DiceResult.IsFumble() {
			msg += " [FUMBLE!]"
		}

		result["message"] = msg

		return mcp.NewJSONResponse(result)
	}

	return tool, handler
}

// rollSaveTool implements the roll_save tool
func (t *DiceTools) rollSaveTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"roll_save",
		"Perform a saving throw for a character. Rolls d20, adds ability modifier and proficiency bonus if the character has proficiency in that save. Supports advantage/disadvantage.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"character_id":  mcp.StringProp("The ID of the character making the save (required)"),
				"ability":       mcp.StringProp("The ability for the saving throw (strength, dexterity, constitution, intelligence, wisdom, charisma) (required)"),
				"dc":            mcp.IntProp("Optional difficulty class (DC) to compare against"),
				"advantage":     mcp.BoolProp("Roll with advantage (roll 2d20, take higher)"),
				"disadvantage":  mcp.BoolProp("Roll with disadvantage (roll 2d20, take lower)"),
			},
			mcp.Required("character_id", "ability"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CharacterID  string `json:"character_id"`
			Ability      string `json:"ability"`
			DC           int    `json:"dc"`
			Advantage    bool   `json:"advantage"`
			Disadvantage bool   `json:"disadvantage"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		resp, err := t.diceService.RollSave(ctx, &service.RollSaveRequest{
			CharacterID:  input.CharacterID,
			Ability:      input.Ability,
			DC:           input.DC,
			Advantage:    input.Advantage,
			Disadvantage: input.Disadvantage,
		})
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		// Build response message
		result := map[string]interface{}{
			"result": resp.Result,
		}

		// Add human-readable message
		msg := fmt.Sprintf("Rolled %s saving throw: %d", input.Ability, resp.Result.DiceResult.Total)

		if input.DC > 0 {
			if resp.Result.Success {
				msg += fmt.Sprintf(" vs DC %d - SUCCESS", input.DC)
			} else {
				msg += fmt.Sprintf(" vs DC %d - FAILURE", input.DC)
			}
		}

		if resp.Result.DiceResult.IsCritical() {
			msg += " [CRITICAL!]"
		} else if resp.Result.DiceResult.IsFumble() {
			msg += " [FUMBLE!]"
		}

		result["message"] = msg

		return mcp.NewJSONResponse(result)
	}

	return tool, handler
}

// Tool list for external registration
var DiceToolNames = []string{
	"roll_dice",
	"roll_check",
	"roll_save",
}
