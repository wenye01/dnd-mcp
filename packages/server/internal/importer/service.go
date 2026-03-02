package importer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dnd-mcp/server/internal/importer/format"
	"github.com/dnd-mcp/server/internal/models"
)

// MapStoreForImport defines the interface for map storage operations needed by import
type MapStoreForImport interface {
	Create(ctx context.Context, gameMap *models.Map) error
	Get(ctx context.Context, id string) (*models.Map, error)
	Update(ctx context.Context, gameMap *models.Map) error
}

// ImportService provides map import functionality
type ImportService struct {
	mu         sync.RWMutex
	detector   FormatDetector
	parsers    map[format.ImportFormat]Parser
	converters map[format.ImportFormat]Converter
	validator  Validator
	mapStore   MapStoreForImport
}

// NewImportService creates a new import service
func NewImportService(mapStore MapStoreForImport) *ImportService {
	return &ImportService{
		detector:   NewDefaultFormatDetector(),
		parsers:    make(map[format.ImportFormat]Parser),
		converters: make(map[format.ImportFormat]Converter),
		validator:  NewDefaultValidator(),
		mapStore:   mapStore,
	}
}

// RegisterParser registers a parser for a specific format
func (s *ImportService) RegisterParser(p Parser) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.parsers[p.Format()] = p
}

// RegisterConverter registers a converter for a specific format
func (s *ImportService) RegisterConverter(c Converter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.converters[c.Format()] = c
}

// RegisterConverterForFormat registers a converter for a specific format
// This allows the same converter to be registered for multiple formats
func (s *ImportService) RegisterConverterForFormat(c Converter, f format.ImportFormat) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.converters[f] = c
}

// SetValidator sets the validator to use
func (s *ImportService) SetValidator(v Validator) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.validator = v
}

// SetDetector sets the format detector to use
func (s *ImportService) SetDetector(d FormatDetector) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.detector = d
}

// Import imports a map from raw data
func (s *ImportService) Import(ctx context.Context, data []byte, opts format.ImportOptions) (*format.ImportResult, error) {
	// Apply default options
	if opts.Format == format.FormatAuto || opts.Format == "" {
		opts.Format = format.FormatAuto
	}

	// Detect format if auto
	detectedFormat := opts.Format
	if opts.Format == format.FormatAuto {
		detectedFormat = s.detector.Detect(data)
		if detectedFormat == format.FormatAuto {
			return nil, fmt.Errorf("unable to detect format")
		}
	}

	// Get parser
	s.mu.RLock()
	parser, ok := s.parsers[detectedFormat]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("no parser registered for format: %s", detectedFormat)
	}

	// Parse the data
	parseResult, err := parser.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	// Get converter
	s.mu.RLock()
	converter, ok := s.converters[detectedFormat]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("no converter registered for format: %s", detectedFormat)
	}

	// Convert to Map model
	gameMap, err := converter.Convert(parseResult.Data, opts)
	if err != nil {
		return nil, fmt.Errorf("convert error: %w", err)
	}

	// Override name if provided
	if opts.Name != "" {
		gameMap.Name = opts.Name
	}

	// Validate the map
	s.mu.RLock()
	validator := s.validator
	s.mu.RUnlock()

	if validator != nil {
		warnings, err := validator.ValidateWithWarnings(gameMap)
		if err != nil {
			return nil, fmt.Errorf("validation error: %w", err)
		}

		// Add validation warnings to result
		parseResult.Warnings = append(parseResult.Warnings, warnings...)
	}

	// Build result
	result := &format.ImportResult{
		Map:      gameMap,
		Warnings: parseResult.Warnings,
		Meta: &format.ImportMeta{
			SourceFormat: detectedFormat,
			SourceSize:   int64(len(data)),
			ImportTime:   time.Now(),
			SourceName:   gameMap.Name,
		},
	}

	// Check for image
	if gameMap.Image != nil {
		result.Meta.HasImage = true
		result.Meta.ImageDimensions = &format.ImageDimensions{
			Width:  gameMap.Image.Width,
			Height: gameMap.Image.Height,
		}
	}

	return result, nil
}

