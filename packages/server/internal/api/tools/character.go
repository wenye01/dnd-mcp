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

// CharacterTools provides character-related MCP tools
type CharacterTools struct {
	characterService *service.CharacterService
}

// NewCharacterTools creates a new CharacterTools instance
func NewCharacterTools(characterService *service.CharacterService) *CharacterTools {
	return &CharacterTools{
		characterService: characterService,
	}
}

// Register registers all character tools with the registry
func (t *CharacterTools) Register(registry *mcp.Registry) {
	registry.MustRegister(t.createCharacterTool())
	registry.MustRegister(t.getCharacterTool())
	registry.MustRegister(t.updateCharacterTool())
	registry.MustRegister(t.listCharactersTool())
	registry.MustRegister(t.deleteCharacterTool())
}

// Tool definitions

func (t *CharacterTools) createCharacterTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"create_character",
		"Create a new player character or NPC in a campaign. For player characters, player_id is required. For NPCs, set is_npc=true and optionally specify npc_type (scripted or generated).",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id": mcp.StringProp("The ID of the campaign this character belongs to (required)"),
				"name":        mcp.StringProp("The character's name (required)"),
				"is_npc":      mcp.BoolProp("Whether this is an NPC (default: false)"),
				"npc_type": mcp.PropWithEnum(
					"The type of NPC: 'scripted' for story NPCs, 'generated' for improvised NPCs",
					"scripted", "generated",
				),
				"player_id":  mcp.StringProp("The player's user ID (required for player characters)"),
				"race":       mcp.StringProp("The character's race (required)"),
				"class":      mcp.StringProp("The character's class (required)"),
				"level":      mcp.IntProp("The character's level (default: 1, range: 1-20)"),
				"background": mcp.StringProp("The character's background"),
				"alignment":  mcp.StringProp("The character's alignment (e.g., 'Lawful Good')"),
				"abilities":  mcp.ObjectProp("Ability scores: strength, dexterity, constitution, intelligence, wisdom, charisma (default: standard array 15,14,13,12,10,8)"),
				"hp":         mcp.ObjectProp("HP values: current, max, temp"),
				"ac":         mcp.IntProp("Armor Class"),
				"speed":      mcp.IntProp("Movement speed in feet (default: 30)"),
				"initiative": mcp.IntProp("Initiative bonus"),
				"skills":     mcp.ObjectProp("Skill bonuses as key-value pairs"),
				"saves":      mcp.ObjectProp("Saving throw bonuses as key-value pairs"),
			},
			mcp.Required("campaign_id", "name", "race", "class"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CampaignID string            `json:"campaign_id"`
			Name       string            `json:"name"`
			IsNPC      bool              `json:"is_npc"`
			NPCType    models.NPCType    `json:"npc_type"`
			PlayerID   string            `json:"player_id"`
			Race       string            `json:"race"`
			Class      string            `json:"class"`
			Level      int               `json:"level"`
			Background string            `json:"background"`
			Alignment  string            `json:"alignment"`
			Abilities  *models.Abilities `json:"abilities"`
			HP         *models.HP        `json:"hp"`
			AC         int               `json:"ac"`
			Speed      int               `json:"speed"`
			Initiative int               `json:"initiative"`
			Skills     map[string]int    `json:"skills"`
			Saves      map[string]int    `json:"saves"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		createReq := &service.CreateCharacterRequest{
			CampaignID: input.CampaignID,
			Name:       input.Name,
			IsNPC:      input.IsNPC,
			NPCType:    input.NPCType,
			PlayerID:   input.PlayerID,
			Race:       input.Race,
			Class:      input.Class,
			Level:      input.Level,
			Background: input.Background,
			Alignment:  input.Alignment,
			Abilities:  input.Abilities,
			HP:         input.HP,
			AC:         input.AC,
			Speed:      input.Speed,
			Initiative: input.Initiative,
			Skills:     input.Skills,
			Saves:      input.Saves,
		}

		character, err := t.characterService.CreateCharacter(ctx, createReq)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"character": character,
			"message":   fmt.Sprintf("Character '%s' created successfully with ID: %s", character.Name, character.ID),
		})
	}

	return tool, handler
}

func (t *CharacterTools) getCharacterTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"get_character",
		"Get detailed information about a specific character by its ID.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"character_id": mcp.StringProp("The unique ID of the character (required)"),
			},
			mcp.Required("character_id"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CharacterID string `json:"character_id"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		character, err := t.characterService.GetCharacter(ctx, input.CharacterID)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"character": character,
		})
	}

	return tool, handler
}

