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

// CombatStore implements store.CombatStore using PostgreSQL
type CombatStore struct {
	pool *pgxpool.Pool
}

// NewCombatStore creates a new combat store
func NewCombatStore(client *Client) *CombatStore {
	return &CombatStore{pool: client.Pool()}
}

// Create creates a new combat
func (s *CombatStore) Create(ctx context.Context, combat *models.Combat) error {
	// Generate UUID if not set
	if combat.ID == "" {
		combat.ID = uuid.New().String()
	}

	// Set started_at if not set
	if combat.StartedAt.IsZero() {
		combat.StartedAt = time.Now()
	}

	// Marshal JSONB fields
	participantsJSON, err := json.Marshal(combat.Participants)
	if err != nil {
		return fmt.Errorf("failed to marshal participants: %w", err)
	}

	logJSON, err := json.Marshal(combat.Log)
	if err != nil {
		return fmt.Errorf("failed to marshal log: %w", err)
	}

	query := `
		INSERT INTO combats (id, campaign_id, status, round, turn_index, participants, map_id, log, started_at, ended_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = s.pool.Exec(ctx, query,
		combat.ID,
		combat.CampaignID,
		string(combat.Status),
		combat.Round,
		combat.TurnIndex,
		participantsJSON,
		nullString(combat.MapID),
		logJSON,
		combat.StartedAt,
		nullTime(combat.EndedAt),
	)

	if err != nil {
		return fmt.Errorf("failed to create combat: %w", err)
	}

	return nil
}

// Get retrieves a combat by ID
func (s *CombatStore) Get(ctx context.Context, id string) (*models.Combat, error) {
	query := `
		SELECT id, campaign_id, status, round, turn_index, participants, map_id, log, started_at, ended_at
		FROM combats
		WHERE id = $1
	`

	return s.scanCombat(ctx, query, id)
}

// GetByCampaign retrieves combats by campaign ID
func (s *CombatStore) GetByCampaign(ctx context.Context, campaignID string) ([]*models.Combat, error) {
	query := `
		SELECT id, campaign_id, status, round, turn_index, participants, map_id, log, started_at, ended_at
		FROM combats
		WHERE campaign_id = $1
		ORDER BY started_at DESC
	`

	rows, err := s.pool.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get combats by campaign: %w", err)
	}
	defer rows.Close()

	var combats []*models.Combat
	for rows.Next() {
		combat, err := scanCombatFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan combat: %w", err)
		}
		combats = append(combats, combat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating combats: %w", err)
	}

	return combats, nil
}

// GetActive retrieves the active combat for a campaign
func (s *CombatStore) GetActive(ctx context.Context, campaignID string) (*models.Combat, error) {
	query := `
		SELECT id, campaign_id, status, round, turn_index, participants, map_id, log, started_at, ended_at
		FROM combats
		WHERE campaign_id = $1 AND status = 'active'
		ORDER BY started_at DESC
		LIMIT 1
	`

	return s.scanCombat(ctx, query, campaignID)
}

// Update updates a combat
func (s *CombatStore) Update(ctx context.Context, combat *models.Combat) error {
	// Marshal JSONB fields
	participantsJSON, err := json.Marshal(combat.Participants)
	if err != nil {
		return fmt.Errorf("failed to marshal participants: %w", err)
	}

	logJSON, err := json.Marshal(combat.Log)
	if err != nil {
		return fmt.Errorf("failed to marshal log: %w", err)
	}

	query := `
		UPDATE combats
		SET status = $1, round = $2, turn_index = $3, participants = $4, map_id = $5, log = $6, ended_at = $7
		WHERE id = $8
	`

	result, err := s.pool.Exec(ctx, query,
		string(combat.Status),
		combat.Round,
		combat.TurnIndex,
		participantsJSON,
		nullString(combat.MapID),
		logJSON,
		nullTime(combat.EndedAt),
		combat.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update combat: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// Delete deletes a combat
func (s *CombatStore) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM combats WHERE id = $1`

	result, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete combat: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// scanCombat scans a single combat using the provided query
func (s *CombatStore) scanCombat(ctx context.Context, query string, args ...interface{}) (*models.Combat, error) {
	row := s.pool.QueryRow(ctx, query, args...)
	return scanCombatFromRow(row)
}

// scanCombatFromRow scans a combat from a single row
func scanCombatFromRow(row pgx.Row) (*models.Combat, error) {
	var (
		id              string
		campaignID      string
		status          string
		round           int
		turnIndex       int
		participantsJSON []byte
		mapID           sql.NullString
		logJSON         []byte
		startedAt       time.Time
		endedAt         sql.NullTime
	)

	err := row.Scan(
		&id,
		&campaignID,
		&status,
		&round,
		&turnIndex,
		&participantsJSON,
		&mapID,
		&logJSON,
		&startedAt,
		&endedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan combat: %w", err)
	}

	combat := &models.Combat{
		ID:         id,
		CampaignID: campaignID,
		Status:     models.CombatStatus(status),
		Round:      round,
		TurnIndex:  turnIndex,
		MapID:      mapID.String,
		StartedAt:  startedAt,
	}

	if endedAt.Valid {
		combat.EndedAt = &endedAt.Time
	}

	// Unmarshal participants
	if len(participantsJSON) > 0 {
		if err := json.Unmarshal(participantsJSON, &combat.Participants); err != nil {
			return nil, fmt.Errorf("failed to unmarshal participants: %w", err)
		}
	}

	// Unmarshal log
	if len(logJSON) > 0 {
		if err := json.Unmarshal(logJSON, &combat.Log); err != nil {
			return nil, fmt.Errorf("failed to unmarshal log: %w", err)
		}
	}

	return combat, nil
}

// scanCombatFromRows scans a combat from rows
func scanCombatFromRows(rows pgx.Rows) (*models.Combat, error) {
	var (
		id              string
		campaignID      string
		status          string
		round           int
		turnIndex       int
		participantsJSON []byte
		mapID           sql.NullString
		logJSON         []byte
		startedAt       time.Time
		endedAt         sql.NullTime
	)

	err := rows.Scan(
		&id,
		&campaignID,
		&status,
		&round,
		&turnIndex,
		&participantsJSON,
		&mapID,
		&logJSON,
		&startedAt,
		&endedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan combat: %w", err)
	}

	combat := &models.Combat{
		ID:         id,
		CampaignID: campaignID,
		Status:     models.CombatStatus(status),
		Round:      round,
		TurnIndex:  turnIndex,
		MapID:      mapID.String,
		StartedAt:  startedAt,
	}

	if endedAt.Valid {
		combat.EndedAt = &endedAt.Time
	}

	// Unmarshal participants
	if len(participantsJSON) > 0 {
		if err := json.Unmarshal(participantsJSON, &combat.Participants); err != nil {
			return nil, fmt.Errorf("failed to unmarshal participants: %w", err)
		}
	}

	// Unmarshal log
	if len(logJSON) > 0 {
		if err := json.Unmarshal(logJSON, &combat.Log); err != nil {
			return nil, fmt.Errorf("failed to unmarshal log: %w", err)
		}
	}

	return combat, nil
}

// nullTime returns a sql.NullTime for a time pointer
func nullTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}
