// Package parser provides format-specific parsers for map import
package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dnd-mcp/server/internal/importer/format"
)

const (
	// MaxScanTokenSize is the maximum size for scanning JSON lines
	// Foundry VTT scenes can be very large (100KB+), so we need a larger buffer
	MaxScanTokenSize = 1024 * 1024 // 1MB
)

// FVTTCompendium represents an FVTT Compendium entry
type FVTTCompendium struct {
	ID   string          `json:"_id"`
	Name string          `json:"name"`
	Type string          `json:"type"` // "Scene" for scenes
	Data json.RawMessage `json:"-"`    // Raw JSON data
}

// FVTTModuleManifest represents the module.json file
type FVTTModuleManifest struct {
	ID           string                  `json:"id"`
	Title        string                  `json:"title"`
	Description  string                  `json:"description"`
	Version      string                  `json:"version"`
	Authors      []FVTTModuleAuthor      `json:"authors"`
	Compatibility FVTTModuleCompatibility `json:"compatibility"`
	Packs        []FVTTModulePack        `json:"packs"`
}

// FVTTModuleAuthor represents a module author
type FVTTModuleAuthor struct {
	Name string `json:"name"`
}

// FVTTModuleCompatibility represents version compatibility
type FVTTModuleCompatibility struct {
	Minimum  string `json:"minimum"`
	Verified string `json:"verified"`
}

// FVTTModulePack represents a compendium pack definition
type FVTTModulePack struct {
	Name string `json:"name"`
	Label string `json:"label"`
	Path string `json:"path"`
	Type string `json:"type"` // "Scene", "Actor", "JournalEntry", etc.
}

// NDJSONParser implements ModuleParser for FVTT Compendium (NDJSON) format
type NDJSONParser struct {
	modulePath  string
	dbPath      string
	manifest    *FVTTModuleManifest
	sceneIndex  map[string]int // name/ID -> line number
	sceneCount  int
}

// NewNDJSONParser creates a new NDJSON parser for FVTT Compendium
func NewNDJSONParser() *NDJSONParser {
	return &NDJSONParser{
		sceneIndex: make(map[string]int),
	}
}

// Open opens a module at the given path
func (p *NDJSONParser) Open(modulePath string) error {
	// Validate module path
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		return fmt.Errorf("module path does not exist: %s", modulePath)
	}

	p.modulePath = modulePath

	// Load module.json manifest
	manifestPath := filepath.Join(modulePath, "module.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read module.json: %w", err)
	}

	var manifest FVTTModuleManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return fmt.Errorf("failed to parse module.json: %w", err)
	}

	p.manifest = &manifest

	// Find the Scene pack
	var scenePack *FVTTModulePack
	for i := range manifest.Packs {
		if manifest.Packs[i].Type == "Scene" {
			scenePack = &manifest.Packs[i]
			break
		}
	}

	if scenePack == nil {
		return fmt.Errorf("no Scene pack found in module")
	}

	// Resolve pack path (may be relative or absolute)
	p.dbPath = scenePack.Path
	if !filepath.IsAbs(p.dbPath) {
		p.dbPath = filepath.Join(modulePath, p.dbPath)
	}

	// Check if pack file exists
	if _, err := os.Stat(p.dbPath); os.IsNotExist(err) {
		return fmt.Errorf("scene pack file does not exist: %s", p.dbPath)
	}

	// Build scene index
	if err := p.buildSceneIndex(); err != nil {
		return fmt.Errorf("failed to build scene index: %w", err)
	}

	return nil
}

// Close closes the module and releases resources
func (p *NDJSONParser) Close() error {
	p.modulePath = ""
	p.dbPath = ""
	p.manifest = nil
	p.sceneIndex = make(map[string]int)
	p.sceneCount = 0
	return nil
}

// ListScenes returns a list of scene names/IDs in the module
func (p *NDJSONParser) ListScenes() ([]string, error) {
	if p.dbPath == "" {
		return nil, fmt.Errorf("parser not opened, call Open first")
	}

	file, err := os.Open(p.dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open pack file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Increase buffer size for large JSON lines
	buf := make([]byte, 0, MaxScanTokenSize)
	scanner.Buffer(buf, MaxScanTokenSize)

	var scenes []string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var scene struct {
			ID   string `json:"_id"`
			Name string `json:"name"`
		}

		if err := json.Unmarshal([]byte(line), &scene); err != nil {
			continue // Skip invalid lines
		}

		// Skip temporary entities
		if strings.HasPrefix(scene.Name, "#[CF_") {
			continue
		}

		displayName := scene.Name
		if displayName == "" {
			displayName = scene.ID
		}
		scenes = append(scenes, displayName)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading pack file: %w", err)
	}

	return scenes, nil
}