// ImportAndSave imports a map and saves it to the store
func (s *ImportService) ImportAndSave(ctx context.Context, campaignID string, data []byte, opts format.ImportOptions) (*format.ImportResult, error) {
	// Import the map
	result, err := s.Import(ctx, data, opts)
	if err != nil {
		return nil, err
	}

	// Set campaign ID
	result.Map.CampaignID = campaignID

	// Check if map already exists
	existing, err := s.mapStore.Get(ctx, result.Map.ID)
	if err == nil && existing != nil {
		if !opts.OverwriteExisting {
			return nil, fmt.Errorf("map with ID %s already exists", result.Map.ID)
		}
		// Update existing
		if err := s.mapStore.Update(ctx, result.Map); err != nil {
			return nil, fmt.Errorf("failed to update map: %w", err)
		}
	} else {
		// Create new
		if err := s.mapStore.Create(ctx, result.Map); err != nil {
			return nil, fmt.Errorf("failed to save map: %w", err)
		}
	}

	return result, nil
}

// ImportFromModule imports maps from an FVTT module
func (s *ImportService) ImportFromModule(ctx context.Context, campaignID string, modulePath string, sceneName string, opts format.ImportOptions) (*format.ImportFromModuleResult, error) {
	// Create NDJSON parser for FVTT modules
	// We use a local import to avoid circular dependency
	type moduleParser interface {
		Open(modulePath string) error
		Close() error
		ListScenes() ([]string, error)
		GetScene(nameOrID string) (*format.ParseResult, error)
		GetAllScenes() ([]*format.ParseResult, error)
		GetModuleInfo() (*ModuleInfo, error)
	}

	// Import the NDJSON parser
	parser := getNDJSONParser()
	if parser == nil {
		return nil, fmt.Errorf("NDJSON parser not available")
	}

	// Open the module
	if err := parser.Open(modulePath); err != nil {
		return nil, fmt.Errorf("failed to open module: %w", err)
	}
	defer parser.Close()

	// Get module info
	moduleInfo, err := parser.GetModuleInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get module info: %w", err)
	}

	// Get the FVTT Scene converter
	s.mu.RLock()
	converter, ok := s.converters[format.FormatFVTTScene]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("no converter registered for FVTT Scene format")
	}

	result := &format.ImportFromModuleResult{
		Maps:       make([]*models.Map, 0),
		Warnings:   make([]string, 0),
		ModuleInfo: &format.ModuleInfo{
			Name:        moduleInfo.Name,
			Title:       moduleInfo.Title,
			Description: moduleInfo.Description,
			Version:     moduleInfo.Version,
			Author:      moduleInfo.Author,
			System:      moduleInfo.System,
			SceneCount:  moduleInfo.SceneCount,
		},
		Meta: &format.ImportMeta{
			SourceFormat:    format.FormatFVTTModule,
			ImportTimestamp: time.Now(),
			SourceName:      moduleInfo.Title,
		},
	}

	// Get scenes to import
	var scenes []*format.ParseResult
	if sceneName != "" {
		// Import specific scene
		scene, err := parser.GetScene(sceneName)
		if err != nil {
			return nil, fmt.Errorf("failed to get scene '%s': %w", sceneName, err)
		}
		scenes = append(scenes, scene)
	} else {
		// Import all scenes
		scenes, err = parser.GetAllScenes()
		if err != nil {
			return nil, fmt.Errorf("failed to get scenes: %w", err)
		}
	}

	// Convert each scene to Map model
	for _, sceneResult := range scenes {
		// Convert to Map model
		gameMap, err := converter.Convert(sceneResult.Data, opts)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to convert scene: %v", err))
			continue
		}

		// Set campaign ID
		gameMap.CampaignID = campaignID

		// Validate the map
		s.mu.RLock()
		validator := s.validator
		s.mu.RUnlock()

		if validator != nil {
			warnings, err := validator.ValidateWithWarnings(gameMap)
			if err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Validation failed for '%s': %v", gameMap.Name, err))
				continue
			}
			result.Warnings = append(result.Warnings, warnings...)
		}

		// Save to store
		if s.mapStore != nil {
			// Check if map already exists
			existing, err := s.mapStore.Get(ctx, gameMap.ID)
			if err == nil && existing != nil {
				if !opts.OverwriteExisting {
					result.Warnings = append(result.Warnings, fmt.Sprintf("Map with ID %s already exists, skipping", gameMap.ID))
					continue
				}
				// Update existing
				if err := s.mapStore.Update(ctx, gameMap); err != nil {
					result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to update map '%s': %v", gameMap.Name, err))
					continue
				}
			} else {
				// Create new
				if err := s.mapStore.Create(ctx, gameMap); err != nil {
					result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to save map '%s': %v", gameMap.Name, err))
					continue
				}
			}
		}

		result.Maps = append(result.Maps, gameMap)
	}

	return result, nil
}

