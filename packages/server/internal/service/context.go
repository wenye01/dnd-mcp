// Package service provides business logic layer implementations
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/store"
)

// MessageStoreForContext defines the interface for message data operations needed by context service
type MessageStoreForContext interface {
	ListByCampaign(ctx context.Context, campaignID string, limit int) ([]*models.Message, error)
	CountByCampaign(ctx context.Context, campaignID string) (int, error)
	Create(ctx context.Context, message *models.Message) error
}

// GameStateStoreForContext defines the interface for game state data operations needed by context service
type GameStateStoreForContext interface {
	Get(ctx context.Context, campaignID string) (*models.GameState, error)
}

// CombatStoreForContext defines the interface for combat data operations needed by context service
type CombatStoreForContext interface {
	GetActive(ctx context.Context, campaignID string) (*models.Combat, error)
	Get(ctx context.Context, id string) (*models.Combat, error)
}

// MapStoreForContext defines the interface for map data operations needed by context service
type MapStoreForContext interface {
	Get(ctx context.Context, id string) (*models.Map, error)
}

// ContextService provides context building and compression
type ContextService struct {
	messageStore   MessageStoreForContext
	characterStore CharacterStore
	gameStateStore GameStateStoreForContext
	combatStore    CombatStoreForContext
	mapStore       MapStoreForContext
	defaultWindow  int // 默认滑动窗口大小，20
}

// NewContextService creates a new context service
func NewContextService(
	messageStore MessageStoreForContext,
	characterStore CharacterStore,
	gameStateStore GameStateStoreForContext,
	combatStore CombatStoreForContext,
	mapStore MapStoreForContext,
) *ContextService {
	return &ContextService{
		messageStore:   messageStore,
		characterStore: characterStore,
		gameStateStore: gameStateStore,
		combatStore:    combatStore,
		mapStore:       mapStore,
		defaultWindow:  20,
	}
}

// SetDefaultWindow sets the default window size
func (s *ContextService) SetDefaultWindow(window int) {
	if window > 0 {
		s.defaultWindow = window
	}
}

// GetContext retrieves compressed context for a campaign
func (s *ContextService) GetContext(ctx context.Context, campaignID string, messageLimit int, includeCombat bool) (*models.Context, error) {
	if campaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}

	// Use default window if limit not specified
	if messageLimit <= 0 {
		messageLimit = s.defaultWindow
	}

	// Build game summary
	gameSummary, err := s.buildGameSummary(ctx, campaignID, includeCombat)
	if err != nil {
		return nil, fmt.Errorf("failed to build game summary: %w", err)
	}

	// Get messages with sliding window
	allMessages, err := s.messageStore.ListByCampaign(ctx, campaignID, 0) // Get all messages
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	// Get raw message count
	rawMessageCount := len(allMessages)

	// Apply sliding window
	compressedMessages := s.applySlidingWindow(allMessages, messageLimit)

	// Convert to value type messages
	messages := make([]models.Message, len(compressedMessages))
	for i, msg := range compressedMessages {
		messages[i] = *msg
	}

	// Estimate tokens
	tokenEstimate := s.estimateTokens(messages)

	// Build context
	result := models.NewContext(campaignID)
	result.GameSummary = gameSummary
	result.Messages = messages
	result.RawMessageCount = rawMessageCount
	result.TokenEstimate = tokenEstimate

	return result, nil
}

// GetRawContext retrieves raw context data (full mode)
func (s *ContextService) GetRawContext(ctx context.Context, campaignID string) (*models.GetRawContextResponse, error) {
	if campaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}

	response := &models.GetRawContextResponse{
		CampaignID: campaignID,
	}

	// Get game state
	gameState, err := s.gameStateStore.Get(ctx, campaignID)
	if err == nil {
		response.GameState = gameState
	}

	// Get characters (party members) - only player characters
	isNPCFalse := false
	characters, err := s.characterStore.List(ctx, &store.CharacterFilter{
		CampaignID: campaignID,
		IsNPC:      &isNPCFalse,
	})
	if err == nil && len(characters) > 0 {
		response.Characters = characters
	}

	// Get active combat if any
	combat, err := s.combatStore.GetActive(ctx, campaignID)
	if err == nil && combat != nil {
		response.Combat = combat
	}

	// Get current map if any
	if gameState != nil && gameState.CurrentMapID != "" {
		gameMap, err := s.mapStore.Get(ctx, gameState.CurrentMapID)
		if err == nil {
			response.Map = gameMap
		}
	}

	// Get all messages
	messages, err := s.messageStore.ListByCampaign(ctx, campaignID, 0)
	if err == nil {
		response.Messages = messages
		response.MessageCount = len(messages)
	}

	return response, nil
}

// SaveMessage saves a message
func (s *ContextService) SaveMessage(ctx context.Context, message *models.Message) error {
	if message == nil {
		return NewServiceError(ErrCodeInvalidInput, "message cannot be nil")
	}

	if err := message.Validate(); err != nil {
		return NewServiceError(ErrCodeInvalidInput, fmt.Sprintf("invalid message: %v", err))
	}

	if err := s.messageStore.Create(ctx, message); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	return nil
}

// applySlidingWindow applies sliding window compression
func (s *ContextService) applySlidingWindow(messages []*models.Message, limit int) []*models.Message {
	if len(messages) <= limit {
		return messages
	}

	// Return the most recent 'limit' messages
	start := len(messages) - limit
	return messages[start:]
}

