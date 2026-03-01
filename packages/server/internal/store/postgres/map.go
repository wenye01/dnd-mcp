package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MapStore implements store.MapStore using PostgreSQL
type MapStore struct {
	pool *pgxpool.Pool
}

// NewMapStore creates a new map store
func NewMapStore(client *Client) *MapStore {
	return &MapStore{pool: client.Pool()}
}

// Create creates a new map
func (s *MapStore) Create(ctx context.Context, gameMap *models.Map) error {
	// Generate UUID if not set
	if gameMap.ID == "" {
		gameMap.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	if gameMap.CreatedAt.IsZero() {
		gameMap.CreatedAt = now
	}
	gameMap.UpdatedAt = now

	// Validate map
	if err := gameMap.Validate(); err != nil {
		return fmt.Errorf("invalid map: %w", err)
	}

	// Marshal JSONB fields
	gridJSON, err := json.Marshal(gameMap.Grid)
	if err != nil {
		return fmt.Errorf("failed to marshal grid: %w", err)
	}

	locationsJSON, err := json.Marshal(gameMap.Locations)
	if err != nil {
		return fmt.Errorf("failed to marshal locations: %w", err)
	}

	tokensJSON, err := json.Marshal(gameMap.Tokens)
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}

	// Marshal extended fields
	var imageJSON []byte
	if gameMap.Image != nil {
		imageJSON, err = json.Marshal(gameMap.Image)
		if err != nil {
			return fmt.Errorf("failed to marshal image: %w", err)
		}
	}

	wallsJSON, err := json.Marshal(gameMap.Walls)
	if err != nil {
		return fmt.Errorf("failed to marshal walls: %w", err)
	}

	var importMetaJSON []byte
	if gameMap.ImportMeta != nil {
		importMetaJSON, err = json.Marshal(gameMap.ImportMeta)
		if err != nil {
			return fmt.Errorf("failed to marshal import_meta: %w", err)
		}
	}

	var visualLocationsJSON []byte
	if len(gameMap.VisualLocations) > 0 {
		visualLocationsJSON, err = json.Marshal(gameMap.VisualLocations)
		if err != nil {
			return fmt.Errorf("failed to marshal visual_locations: %w", err)
		}
	}

	query := `
		INSERT INTO maps (id, campaign_id, name, type, mode, grid, locations, tokens, parent_id, image, walls, import_meta, visual_locations, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err = s.pool.Exec(ctx, query,
		gameMap.ID,
		gameMap.CampaignID,
		gameMap.Name,
		string(gameMap.Type),
		string(gameMap.Mode),
		gridJSON,
		locationsJSON,
		tokensJSON,
		nullString(gameMap.ParentID),
		imageJSON,
		wallsJSON,
		importMetaJSON,
		visualLocationsJSON,
		gameMap.CreatedAt,
		gameMap.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create map: %w", err)
	}

	return nil
}

// Get retrieves a map by ID
func (s *MapStore) Get(ctx context.Context, id string) (*models.Map, error) {
	query := `
		SELECT id, campaign_id, name, type, mode, grid, locations, tokens, parent_id, image, walls, import_meta, visual_locations, created_at, updated_at
		FROM maps
		WHERE id = $1
	`

	return s.scanMap(ctx, query, id)
}

// GetByCampaign retrieves maps by campaign ID
func (s *MapStore) GetByCampaign(ctx context.Context, campaignID string) ([]*models.Map, error) {
	query := `
		SELECT id, campaign_id, name, type, mode, grid, locations, tokens, parent_id, image, walls, import_meta, visual_locations, created_at, updated_at
		FROM maps
		WHERE campaign_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.pool.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to list maps: %w", err)
	}
	defer rows.Close()

	var maps []*models.Map
	for rows.Next() {
		gameMap, err := scanMapFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan map: %w", err)
		}
		maps = append(maps, gameMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating maps: %w", err)
	}

	return maps, nil
}