func (t *CharacterTools) updateCharacterTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"update_character",
		"Update character information. Only provided fields will be updated. For HP changes, consider using the HP-specific operations.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"character_id": mcp.StringProp("The unique ID of the character to update (required)"),
				"name":         mcp.StringProp("New character name"),
				"race":         mcp.StringProp("New race"),
				"class":        mcp.StringProp("New class"),
				"level":        mcp.IntProp("New level (1-20)"),
				"background":   mcp.StringProp("New background"),
				"alignment":    mcp.StringProp("New alignment"),
				"abilities":    mcp.ObjectProp("New ability scores"),
				"hp":           mcp.ObjectProp("New HP values (current, max, temp)"),
				"ac":           mcp.IntProp("New Armor Class"),
				"speed":        mcp.IntProp("New movement speed"),
				"initiative":   mcp.IntProp("New initiative bonus"),
				"skills":       mcp.ObjectProp("New skill bonuses"),
				"saves":        mcp.ObjectProp("New saving throw bonuses"),
			},
			mcp.Required("character_id"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CharacterID string            `json:"character_id"`
			Name        *string           `json:"name"`
			Race        *string           `json:"race"`
			Class       *string           `json:"class"`
			Level       *int              `json:"level"`
			Background  *string           `json:"background"`
			Alignment   *string           `json:"alignment"`
			Abilities   *models.Abilities `json:"abilities"`
			HP          *models.HP        `json:"hp"`
			AC          *int              `json:"ac"`
			Speed       *int              `json:"speed"`
			Initiative  *int              `json:"initiative"`
			Skills      map[string]int    `json:"skills"`
			Saves       map[string]int    `json:"saves"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		updateReq := &service.UpdateCharacterRequest{
			Name:       input.Name,
			Race:       input.Race,
			Class:      input.Class,
			Level:      input.Level,
			Background: input.Background,
			Alignment:  input.Alignment,
			Abilities:  input.Abilities,
			HP:         input.HP,
			AC:         input.AC,
			Speed:      input.Speed,
			Initiative: input.Initiative,
			Skills:     input.Skills,
			Saves:      input.Saves,
		}

		character, err := t.characterService.UpdateCharacter(ctx, input.CharacterID, updateReq)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"character": character,
			"message":   fmt.Sprintf("Character '%s' updated successfully", character.Name),
		})
	}

	return tool, handler
}

func (t *CharacterTools) listCharactersTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"list_characters",
		"List characters in a campaign with optional filtering by NPC status or player ID.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id": mcp.StringProp("The ID of the campaign (required)"),
				"is_npc":      mcp.BoolProp("Filter by NPC status: true for NPCs only, false for player characters only"),
				"player_id":   mcp.StringProp("Filter by player ID"),
				"npc_type": mcp.PropWithEnum(
					"Filter by NPC type",
					"scripted", "generated",
				),
				"limit":  mcp.IntProp("Maximum number of characters to return (default: 50)"),
				"offset": mcp.IntProp("Number of characters to skip for pagination"),
			},
			mcp.Required("campaign_id"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CampaignID string         `json:"campaign_id"`
			IsNPC      *bool          `json:"is_npc"`
			PlayerID   string         `json:"player_id"`
			NPCType    models.NPCType `json:"npc_type"`
			Limit      int            `json:"limit"`
			Offset     int            `json:"offset"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		listReq := &service.ListCharactersRequest{
			CampaignID: input.CampaignID,
			IsNPC:      input.IsNPC,
			PlayerID:   input.PlayerID,
			NPCType:    input.NPCType,
			Limit:      input.Limit,
			Offset:     input.Offset,
		}

		characters, err := t.characterService.ListCharacters(ctx, listReq)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		// Return simplified summaries
		summaries := make([]map[string]interface{}, len(characters))
		for i, c := range characters {
			summary := map[string]interface{}{
				"id":          c.ID,
				"name":        c.Name,
				"campaign_id": c.CampaignID,
				"is_npc":      c.IsNPC,
				"race":        c.Race,
				"class":       c.Class,
				"level":       c.Level,
				"created_at":  c.CreatedAt,
			}

			// Add HP info if available
			if c.HP != nil {
				summary["hp_current"] = c.HP.Current
				summary["hp_max"] = c.HP.Max
			}

			// Add NPC type if applicable
			if c.IsNPC && c.NPCType != "" {
				summary["npc_type"] = c.NPCType
			}

			// Add player ID for player characters
			if !c.IsNPC && c.PlayerID != "" {
				summary["player_id"] = c.PlayerID
			}

			summaries[i] = summary
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"characters": summaries,
			"count":      len(summaries),
		})
	}

	return tool, handler
}

func (t *CharacterTools) deleteCharacterTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"delete_character",
		"Delete a character by its ID. This action cannot be undone.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"character_id": mcp.StringProp("The unique ID of the character to delete (required)"),
			},
			mcp.Required("character_id"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CharacterID string `json:"character_id"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		err := t.characterService.DeleteCharacter(ctx, input.CharacterID)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"success":      true,
			"message":      fmt.Sprintf("Character %s has been deleted", input.CharacterID),
			"character_id": input.CharacterID,
		})
	}

	return tool, handler
}

// Tool list for external registration
var CharacterToolNames = []string{
	"create_character",
	"get_character",
	"update_character",
	"list_characters",
	"delete_character",
}
