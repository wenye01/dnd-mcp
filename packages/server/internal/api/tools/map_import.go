// Package tools provides MCP tool implementations
package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dnd-mcp/server/internal/importer"
	"github.com/dnd-mcp/server/internal/importer/format"
	"github.com/dnd-mcp/server/internal/mcp"
)

// ImportTools provides map import MCP tools
type ImportTools struct {
	importService *importer.ImportService
}

// NewImportTools creates a new ImportTools instance
func NewImportTools(importService *importer.ImportService) *ImportTools {
	return &ImportTools{
		importService: importService,
	}
}

// Register registers all import tools with the registry
func (t *ImportTools) Register(registry *mcp.Registry) {
	registry.MustRegister(t.getImportMapTool())
	registry.MustRegister(t.getImportMapFromModuleTool())
}

// Tool definitions

func (t *ImportTools) getImportMapTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"import_map",
		"Import a map from Universal VTT (.uvtt) or Foundry VTT Scene (.json) format. Automatically detects the format and converts it to the internal map model. The imported map is saved to the campaign and can be used for battle maps or world maps.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id": mcp.StringProp("The ID of the campaign to associate the map with (required)"),
				"data": mcp.StringProp("The map data as a JSON string or Base64-encoded binary data (required). For .uvtt files, provide the JSON content directly. For binary data, Base64 encode it first."),
				"options": mcp.ObjectProp("Import options to customize the import behavior (optional)"),
				"encoding": mcp.PropWithEnum(
					"The encoding of the data parameter: 'json' for JSON string, 'base64' for Base64-encoded data (optional, default: 'json')",
					"json", "base64",
				),
			},
			mcp.Required("campaign_id", "data"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CampaignID string          `json:"campaign_id"`
			Data       string          `json:"data"`
			Options    json.RawMessage `json:"options,omitempty"`
			Encoding   string          `json:"encoding,omitempty"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		// Validate campaign_id
		if input.CampaignID == "" {
			return mcp.NewErrorResponse(fmt.Errorf("campaign_id is required"))
		}

		// Decode the data
		var raw_data []byte
		var err error

		encoding := strings.ToLower(input.Encoding)
		if encoding == "" || encoding == "json" {
			// Treat as JSON string
			raw_data = []byte(input.Data)
		} else if encoding == "base64" {
			// Decode Base64
			raw_data, err = base64.StdEncoding.DecodeString(input.Data)
			if err != nil {
				return mcp.NewErrorResponse(fmt.Errorf("failed to decode base64 data: %w", err))
			}
		} else {
			return mcp.NewErrorResponse(fmt.Errorf("unsupported encoding: %s (supported: json, base64)", encoding))
		}

		// Parse options
		opts := format.ImportOptions{
			CampaignID:    input.CampaignID,
			Format:        format.FormatAuto,
			ImportTokens:  true,
			ImportWalls:   true,
			ImportLights:  true,
		}

		if len(input.Options) > 0 {
			var options struct {
				Format            string  `json:"format"`
				Name              string  `json:"name"`
				ImportTokens      *bool   `json:"import_tokens"`
				ImportWalls       *bool   `json:"import_walls"`
				ImportLights      *bool   `json:"import_lights"`
				Scale             float64 `json:"scale"`
				OffsetX           int     `json:"offset_x"`
				OffsetY           int     `json:"offset_y"`
				OverwriteExisting bool    `json:"overwrite_existing"`
			}

			if err := json.Unmarshal(input.Options, &options); err != nil {
				return mcp.NewErrorResponse(fmt.Errorf("failed to parse options: %w", err))
			}

			// Apply options
			if options.Format != "" {
				opts.Format = format.ImportFormat(options.Format)
			}
			if options.Name != "" {
				opts.Name = options.Name
			}
			if options.ImportTokens != nil {
				opts.ImportTokens = *options.ImportTokens
			}
			if options.ImportWalls != nil {
				opts.ImportWalls = *options.ImportWalls
			}
			if options.ImportLights != nil {
				opts.ImportLights = *options.ImportLights
			}
			if options.Scale > 0 {
				opts.Scale = options.Scale
			}
			if options.OffsetX > 0 {
				opts.OffsetX = options.OffsetX
			}
			if options.OffsetY > 0 {
				opts.OffsetY = options.OffsetY
			}
			opts.OverwriteExisting = options.OverwriteExisting
		}

		// Import the map
		result, err := t.importService.ImportAndSave(ctx, input.CampaignID, raw_data, opts)
		if err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("failed to import map: %w", err))
		}

		// Build response
		response := map[string]interface{}{
			"message": fmt.Sprintf("Successfully imported map '%s'", result.Map.Name),
			"map": map[string]interface{}{
				"id":   result.Map.ID,
				"name": result.Map.Name,
				"type": result.Map.Type,
				"mode": string(result.Map.Mode),
			},
			"import_meta": map[string]interface{}{
				"source_format": string(result.Meta.SourceFormat),
				"source_size":   result.Meta.SourceSize,
				"import_time":   result.Meta.ImportTime,
			},
		}

		// Add grid information
		if result.Map.Grid != nil {
			response["map"].(map[string]interface{})["width"] = result.Map.Grid.Width
			response["map"].(map[string]interface{})["height"] = result.Map.Grid.Height
			response["map"].(map[string]interface{})["cell_size"] = result.Map.Grid.CellSize
		}

		// Add image information
		if result.Map.Image != nil {
			response["map"].(map[string]interface{})["image"] = map[string]interface{}{
				"url":    result.Map.Image.URL,
				"width":  result.Map.Image.Width,
				"height": result.Map.Image.Height,
			}
		}

		// Add counts
		wallsCount := len(result.Map.Walls)
		tokensCount := len(result.Map.Tokens)

		response["map"].(map[string]interface{})["walls_count"] = wallsCount
		response["map"].(map[string]interface{})["tokens_count"] = tokensCount

		// Add warnings if any
		if len(result.Warnings) > 0 {
			response["import_meta"].(map[string]interface{})["warnings"] = result.Warnings
		}

		// Add skipped info if any
		if result.Skipped != nil {
			response["import_meta"].(map[string]interface{})["skipped"] = result.Skipped
		}

		return mcp.NewJSONResponse(response)
	}

	return tool, handler
}

// Tool list for external registration
var ImportToolNames = []string{
	"import_map",
	"import_map_from_module",
}

func (t *ImportTools) getImportMapFromModuleTool() (mcp.Tool, mcp.ToolHandler) {
	tool := mcp.NewTool(
		"import_map_from_module",
		"Import maps from a Foundry VTT module directory. Reads the module.json manifest and imports scenes from the compendium packs. Can import all scenes or a specific scene by name.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"campaign_id":  mcp.StringProp("The ID of the campaign to associate the imported maps with (required)"),
				"module_path":  mcp.StringProp("The file system path to the FVTT module directory containing module.json (required)"),
				"scene_name":   mcp.StringProp("The name of a specific scene to import (optional). If not provided, all scenes will be imported."),
				"import_tokens": mcp.BoolProp("Whether to import tokens from the scenes (optional, default: true)"),
				"import_walls":  mcp.BoolProp("Whether to import walls from the scenes (optional, default: true)"),
			},
			mcp.Required("campaign_id", "module_path"),
		),
	)

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var input struct {
			CampaignID    string `json:"campaign_id"`
			ModulePath    string `json:"module_path"`
			SceneName     string `json:"scene_name,omitempty"`
			ImportTokens  *bool  `json:"import_tokens,omitempty"`
			ImportWalls   *bool  `json:"import_walls,omitempty"`
		}

		if err := json.Unmarshal(req.Arguments, &input); err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("invalid arguments: %w", err))
		}

		// Validate required fields
		if input.CampaignID == "" {
			return mcp.NewErrorResponse(fmt.Errorf("campaign_id is required"))
		}
		if input.ModulePath == "" {
			return mcp.NewErrorResponse(fmt.Errorf("module_path is required"))
		}

		// Build import options
		opts := format.ImportOptions{
			CampaignID:   input.CampaignID,
			Format:       format.FormatFVTTModule,
			ImportTokens: true,
			ImportWalls:  true,
			ImportLights: true,
		}

		if input.ImportTokens != nil {
			opts.ImportTokens = *input.ImportTokens
		}
		if input.ImportWalls != nil {
			opts.ImportWalls = *input.ImportWalls
		}

		// Import from module
		result, err := t.importService.ImportFromModule(ctx, input.CampaignID, input.ModulePath, input.SceneName, opts)
		if err != nil {
			return mcp.NewErrorResponse(fmt.Errorf("failed to import from module: %w", err))
		}

		// Build response
		mapsData := make([]map[string]interface{}, 0, len(result.Maps))
		for _, gameMap := range result.Maps {
			mapInfo := map[string]interface{}{
				"id":           gameMap.ID,
				"name":         gameMap.Name,
				"type":         gameMap.Type,
				"mode":         string(gameMap.Mode),
				"walls_count":  len(gameMap.Walls),
				"tokens_count": len(gameMap.Tokens),
			}
			if gameMap.Grid != nil {
				mapInfo["width"] = gameMap.Grid.Width
				mapInfo["height"] = gameMap.Grid.Height
				mapInfo["cell_size"] = gameMap.Grid.CellSize
			}
			if gameMap.Image != nil {
				mapInfo["image_url"] = gameMap.Image.URL
			}
			mapsData = append(mapsData, mapInfo)
		}

		response := map[string]interface{}{
			"message":       fmt.Sprintf("Successfully imported %d map(s) from module", len(result.Maps)),
			"imported_count": len(result.Maps),
			"maps":          mapsData,
		}

		// Add module info if available
		if result.ModuleInfo != nil {
			response["module_info"] = map[string]interface{}{
				"name":        result.ModuleInfo.Name,
				"title":       result.ModuleInfo.Title,
				"description": result.ModuleInfo.Description,
				"version":     result.ModuleInfo.Version,
				"author":      result.ModuleInfo.Author,
				"system":      result.ModuleInfo.System,
				"scene_count": result.ModuleInfo.SceneCount,
			}
		}

		// Add warnings if any
		if len(result.Warnings) > 0 {
			response["warnings"] = result.Warnings
		}

		// Add import meta if available
		if result.Meta != nil {
			response["import_meta"] = map[string]interface{}{
				"source_format": string(result.Meta.SourceFormat),
				"import_time":   result.Meta.ImportTime,
			}
		}

		return mcp.NewJSONResponse(response)
	}

	return tool, handler
}