// GetScene retrieves a specific scene by name or ID
func (p *NDJSONParser) GetScene(nameOrID string) (*format.ParseResult, error) {
	if p.dbPath == "" {
		return nil, fmt.Errorf("parser not opened, call Open first")
	}

	file, err := os.Open(p.dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open pack file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Increase buffer size for large JSON lines
	buf := make([]byte, 0, MaxScanTokenSize)
	scanner.Buffer(buf, MaxScanTokenSize)

	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		if strings.TrimSpace(line) == "" {
			continue
		}

		// Parse to check ID/Name
		var scene struct {
			ID   string          `json:"_id"`
			Name string          `json:"name"`
			Data json.RawMessage `json:"-"`
		}

		if err := json.Unmarshal([]byte(line), &scene); err != nil {
			continue
		}

		// Check if this is the scene we're looking for
		if scene.ID == nameOrID || scene.Name == nameOrID {
			// Parse full scene data
			var sceneData map[string]interface{}
			if err := json.Unmarshal([]byte(line), &sceneData); err != nil {
				return nil, fmt.Errorf("failed to parse scene data: %w", err)
			}

			return &format.ParseResult{
				Data:   sceneData,
				Format: format.FormatFVTTScene,
			}, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading pack file: %w", err)
	}

	return nil, fmt.Errorf("scene not found: %s", nameOrID)
}

// GetModuleInfo returns information about the module
func (p *NDJSONParser) GetModuleInfo() (*format.ModuleInfo, error) {
	if p.manifest == nil {
		return nil, fmt.Errorf("parser not opened, call Open first")
	}

	info := &format.ModuleInfo{
		Name:        p.manifest.ID,
		Title:       p.manifest.Title,
		Description: p.manifest.Description,
		Version:     p.manifest.Version,
		System:      "dnd5e", // Default assumption for this project
		SceneCount:  p.sceneCount,
	}

	if len(p.manifest.Authors) > 0 {
		info.Author = p.manifest.Authors[0].Name
	}

	return info, nil
}

// buildSceneIndex builds an index of scenes for faster lookup
func (p *NDJSONParser) buildSceneIndex() error {
	file, err := os.Open(p.dbPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Increase buffer size for large JSON lines
	buf := make([]byte, 0, MaxScanTokenSize)
	scanner.Buffer(buf, MaxScanTokenSize)

	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		if strings.TrimSpace(line) == "" {
			continue
		}

		var scene struct {
			ID   string `json:"_id"`
			Name string `json:"name"`
		}

		if err := json.Unmarshal([]byte(line), &scene); err != nil {
			continue
		}

		// Skip temporary entities
		if !strings.HasPrefix(scene.Name, "#[CF_") {
			p.sceneIndex[scene.ID] = lineNum
			if scene.Name != "" {
				p.sceneIndex[scene.Name] = lineNum
			}
			p.sceneCount++
		}
	}

	return scanner.Err()
}

// GetSceneByID retrieves a scene by its exact ID
func (p *NDJSONParser) GetSceneByID(id string) (*format.ParseResult, error) {
	return p.GetScene(id)
}

// GetSceneByName retrieves a scene by its exact name
func (p *NDJSONParser) GetSceneByName(name string) (*format.ParseResult, error) {
	return p.GetScene(name)
}

// GetAllScenes retrieves all scenes from the module
func (p *NDJSONParser) GetAllScenes() ([]*format.ParseResult, error) {
	if p.dbPath == "" {
		return nil, fmt.Errorf("parser not opened, call Open first")
	}

	file, err := os.Open(p.dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open pack file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Increase buffer size for large JSON lines
	buf := make([]byte, 0, MaxScanTokenSize)
	scanner.Buffer(buf, MaxScanTokenSize)

	var scenes []*format.ParseResult

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var scene struct {
			ID   string `json:"_id"`
			Name string `json:"name"`
		}

		if err := json.Unmarshal([]byte(line), &scene); err != nil {
			continue
		}

		// Skip temporary entities
		if strings.HasPrefix(scene.Name, "#[CF_") {
			continue
		}

		// Parse full scene data
		var sceneData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &sceneData); err != nil {
			continue
		}

		scenes = append(scenes, &format.ParseResult{
			Data:   sceneData,
			Format: format.FormatFVTTScene,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading pack file: %w", err)
	}

	return scenes, nil
}