// Update updates a map
func (s *MapStore) Update(ctx context.Context, gameMap *models.Map) error {
	// Validate map
	if err := gameMap.Validate(); err != nil {
		return fmt.Errorf("invalid map: %w", err)
	}

	gameMap.UpdatedAt = time.Now()

	// Marshal JSONB fields
	gridJSON, err := json.Marshal(gameMap.Grid)
	if err != nil {
		return fmt.Errorf("failed to marshal grid: %w", err)
	}

	locationsJSON, err := json.Marshal(gameMap.Locations)
	if err != nil {
		return fmt.Errorf("failed to marshal locations: %w", err)
	}

	tokensJSON, err := json.Marshal(gameMap.Tokens)
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}

	// Marshal extended fields
	var imageJSON []byte
	if gameMap.Image != nil {
		imageJSON, err = json.Marshal(gameMap.Image)
		if err != nil {
			return fmt.Errorf("failed to marshal image: %w", err)
		}
	}

	wallsJSON, err := json.Marshal(gameMap.Walls)
	if err != nil {
		return fmt.Errorf("failed to marshal walls: %w", err)
	}

	var importMetaJSON []byte
	if gameMap.ImportMeta != nil {
		importMetaJSON, err = json.Marshal(gameMap.ImportMeta)
		if err != nil {
			return fmt.Errorf("failed to marshal import_meta: %w", err)
		}
	}

	var visualLocationsJSON []byte
	if len(gameMap.VisualLocations) > 0 {
		visualLocationsJSON, err = json.Marshal(gameMap.VisualLocations)
		if err != nil {
			return fmt.Errorf("failed to marshal visual_locations: %w", err)
		}
	}

	query := `
		UPDATE maps
		SET name = $1, type = $2, mode = $3, grid = $4, locations = $5, tokens = $6, parent_id = $7, image = $8, walls = $9, import_meta = $10, visual_locations = $11, updated_at = $12
		WHERE id = $13
	`

	result, err := s.pool.Exec(ctx, query,
		gameMap.Name,
		string(gameMap.Type),
		string(gameMap.Mode),
		gridJSON,
		locationsJSON,
		tokensJSON,
		nullString(gameMap.ParentID),
		imageJSON,
		wallsJSON,
		importMetaJSON,
		visualLocationsJSON,
		gameMap.UpdatedAt,
		gameMap.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update map: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// Delete deletes a map
func (s *MapStore) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM maps WHERE id = $1`

	result, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete map: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// GetWorldMap retrieves the world map for a campaign
func (s *MapStore) GetWorldMap(ctx context.Context, campaignID string) (*models.Map, error) {
	query := `
		SELECT id, campaign_id, name, type, mode, grid, locations, tokens, parent_id, image, walls, import_meta, visual_locations, created_at, updated_at
		FROM maps
		WHERE campaign_id = $1 AND type = $2 AND parent_id IS NULL
		LIMIT 1
	`

	return s.scanMap(ctx, query, campaignID, string(models.MapTypeWorld))
}

// GetBattleMap retrieves a battle map by ID
func (s *MapStore) GetBattleMap(ctx context.Context, id string) (*models.Map, error) {
	query := `
		SELECT id, campaign_id, name, type, mode, grid, locations, tokens, parent_id, image, walls, import_meta, visual_locations, created_at, updated_at
		FROM maps
		WHERE id = $1 AND type = $2
	`

	return s.scanMap(ctx, query, id, string(models.MapTypeBattle))
}

// GetByParent retrieves battle maps by parent location
func (s *MapStore) GetByParent(ctx context.Context, parentID string) ([]*models.Map, error) {
	query := `
		SELECT id, campaign_id, name, type, mode, grid, locations, tokens, parent_id, image, walls, import_meta, visual_locations, created_at, updated_at
		FROM maps
		WHERE parent_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.pool.Query(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list maps by parent: %w", err)
	}
	defer rows.Close()

	var maps []*models.Map
	for rows.Next() {
		gameMap, err := scanMapFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan map: %w", err)
		}
		maps = append(maps, gameMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating maps: %w", err)
	}

	return maps, nil
}

