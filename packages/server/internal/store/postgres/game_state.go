package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GameStateStore implements store.GameStateStore using PostgreSQL
type GameStateStore struct {
	pool *pgxpool.Pool
}

// NewGameStateStore creates a new game state store
func NewGameStateStore(client *Client) *GameStateStore {
	return &GameStateStore{pool: client.Pool()}
}

// Create creates a new game state
func (s *GameStateStore) Create(ctx context.Context, gameState *models.GameState) error {
	// Set ID to campaign ID if not set
	if gameState.ID == "" {
		gameState.ID = gameState.CampaignID
	}

	// Set timestamp
	gameState.UpdatedAt = time.Now()

	// Marshal JSONB fields
	gameTimeJSON, err := json.Marshal(gameState.GameTime)
	if err != nil {
		return fmt.Errorf("failed to marshal game_time: %w", err)
	}

	var partyPositionJSON []byte
	if gameState.PartyPosition != nil {
		partyPositionJSON, err = json.Marshal(gameState.PartyPosition)
		if err != nil {
			return fmt.Errorf("failed to marshal party_position: %w", err)
		}
	}

	var playerMarkerJSON []byte
	if gameState.PlayerMarker != nil {
		playerMarkerJSON, err = json.Marshal(gameState.PlayerMarker)
		if err != nil {
			return fmt.Errorf("failed to marshal player_marker: %w", err)
		}
	}

	query := `
		INSERT INTO game_states (id, campaign_id, game_time, party_position, current_map_id, current_map_type, weather, active_combat_id, player_marker, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = s.pool.Exec(ctx, query,
		gameState.ID,
		gameState.CampaignID,
		gameTimeJSON,
		nullJSON(partyPositionJSON),
		nullString(gameState.CurrentMapID),
		string(gameState.CurrentMapType),
		nullString(gameState.Weather),
		nullString(gameState.ActiveCombatID),
		nullJSON(playerMarkerJSON),
		gameState.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create game state: %w", err)
	}

	return nil
}

// Get retrieves game state by campaign ID
func (s *GameStateStore) Get(ctx context.Context, campaignID string) (*models.GameState, error) {
	query := `
		SELECT id, campaign_id, game_time, party_position, current_map_id, current_map_type, weather, active_combat_id, player_marker, updated_at
		FROM game_states
		WHERE campaign_id = $1
	`

	return s.scanGameState(ctx, query, campaignID)
}

// GetByID retrieves game state by its own ID
func (s *GameStateStore) GetByID(ctx context.Context, id string) (*models.GameState, error) {
	query := `
		SELECT id, campaign_id, game_time, party_position, current_map_id, current_map_type, weather, active_combat_id, player_marker, updated_at
		FROM game_states
		WHERE id = $1
	`

	return s.scanGameState(ctx, query, id)
}

// Update updates a game state
func (s *GameStateStore) Update(ctx context.Context, gameState *models.GameState) error {
	gameState.UpdatedAt = time.Now()

	// Marshal JSONB fields
	gameTimeJSON, err := json.Marshal(gameState.GameTime)
	if err != nil {
		return fmt.Errorf("failed to marshal game_time: %w", err)
	}

	var partyPositionJSON []byte
	if gameState.PartyPosition != nil {
		partyPositionJSON, err = json.Marshal(gameState.PartyPosition)
		if err != nil {
			return fmt.Errorf("failed to marshal party_position: %w", err)
		}
	}

	var playerMarkerJSON []byte
	if gameState.PlayerMarker != nil {
		playerMarkerJSON, err = json.Marshal(gameState.PlayerMarker)
		if err != nil {
			return fmt.Errorf("failed to marshal player_marker: %w", err)
		}
	}

	query := `
		UPDATE game_states
		SET game_time = $1, party_position = $2, current_map_id = $3, current_map_type = $4, weather = $5, active_combat_id = $6, player_marker = $7, updated_at = $8
		WHERE campaign_id = $9
	`

	result, err := s.pool.Exec(ctx, query,
		gameTimeJSON,
		nullJSON(partyPositionJSON),
		nullString(gameState.CurrentMapID),
		string(gameState.CurrentMapType),
		nullString(gameState.Weather),
		nullString(gameState.ActiveCombatID),
		nullJSON(playerMarkerJSON),
		gameState.UpdatedAt,
		gameState.CampaignID,
	)

	if err != nil {
		return fmt.Errorf("failed to update game state: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// Delete deletes a game state
func (s *GameStateStore) Delete(ctx context.Context, campaignID string) error {
	query := `DELETE FROM game_states WHERE campaign_id = $1`

	result, err := s.pool.Exec(ctx, query, campaignID)
	if err != nil {
		return fmt.Errorf("failed to delete game state: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// scanGameState scans a single game state using the provided query
func (s *GameStateStore) scanGameState(ctx context.Context, query string, args ...interface{}) (*models.GameState, error) {
	row := s.pool.QueryRow(ctx, query, args...)
	return scanGameStateFromRow(row)
}

// scanGameStateFromRow scans a game state from a row
func scanGameStateFromRow(row pgx.Row) (*models.GameState, error) {
	var (
		id                string
		campaignID        string
		gameTimeJSON      []byte
		partyPositionJSON []byte
		currentMapID      sql.NullString
		currentMapType    string
		weather           sql.NullString
		activeCombatID    sql.NullString
		playerMarkerJSON  []byte
		updatedAt         time.Time
	)

	err := row.Scan(
		&id,
		&campaignID,
		&gameTimeJSON,
		&partyPositionJSON,
		&currentMapID,
		&currentMapType,
		&weather,
		&activeCombatID,
		&playerMarkerJSON,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan game state: %w", err)
	}

	// Unmarshal game time
	var gameTime models.GameTime
	if len(gameTimeJSON) > 0 {
		if err := json.Unmarshal(gameTimeJSON, &gameTime); err != nil {
			return nil, fmt.Errorf("failed to unmarshal game_time: %w", err)
		}
	}

	// Unmarshal party position
	var partyPosition *models.Position
	if len(partyPositionJSON) > 0 {
		var pos models.Position
		if err := json.Unmarshal(partyPositionJSON, &pos); err != nil {
			return nil, fmt.Errorf("failed to unmarshal party_position: %w", err)
		}
		partyPosition = &pos
	}

	// Unmarshal player marker
	var playerMarker *models.PlayerMarker
	if len(playerMarkerJSON) > 0 {
		var marker models.PlayerMarker
		if err := json.Unmarshal(playerMarkerJSON, &marker); err != nil {
			return nil, fmt.Errorf("failed to unmarshal player_marker: %w", err)
		}
		playerMarker = &marker
	}

	gameState := &models.GameState{
		ID:             id,
		CampaignID:     campaignID,
		GameTime:       &gameTime,
		PartyPosition:  partyPosition,
		CurrentMapID:   currentMapID.String,
		CurrentMapType: models.MapType(currentMapType),
		Weather:        weather.String,
		ActiveCombatID: activeCombatID.String,
		PlayerMarker:   playerMarker,
		UpdatedAt:      updatedAt,
	}

	return gameState, nil
}

// nullJSON returns nil for empty JSON, otherwise returns the bytes
func nullJSON(b []byte) interface{} {
	if len(b) == 0 {
		return nil
	}
	return b
}
