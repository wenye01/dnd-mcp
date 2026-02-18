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

// CampaignTools provides campaign-related MCP tools
type CampaignTools struct {
	campaignService *service.CampaignService
}

// NewCampaignTools creates a new CampaignTools instance
func NewCampaignTools(campaignService *service.CampaignService) *CampaignTools {
	return &CampaignTools{
		campaignService: campaignService,
	}
}

// Register registers all campaign tools with the registry
func (t *CampaignTools) Register(registry *mcp.Registry) {
	registry.MustRegister(t.createCampaignTool())
	registry.MustRegister(t.getCampaignTool())
	registry.MustRegister(t.listCampaignsTool())
	registry.MustRegister(t.deleteCampaignTool())
	registry.MustRegister(t.getCampaignSummaryTool())
}

// Tool definitions

func (t *CampaignTools) createCampaignTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"create_campaign",
		"Create a new D&D campaign with specified settings. Returns the created campaign with its ID.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"name":        mcp.StringProp("The name of the campaign (required)"),
				"description": mcp.StringProp("A description of the campaign setting and theme"),
				"dm_id":       mcp.StringProp("The DM (Dungeon Master) user ID (required)"),
				"settings": mcp.ObjectProp("Campaign settings including max_players, start_level, ruleset, house_rules, context_window"),
			},
			mcp.Required("name", "dm_id"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			Name        string                          `json:"name"`
			Description string                          `json:"description"`
			DMID        string                          `json:"dm_id"`
			Settings    *service.CampaignSettingsInput `json:"settings"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		createReq := &service.CreateCampaignRequest{
			Name:        input.Name,
			Description: input.Description,
			DMID:        input.DMID,
			Settings:    input.Settings,
		}

		campaign, err := t.campaignService.CreateCampaign(ctx, createReq)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"campaign": campaign,
			"message":  fmt.Sprintf("Campaign '%s' created successfully with ID: %s", campaign.Name, campaign.ID),
		})
	}

	return tool, handler
}

func (t *CampaignTools) getCampaignTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"get_campaign",
		"Get detailed information about a specific campaign by its ID.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id": mcp.StringProp("The unique ID of the campaign (required)"),
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

		campaign, err := t.campaignService.GetCampaign(ctx, input.CampaignID)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"campaign": campaign,
		})
	}

	return tool, handler
}

func (t *CampaignTools) listCampaignsTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"list_campaigns",
		"List all campaigns with optional filtering by status or DM. Returns a list of campaign summaries.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"status": mcp.PropWithEnum(
					"Filter by campaign status",
					"active", "paused", "finished", "archived",
				),
				"dm_id":           mcp.StringProp("Filter by DM user ID"),
				"include_deleted": mcp.BoolProp("Include soft-deleted campaigns (default: false)"),
				"limit":           mcp.IntProp("Maximum number of campaigns to return (default: 50)"),
				"offset":          mcp.IntProp("Number of campaigns to skip for pagination"),
			},
			[]string{},
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			Status         string `json:"status"`
			DMID           string `json:"dm_id"`
			IncludeDeleted bool   `json:"include_deleted"`
			Limit          int    `json:"limit"`
			Offset         int    `json:"offset"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		listReq := &service.ListCampaignsRequest{
			Status:         models.CampaignStatus(input.Status),
			DMID:           input.DMID,
			IncludeDeleted: input.IncludeDeleted,
			Limit:          input.Limit,
			Offset:         input.Offset,
		}

		campaigns, err := t.campaignService.ListCampaigns(ctx, listReq)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		// Return simplified summaries
		summaries := make([]map[string]interface{}, len(campaigns))
		for i, c := range campaigns {
			summaries[i] = map[string]interface{}{
				"id":          c.ID,
				"name":        c.Name,
				"description": c.Description,
				"dm_id":       c.DMID,
				"status":      c.Status,
				"created_at":  c.CreatedAt,
			}
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"campaigns": summaries,
			"count":     len(summaries),
		})
	}

	return tool, handler
}

func (t *CampaignTools) deleteCampaignTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"delete_campaign",
		"Soft delete a campaign by its ID. The campaign will be archived and no longer accessible.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id": mcp.StringProp("The unique ID of the campaign to delete (required)"),
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

		err := t.campaignService.DeleteCampaign(ctx, input.CampaignID)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"success":    true,
			"message":    fmt.Sprintf("Campaign %s has been deleted", input.CampaignID),
			"campaign_id": input.CampaignID,
		})
	}

	return tool, handler
}

func (t *CampaignTools) getCampaignSummaryTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"get_campaign_summary",
		"Get a summary of the campaign state for LLM context. Includes campaign info, game time, current location, and party status.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id": mcp.StringProp("The unique ID of the campaign (required)"),
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

		summary, err := t.campaignService.GetCampaignSummary(ctx, input.CampaignID)
		if err != nil {
			return mcp.NewErrorResponse(err)
		}

		return mcp.NewJSONResponse(map[string]interface{}{
			"summary": summary,
		})
	}

	return tool, handler
}

// Tool list for external registration
var CampaignToolNames = []string{
	"create_campaign",
	"get_campaign",
	"list_campaigns",
	"delete_campaign",
	"get_campaign_summary",
}