// scanMap scans a single map using the provided query
func (s *MapStore) scanMap(ctx context.Context, query string, args ...interface{}) (*models.Map, error) {
	row := s.pool.QueryRow(ctx, query, args...)
	return scanMapFromRow(row)
}

// scanMapFromRow scans a map from a row
func scanMapFromRow(row pgx.Row) (*models.Map, error) {
	var (
		id                 string
		campaignID         string
		name               string
		mapType            string
		mapMode            string
		gridJSON           []byte
		locationsJSON      []byte
		tokensJSON         []byte
		parentID           sql.NullString
		imageJSON          []byte
		wallsJSON          []byte
		importMetaJSON     []byte
		visualLocationsJSON []byte
		createdAt          time.Time
		updatedAt          time.Time
	)

	err := row.Scan(
		&id,
		&campaignID,
		&name,
		&mapType,
		&mapMode,
		&gridJSON,
		&locationsJSON,
		&tokensJSON,
		&parentID,
		&imageJSON,
		&wallsJSON,
		&importMetaJSON,
		&visualLocationsJSON,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan map: %w", err)
	}

	// Unmarshal grid
	var grid models.Grid
	if len(gridJSON) > 0 {
		if err := json.Unmarshal(gridJSON, &grid); err != nil {
			return nil, fmt.Errorf("failed to unmarshal grid: %w", err)
		}
	}

	// Unmarshal locations
	var locations []models.Location
	if len(locationsJSON) > 0 {
		if err := json.Unmarshal(locationsJSON, &locations); err != nil {
			return nil, fmt.Errorf("failed to unmarshal locations: %w", err)
		}
	}

	// Unmarshal tokens
	var tokens []models.Token
	if len(tokensJSON) > 0 {
		if err := json.Unmarshal(tokensJSON, &tokens); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tokens: %w", err)
		}
	}

	// Unmarshal image (optional)
	var image *models.MapImage
	if len(imageJSON) > 0 {
		if err := json.Unmarshal(imageJSON, &image); err != nil {
			return nil, fmt.Errorf("failed to unmarshal image: %w", err)
		}
	}

	// Unmarshal walls (optional)
	var walls models.Walls
	if len(wallsJSON) > 0 && string(wallsJSON) != "[]" {
		if err := json.Unmarshal(wallsJSON, &walls); err != nil {
			return nil, fmt.Errorf("failed to unmarshal walls: %w", err)
		}
	}

	// Unmarshal import_meta (optional)
	var importMeta *models.MapImportMeta
	if len(importMetaJSON) > 0 {
		if err := json.Unmarshal(importMetaJSON, &importMeta); err != nil {
			return nil, fmt.Errorf("failed to unmarshal import_meta: %w", err)
		}
	}

	// Unmarshal visual_locations (optional)
	var visualLocations []models.VisualLocation
	if len(visualLocationsJSON) > 0 && string(visualLocationsJSON) != "[]" {
		if err := json.Unmarshal(visualLocationsJSON, &visualLocations); err != nil {
			return nil, fmt.Errorf("failed to unmarshal visual_locations: %w", err)
		}
	}

	gameMap := &models.Map{
		ID:              id,
		CampaignID:      campaignID,
		Name:            name,
		Type:            models.MapType(mapType),
		Mode:            models.MapMode(mapMode),
		Grid:            &grid,
		Locations:       locations,
		Tokens:          tokens,
		Image:           image,
		Walls:           walls,
		ImportMeta:      importMeta,
		VisualLocations: visualLocations,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}

	if parentID.Valid {
		gameMap.ParentID = parentID.String
	}

	return gameMap, nil
}