// buildGameSummary builds game state summary
func (s *ContextService) buildGameSummary(ctx context.Context, campaignID string, includeCombat bool) (*models.GameSummary, error) {
	summary := &models.GameSummary{
		Time:     "Unknown time",
		Location: "Unknown location",
		Weather:  "Clear",
		Party:    make([]models.PartyMember, 0),
	}

	// Get game state
	gameState, err := s.gameStateStore.Get(ctx, campaignID)
	if err == nil && gameState != nil {
		// Build time description
		if gameState.GameTime != nil {
			summary.Time = fmt.Sprintf("Year %d, Month %d, Day %d, %02d:%02d (%s)",
				gameState.GameTime.Year,
				gameState.GameTime.Month,
				gameState.GameTime.Day,
				gameState.GameTime.Hour,
				gameState.GameTime.Minute,
				gameState.GameTime.Phase,
			)
		}

		// Build location description
		if gameState.PartyPosition != nil {
			summary.Location = fmt.Sprintf("Position (%d, %d)", gameState.PartyPosition.X, gameState.PartyPosition.Y)
		}
		if gameState.CurrentMapID != "" {
			summary.Location = fmt.Sprintf("%s (Map: %s)", summary.Location, gameState.CurrentMapID)
		}

		// Weather
		if gameState.Weather != "" {
			summary.Weather = gameState.Weather
		}

		// Check if in combat
		summary.InCombat = gameState.IsInCombat()
	}

	// Get party members (player characters only)
	isNPCFalse := false
	characters, err := s.characterStore.List(ctx, &store.CharacterFilter{
		CampaignID: campaignID,
		IsNPC:      &isNPCFalse,
	})
	if err == nil {
		for _, char := range characters {
			member := models.PartyMember{
				ID:    char.ID,
				Name:  char.Name,
				Class: char.Class,
			}

			// Build HP description
			if char.HP != nil {
				percentage := float64(char.HP.Current) / float64(char.HP.Max) * 100
				switch {
				case percentage == 0:
					member.HP = "Unconscious/Dead"
				case percentage < 25:
					member.HP = "Critical"
				case percentage < 50:
					member.HP = "Wounded"
				case percentage < 75:
					member.HP = "Lightly wounded"
				default:
					member.HP = "Healthy"
				}
			} else {
				member.HP = "Unknown"
			}

			summary.Party = append(summary.Party, member)
		}
	}

	// Add combat summary if requested and in combat
	if includeCombat && summary.InCombat {
		combat, err := s.combatStore.GetActive(ctx, campaignID)
		if err == nil && combat != nil {
			// Get character names for participants
			participantNames := make([]string, 0, len(combat.Participants))
			for _, p := range combat.Participants {
				// Try to find character name
				var name string
				characters, _ := s.characterStore.List(ctx, &store.CharacterFilter{
					CampaignID: campaignID,
				})
				for _, char := range characters {
					if char.ID == p.CharacterID {
						name = char.Name
						break
					}
				}
				if name == "" {
					name = p.CharacterID // Use ID if name not found
				}
				participantNames = append(participantNames, name)
			}

			summary.Combat = &models.CombatSummary{
				Round:        combat.Round,
				TurnIndex:    combat.TurnIndex,
				Participants: participantNames,
			}
		}
	}

	return summary, nil
}

// estimateTokens estimates token count (simple approximation: characters / 4)
func (s *ContextService) estimateTokens(messages []models.Message) int {
	totalChars := 0
	for _, msg := range messages {
		// Count content characters
		totalChars += len(msg.Content)
		// Count role characters
		totalChars += len(string(msg.Role))
		// Count tool calls if any
		for _, tc := range msg.ToolCalls {
			totalChars += len(tc.Name)
			totalChars += len(tc.ID)
			// Arguments are harder to estimate, add a fixed amount
			totalChars += 50
		}
	}
	// Simple approximation: 1 token ≈ 4 characters
	return totalChars / 4
}

// FormatContextForLLM formats context for LLM consumption
func (s *ContextService) FormatContextForLLM(ctx *models.Context) string {
	var sb strings.Builder

	sb.WriteString("=== Game State Summary ===\n")
	if ctx.GameSummary != nil {
		sb.WriteString(fmt.Sprintf("Time: %s\n", ctx.GameSummary.Time))
		sb.WriteString(fmt.Sprintf("Location: %s\n", ctx.GameSummary.Location))
		sb.WriteString(fmt.Sprintf("Weather: %s\n", ctx.GameSummary.Weather))
		sb.WriteString(fmt.Sprintf("In Combat: %v\n", ctx.GameSummary.InCombat))

		if len(ctx.GameSummary.Party) > 0 {
			sb.WriteString("\nParty Members:\n")
			for _, member := range ctx.GameSummary.Party {
				sb.WriteString(fmt.Sprintf("  - %s (%s): HP %s\n", member.Name, member.Class, member.HP))
			}
		}

		if ctx.GameSummary.Combat != nil {
			sb.WriteString(fmt.Sprintf("\nCombat Status:\n"))
			sb.WriteString(fmt.Sprintf("  Round: %d\n", ctx.GameSummary.Combat.Round))
			sb.WriteString(fmt.Sprintf("  Turn: %d\n", ctx.GameSummary.Combat.TurnIndex))
			sb.WriteString(fmt.Sprintf("  Participants: %s\n", strings.Join(ctx.GameSummary.Combat.Participants, ", ")))
		}
	}

	sb.WriteString(fmt.Sprintf("\n=== Conversation History (%d messages, compressed from %d total) ===\n", len(ctx.Messages), ctx.RawMessageCount))
	sb.WriteString(fmt.Sprintf("Estimated tokens: %d\n", ctx.TokenEstimate))

	sb.WriteString("\n")
	for _, msg := range ctx.Messages {
		sb.WriteString(fmt.Sprintf("[%s]: %s\n", msg.Role, msg.Content))
	}

	return sb.String()
}
