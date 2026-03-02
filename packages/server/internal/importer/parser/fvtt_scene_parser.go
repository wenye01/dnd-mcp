// Package parser provides Foundry VTT Scene JSON parser
// 规则参考: Foundry VTT v10 Scene Data Format
package parser

import (
	"encoding/json"
	"fmt"

	"github.com/dnd-mcp/server/internal/importer/format"
)

// FVTTSceneParser parses Foundry VTT Scene JSON files
type FVTTSceneParser struct{}

// NewFVTTSceneParser creates a new FVTT Scene parser
func NewFVTTSceneParser() *FVTTSceneParser {
	return &FVTTSceneParser{}
}

// CanParse checks if the data is a valid FVTT Scene JSON
func (p *FVTTSceneParser) CanParse(data []byte) bool {
	// Quick check for common FVTT Scene field names
	dataStr := string(data)

	// Check for required fields
	hasID := containsKey(dataStr, "_id")
	hasName := containsKey(dataStr, "name")
	hasWalls := containsKey(dataStr, "walls")
	hasTokens := containsKey(dataStr, "tokens")

	// Must have at least _id and name, and one of walls/tokens
	return hasID && hasName && (hasWalls || hasTokens)
}

// Parse parses FVTT Scene JSON data
func (p *FVTTSceneParser) Parse(data []byte) (*format.ParseResult, error) {
	var scene format.FVTTScene

	// Parse JSON
	if err := json.Unmarshal(data, &scene); err != nil {
		return nil, format.NewParseError(format.FormatFVTTScene, "failed to parse JSON", err)
	}

	// Validate required fields
	if scene.ID == "" {
		return nil, format.NewParseError(format.FormatFVTTScene, "missing required field: _id", nil)
	}
	if scene.Name == "" {
		return nil, format.NewParseError(format.FormatFVTTScene, "missing required field: name", nil)
	}

	// Validate dimensions
	if scene.Width <= 0 {
		return nil, format.NewParseError(format.FormatFVTTScene, "invalid width: must be positive", nil)
	}
	if scene.Height <= 0 {
		return nil, format.NewParseError(format.FormatFVTTScene, "invalid height: must be positive", nil)
	}

	// Validate grid configuration
	if scene.HasGrid() && scene.Grid <= 0 {
		return nil, format.NewParseError(format.FormatFVTTScene, "invalid grid size: must be positive when grid is enabled", nil)
	}

	// Collect warnings
	warnings := p.validateScene(&scene)

	// Create parse result
	result := &format.ParseResult{
		Data:     &scene,
		Warnings: warnings,
		Format:   format.FormatFVTTScene,
	}

	return result, nil
}

// Format returns the format this parser handles
func (p *FVTTSceneParser) Format() format.ImportFormat {
	return format.FormatFVTTScene
}

// validateScene performs additional validation and collects warnings
func (p *FVTTSceneParser) validateScene(scene *format.FVTTScene) []string {
	warnings := []string{}

	// Check for unsupported grid types
	if scene.IsHexGrid() {
		warnings = append(warnings, "hexagonal grids are not fully supported, converting to square grid")
	}

	// Check for large scenes
	if scene.Width > 4000 || scene.Height > 4000 {
		warnings = append(warnings, fmt.Sprintf("scene is very large (%dx%d pixels), may impact performance", scene.Width, scene.Height))
	}

	// Check for tiles without proper images
	for i, tile := range scene.Tiles {
		if tile.Image == "" {
			warnings = append(warnings, fmt.Sprintf("tile %d (%s) has no image path", i, tile.ID))
		}
	}

	// Check for tokens with missing actor links but with actorId set
	for _, token := range scene.Tokens {
		if token.ActorID != "" && !token.ActorLink {
			// This is just a copy of the actor, not linked
			// No warning needed, just informational
		}
	}

	// Check for invalid wall coordinates
	for i, wall := range scene.Walls {
		if len(wall.C) < 2 {
			warnings = append(warnings, fmt.Sprintf("wall %d (%s) has invalid coordinates", i, wall.ID))
		}
	}

	// Check for unset grid distance with grid enabled
	if scene.HasGrid() && scene.GridDistance == 0 {
		warnings = append(warnings, "grid distance not set, defaulting to 5 feet per square")
	}

	return warnings
}

// containsKey checks if a JSON key exists in the data
func containsKey(data, key string) bool {
	// Simple string search for the key
	// This is a quick check before full JSON parsing
	// We look for the key followed by a colon or space
	for i := 0; i < len(data)-len(key)-2; i++ {
		if data[i:i+1] == "\"" && data[i+1:i+1+len(key)] == key {
			// Check if followed by quote (JSON key pattern)
			j := i + 1 + len(key)
			if j < len(data) && data[j:j+1] == "\"" {
				return true
			}
		}
	}
	return false
}
