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

// ContextTools provides context-related MCP tools
type ContextTools struct {
	contextService *service.ContextService
}

// NewContextTools creates a new ContextTools instance
func NewContextTools(contextService *service.ContextService) *ContextTools {
	return &ContextTools{
		contextService: contextService,
	}
}

// Register registers all context tools with the registry
func (t *ContextTools) Register(registry *mcp.Registry) {
	registry.MustRegister(t.getContextTool())
	registry.MustRegister(t.getRawContextTool())
	registry.MustRegister(t.saveMessageTool())
}

// getContextTool implements the get_context tool
func (t *ContextTools) getContextTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"get_context",
		"Get compressed context for a campaign (simplified mode). Returns game summary and recent messages with sliding window compression.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id":    mcp.StringProp("The campaign ID to get context for (required)"),
				"message_limit":  mcp.IntProp("Maximum number of recent messages to include (default: 20)"),
				"include_combat": mcp.BoolProp("Whether to include combat details (default: true)"),
			},
			mcp.Required("campaign_id"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CampaignID    string `json:"campaign_id"`
			MessageLimit  int    `json:"message_limit"`
			IncludeCombat bool   `json:"include_combat"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		// Set default for include_combat
		if input.MessageLimit == 0 {
			input.MessageLimit = 20
		}

		contextData, err := t.contextService.GetContext(ctx, input.CampaignID, input.MessageLimit, input.IncludeCombat)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"context": contextData,
		})
	}

	return tool, handler
}

// getRawContextTool implements the get_raw_context tool
func (t *ContextTools) getRawContextTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"get_raw_context",
		"Get complete raw context data for a campaign (full mode). Returns all campaign data including characters, game state, combat, map, and messages.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id": mcp.StringProp("The campaign ID to get raw context for (required)"),
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

		rawContext, err := t.contextService.GetRawContext(ctx, input.CampaignID)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		return mcp.NewJSONResponse(rawContext)
	}

	return tool, handler
}

// saveMessageTool implements the save_message tool
func (t *ContextTools) saveMessageTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"save_message",
		"Save a dialogue message to the campaign history. Supports user, assistant, and system messages.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id": mcp.StringProp("The campaign ID to save the message to (required)"),
				"role":        mcp.StringProp("The message role: 'user', 'assistant', or 'system' (required)"),
				"content":     mcp.StringProp("The message content (required)"),
				"player_id":   mcp.StringProp("The player ID (required for user messages)"),
				"tool_calls":  mcp.ObjectProp("Array of tool calls made by the assistant (optional, for assistant messages)"),
			},
			mcp.Required("campaign_id", "role", "content"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CampaignID string                  `json:"campaign_id"`
			Role       string                  `json:"role"`
			Content    string                  `json:"content"`
			PlayerID   string                  `json:"player_id"`
			ToolCalls  []map[string]interface{} `json:"tool_calls"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		// Create message
		message := models.NewMessage(input.CampaignID, models.MessageRole(input.Role), input.Content)
		message.PlayerID = input.PlayerID

		// Convert tool calls if provided
		if len(input.ToolCalls) > 0 {
			toolCalls := make([]models.ToolCall, len(input.ToolCalls))
			for i, tc := range input.ToolCalls {
				toolCalls[i] = models.ToolCall{
					ID:        getStringField(tc, "id"),
					Name:      getStringField(tc, "name"),
					Arguments: getMapField(tc, "arguments"),
				}
			}
			message.ToolCalls = toolCalls
		}

		if err := t.contextService.SaveMessage(ctx, message); err != nil {
			return mcp.NewErrorResponse(err)
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"message": message,
		})
	}

	return tool, handler
}

// getStringField safely extracts a string field from a map
func getStringField(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getMapField safely extracts a map field from a map
func getMapField(m map[string]interface{}, key string) map[string]interface{} {
	if val, ok := m[key]; ok {
		if mp, ok := val.(map[string]interface{}); ok {
			return mp
		}
	}
	return nil
}

// Tool list for external registration
var ContextToolNames = []string{
	"get_context",
	"get_raw_context",
	"save_message",
}
