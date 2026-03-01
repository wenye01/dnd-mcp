// Package tools provides MCP tool implementations
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dnd-mcp/server/internal/mcp"
	"github.com/dnd-mcp/server/internal/service"
)

// CombatTools provides combat-related MCP tools
type CombatTools struct {
	combatService *service.CombatService
}

// NewCombatTools creates a new CombatTools instance
func NewCombatTools(combatService *service.CombatService) *CombatTools {
	return &CombatTools{
		combatService: combatService,
	}
}

// Register registers all combat tools with the registry
func (t *CombatTools) Register(registry *mcp.Registry) {
	registry.MustRegister(t.startCombatTool())
	registry.MustRegister(t.getCombatStateTool())
	registry.MustRegister(t.attackTool())
	registry.MustRegister(t.castSpellTool())
	registry.MustRegister(t.endTurnTool())
	registry.MustRegister(t.endCombatTool())
}

// startCombatTool implements the start_combat tool
func (t *CombatTools) startCombatTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"start_combat",
		"Start a new combat encounter. Rolls initiative for all participants, sorts them by initiative order, and returns the combat state.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id": mcp.StringProp("The ID of the campaign (required)"),
				"participant_ids": mcp.ArrayProp("List of character IDs participating in combat (required)"),
				"map_id": mcp.StringProp("Optional map ID for the combat encounter"),
			},
			mcp.Required("campaign_id", "participant_ids"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CampaignID     string   `json:"campaign_id"`
			ParticipantIDs []string `json:"participant_ids"`
			MapID          string   `json:"map_id"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		startReq := &service.StartCombatRequest{
			CampaignID:     input.CampaignID,
			ParticipantIDs: input.ParticipantIDs,
			MapID:          input.MapID,
		}

		combat, err := t.combatService.StartCombat(ctx, startReq)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		// Build participant summary
		participants := make([]map[string]interface{}, len(combat.Participants))
		for i, p := range combat.Participants {
			participants[i] = map[string]interface{}{
				"character_id": p.CharacterID,
				"initiative":   p.Initiative,
				"has_acted":    p.HasActed,
			}
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"combat": map[string]interface{}{
				"id":             combat.ID,
				"campaign_id":    combat.CampaignID,
				"round":          combat.Round,
				"turn_index":     combat.TurnIndex,
				"status":         combat.Status,
				"participants":   participants,
				"active":         combat.IsActive(),
			},
			"message": fmt.Sprintf("Combat started with %d participants. Round 1 begins.", len(combat.Participants)),
		})
	}

	return tool, handler
}

// getCombatStateTool implements the get_combat_state tool
func (t *CombatTools) getCombatStateTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"get_combat_state",
		"Get the current state of a combat encounter, including round, turn, participants order, and combat log.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"combat_id": mcp.StringProp("The unique ID of the combat encounter (required)"),
			},
			mcp.Required("combat_id"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CombatID string `json:"combat_id"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		combat, err := t.combatService.GetCombatState(ctx, input.CombatID)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		// Build participant summary
		participants := make([]map[string]interface{}, len(combat.Participants))
		for i, p := range combat.Participants {
			participants[i] = map[string]interface{}{
				"character_id": p.CharacterID,
				"initiative":   p.Initiative,
				"has_acted":    p.HasActed,
				"conditions":   p.Conditions,
			}
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"combat": map[string]interface{}{
				"id":             combat.ID,
				"campaign_id":    combat.CampaignID,
				"round":          combat.Round,
				"turn_index":     combat.TurnIndex,
				"status":         combat.Status,
				"participants":   participants,
				"active":         combat.IsActive(),
				"log":            combat.Log,
			},
		})
	}

	return tool, handler
}