// DefaultFormatDetector is the default format detector
type DefaultFormatDetector struct{}

// NewDefaultFormatDetector creates a new default format detector
func NewDefaultFormatDetector() *DefaultFormatDetector {
	return &DefaultFormatDetector{}
}

// Detect attempts to detect the format of the given data
func (d *DefaultFormatDetector) Detect(data []byte) format.ImportFormat {
	// Check for UVTT format (contains "format" or "resolution" field)
	if len(data) > 0 && data[0] == '{' {
		// JSON format - check for UVTT vs FVTT Scene
		dataStr := string(data)

		// UVTT typically has "resolution" field
		if containsField(dataStr, "resolution") && containsField(dataStr, "format") {
			return format.FormatUVTT
		}

		// FVTT Scene typically has "_id" and "grid" fields
		if containsField(dataStr, "_id") && containsField(dataStr, "grid") {
			return format.FormatFVTTScene
		}
	}

	// Could be LevelDB (binary format starting with specific bytes)
	// LevelDB files start with specific magic bytes
	if len(data) > 4 && data[0] == 0x00 && data[1] == 0x00 {
		return format.FormatFVTTModule
	}

	return format.FormatAuto
}

// containsField checks if a JSON string contains a field name
func containsField(jsonStr, field string) bool {
	// Simple check - could be improved with proper JSON parsing
	for i := 0; i < len(jsonStr)-len(field)-2; i++ {
		if jsonStr[i] == '"' && jsonStr[i+len(field)+1] == '"' {
			if jsonStr[i+1:i+len(field)+1] == field {
				return true
			}
		}
	}
	return false
}

// DefaultValidator is the default map validator
type DefaultValidator struct{}

// NewDefaultValidator creates a new default validator
func NewDefaultValidator() *DefaultValidator {
	return &DefaultValidator{}
}

// Validate validates a map
func (v *DefaultValidator) Validate(gameMap *models.Map) error {
	return gameMap.Validate()
}

// ValidateWithWarnings validates a map and returns warnings
func (v *DefaultValidator) ValidateWithWarnings(gameMap *models.Map) ([]string, error) {
	var warnings []string

	// Basic validation
	if err := gameMap.Validate(); err != nil {
		return warnings, err
	}

	// Check for potential issues
	if gameMap.Grid.Width <= 0 || gameMap.Grid.Height <= 0 {
		warnings = append(warnings, "map has zero or negative dimensions")
	}

	if gameMap.Mode == models.MapModeImage && gameMap.Image == nil {
		warnings = append(warnings, "image mode map has no image data")
	}

	if len(gameMap.Tokens) > 100 {
		warnings = append(warnings, fmt.Sprintf("large number of tokens (%d) may impact performance", len(gameMap.Tokens)))
	}

	return warnings, nil
}

// ndjsonParser is an interface for the NDJSON parser
// This is defined locally to avoid import cycles
type ndjsonParser interface {
	Open(modulePath string) error
	Close() error
	ListScenes() ([]string, error)
	GetScene(nameOrID string) (*format.ParseResult, error)
	GetAllScenes() ([]*format.ParseResult, error)
	GetModuleInfo() (*format.ModuleInfo, error)
}

// getNDJSONParser creates and returns an NDJSON parser instance
// This function is implemented in a separate file to avoid import cycles
var getNDJSONParser = func() ndjsonParser {
	// This will be set by the init function in ndjson_bridge.go
	return nil
}