// attackTool implements the attack tool
func (t *CombatTools) attackTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"attack",
		"Execute an attack in combat. Must be the attacker's turn. Rolls attack dice, determines hit/miss, calculates damage, and updates target HP.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"combat_id":    mcp.StringProp("The ID of the combat encounter (required)"),
				"attacker_id":  mcp.StringProp("The ID of the attacking character (required)"),
				"target_id":    mcp.StringProp("The ID of the target character (required)"),
				"advantage":    mcp.BoolProp("Roll with advantage (roll 2d20, take higher)"),
				"disadvantage": mcp.BoolProp("Roll with disadvantage (roll 2d20, take lower)"),
			},
			mcp.Required("combat_id", "attacker_id", "target_id"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CombatID     string `json:"combat_id"`
			AttackerID   string `json:"attacker_id"`
			TargetID     string `json:"target_id"`
			Advantage    bool   `json:"advantage"`
			Disadvantage bool   `json:"disadvantage"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		attackReq := &service.AttackRequest{
			CombatID:     input.CombatID,
			AttackerID:   input.AttackerID,
			TargetID:     input.TargetID,
			Advantage:    input.Advantage,
			Disadvantage: input.Disadvantage,
		}

		resp, err := t.combatService.Attack(ctx, attackReq)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		// Build result message
		result := resp.Result
		var message string
		if result.Hit {
			if result.Crit {
				message = fmt.Sprintf("CRITICAL HIT! Attack roll %d vs AC %d, dealing %d %s damage!",
					result.AttackRoll.Total, result.TargetAC, result.Damage, result.DamageType)
			} else {
				message = fmt.Sprintf("Hit! Attack roll %d vs AC %d, dealing %d %s damage.",
					result.AttackRoll.Total, result.TargetAC, result.Damage, result.DamageType)
			}
			if result.TargetHP != nil {
				message += fmt.Sprintf(" Target HP: %d/%d", result.TargetHP.Current, result.TargetHP.Max)
			}
			if result.TargetDown {
				message += " Target is down!"
			}
		} else {
			if result.AttackRoll.IsFumble() {
				message = fmt.Sprintf("FUMBLE! Natural 1 - Attack roll %d vs AC %d, miss!",
					result.AttackRoll.Total, result.TargetAC)
			} else {
				message = fmt.Sprintf("Miss! Attack roll %d vs AC %d.", result.AttackRoll.Total, result.TargetAC)
			}
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"result": map[string]interface{}{
				"hit":         result.Hit,
				"crit":        result.Crit,
				"attack_roll": result.AttackRoll.Total,
				"target_ac":   result.TargetAC,
				"damage":      result.Damage,
				"damage_type": result.DamageType,
				"target_hp":   result.TargetHP,
				"target_down": result.TargetDown,
			},
			"target_dead": resp.TargetDead,
			"message":     message,
		})
	}

	return tool, handler
}

// castSpellTool implements the cast_spell tool
func (t *CombatTools) castSpellTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"cast_spell",
		"Cast a spell in combat. Must be the caster's turn. Applies damage or healing to targets and updates their HP.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"combat_id":   mcp.StringProp("The ID of the combat encounter (required)"),
				"caster_id":   mcp.StringProp("The ID of the caster character (required)"),
				"spell_id":    mcp.StringProp("The ID of the spell to cast"),
				"spell_name":  mcp.StringProp("The name of the spell to cast"),
				"target_ids":  mcp.ArrayProp("List of target character IDs (required)"),
				"level":       mcp.IntProp("The spell level to cast at (for upcasting)"),
				"damage":      mcp.StringProp("Damage formula (e.g., '2d6', '3d8')"),
				"damage_type": mcp.StringProp("Type of damage (e.g., 'fire', 'cold', 'necrotic')"),
				"is_healing":  mcp.BoolProp("Whether this is a healing spell"),
			},
			mcp.Required("combat_id", "caster_id", "target_ids"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CombatID   string   `json:"combat_id"`
			CasterID   string   `json:"caster_id"`
			SpellID    string   `json:"spell_id"`
			SpellName  string   `json:"spell_name"`
			TargetIDs  []string `json:"target_ids"`
			Level      int      `json:"level"`
			Damage     string   `json:"damage"`
			DamageType string   `json:"damage_type"`
			IsHealing  bool     `json:"is_healing"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		castReq := &service.CastSpellRequest{
			CombatID:   input.CombatID,
			CasterID:   input.CasterID,
			SpellID:    input.SpellID,
			SpellName:  input.SpellName,
			TargetIDs:  input.TargetIDs,
			Level:      input.Level,
			Damage:     input.Damage,
			DamageType: input.DamageType,
			IsHealing:  input.IsHealing,
		}

		resp, err := t.combatService.CastSpell(ctx, castReq)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		// Build result message
		result := resp.Result
		spellName := result.SpellName
		if spellName == "" {
			spellName = result.SpellID
		}

		var message string
		if result.IsHealing {
			message = fmt.Sprintf("%s healed %d points", spellName, result.Damage)
		} else {
			message = fmt.Sprintf("%s dealt %d %s damage", spellName, result.Damage, result.DamageType)
		}

		// Build target results
		targetResults := make([]map[string]interface{}, len(result.Results))
		for i, tr := range result.Results {
			targetResults[i] = map[string]interface{}{
				"target_id": tr.TargetID,
				"hit":       tr.Hit,
				"damage":    tr.Damage,
				"target_hp": tr.TargetHP,
			}
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"result": map[string]interface{}{
				"spell_id":     result.SpellID,
				"spell_name":   result.SpellName,
				"level":        result.Level,
				"caster_id":    result.CasterID,
				"damage":       result.Damage,
				"damage_type":  result.DamageType,
				"is_healing":   result.IsHealing,
				"target_results": targetResults,
			},
			"message": message,
		})
	}

	return tool, handler
}

// endTurnTool implements the end_turn tool
func (t *CombatTools) endTurnTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"end_turn",
		"End the current character's turn and advance to the next participant. If this is the last participant, a new round begins. Also processes condition durations.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"combat_id": mcp.StringProp("The unique ID of the combat encounter (required)"),
			},
			mcp.Required("combat_id"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CombatID string `json:"combat_id"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		resp, err := t.combatService.AdvanceTurn(ctx, input.CombatID)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		// Build current participant info
		currentParticipant := resp.Combat.GetCurrentParticipant()
		var currentTurnInfo map[string]interface{}
		if currentParticipant != nil {
			currentTurnInfo = map[string]interface{}{
				"character_id": currentParticipant.CharacterID,
				"initiative":   currentParticipant.Initiative,
				"has_acted":    currentParticipant.HasActed,
				"conditions":   currentParticipant.Conditions,
			}
		}

		// Build message
		var message string
		if resp.NewRound {
			message = fmt.Sprintf("Round %d begins. ", resp.Combat.Round)
		}
		if resp.CurrentTurnName != "" {
			message += fmt.Sprintf("It is now %s's turn.", resp.CurrentTurnName)
		} else if currentParticipant != nil {
			message += fmt.Sprintf("It is now character %s's turn.", currentParticipant.CharacterID)
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"combat": map[string]interface{}{
				"id":             resp.Combat.ID,
				"round":          resp.Combat.Round,
				"turn_index":     resp.Combat.TurnIndex,
				"new_round":      resp.NewRound,
				"current_turn":   currentTurnInfo,
				"participant_count": len(resp.Combat.Participants),
			},
			"current_turn_name": resp.CurrentTurnName,
			"message":           message,
		})
	}

	return tool, handler
}

// Tool list for external registration
var CombatToolNames = []string{
	"start_combat",
	"get_combat_state",
	"attack",
	"cast_spell",
	"end_turn",
	"end_combat",
}

// endCombatTool implements the end_combat tool
func (t *CombatTools) endCombatTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"end_combat",
		"End a combat encounter and generate a combat summary report with statistics for each participant including damage dealt, damage taken, healing received, and survival status.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"combat_id": mcp.StringProp("The unique ID of the combat encounter to end (required)"),
			},
			mcp.Required("combat_id"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CombatID string `json:"combat_id"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		resp, err := t.combatService.EndCombatWithSummary(ctx, input.CombatID)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		// Build participant summaries
		participants := make([]map[string]interface{}, len(resp.Summary.Participants))
		for i, p := range resp.Summary.Participants {
			participants[i] = map[string]interface{}{
				"character_id":     p.CharacterID,
				"character_name":   p.CharacterName,
				"damage_dealt":     p.DamageDealt,
				"damage_taken":     p.DamageTaken,
				"healing_received": p.HealingReceived,
				"kills":            p.Kills,
				"final_hp":         p.FinalHP,
				"max_hp":           p.MaxHP,
				"survived":         p.Survived,
			}
		}

		// Build summary
		summary := map[string]interface{}{
			"total_rounds":  resp.Summary.TotalRounds,
			"total_turns":   resp.Summary.TotalTurns,
			"duration":      resp.Summary.Duration,
			"participants":  participants,
		}

		// Build message
		survivors := 0
		casualties := 0
		for _, p := range resp.Summary.Participants {
			if p.Survived {
				survivors++
			} else {
				casualties++
			}
		}

		message := fmt.Sprintf("Combat ended after %d rounds. Survivors: %d, Casualties: %d",
			resp.Summary.TotalRounds, survivors, casualties)

		return mcp.NewJSONResponse(map[string]interface{}{
			"combat": map[string]interface{}{
				"id":          resp.Combat.ID,
				"campaign_id": resp.Combat.CampaignID,
				"status":      resp.Combat.Status,
				"round":       resp.Combat.Round,
				"started_at":  resp.Combat.StartedAt,
				"ended_at":    resp.Combat.EndedAt,
			},
			"summary": summary,
			"message": message,
		})
	}

	return tool, handler
}
