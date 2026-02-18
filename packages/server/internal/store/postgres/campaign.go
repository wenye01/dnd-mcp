package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/store"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CampaignStore implements store.CampaignStore using PostgreSQL
type CampaignStore struct {
	pool *pgxpool.Pool
}

// NewCampaignStore creates a new campaign store
func NewCampaignStore(client *Client) *CampaignStore {
	return &CampaignStore{pool: client.Pool()}
}

// Create creates a new campaign
func (s *CampaignStore) Create(ctx context.Context, campaign *models.Campaign) error {
	// Generate UUID if not set
	if campaign.ID == "" {
		campaign.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	if campaign.CreatedAt.IsZero() {
		campaign.CreatedAt = now
	}
	campaign.UpdatedAt = now

	// Marshal settings to JSONB
	settingsJSON, err := json.Marshal(campaign.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		INSERT INTO campaigns (id, name, description, dm_id, settings, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = s.pool.Exec(ctx, query,
		campaign.ID,
		campaign.Name,
		nullString(campaign.Description),
		campaign.DMID,
		settingsJSON,
		string(campaign.Status),
		campaign.CreatedAt,
		campaign.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create campaign: %w", err)
	}

	return nil
}

// Get retrieves a campaign by ID
func (s *CampaignStore) Get(ctx context.Context, id string) (*models.Campaign, error) {
	query := `
		SELECT id, name, description, dm_id, settings, status, created_at, updated_at, deleted_at
		FROM campaigns
		WHERE id = $1 AND deleted_at IS NULL
	`

	return s.scanCampaign(ctx, query, id)
}

// GetByIDAndDM retrieves a campaign by ID and DM ID
func (s *CampaignStore) GetByIDAndDM(ctx context.Context, id, dmID string) (*models.Campaign, error) {
	query := `
		SELECT id, name, description, dm_id, settings, status, created_at, updated_at, deleted_at
		FROM campaigns
		WHERE id = $1 AND dm_id = $2 AND deleted_at IS NULL
	`

	return s.scanCampaign(ctx, query, id, dmID)
}

// List lists campaigns with optional filters
func (s *CampaignStore) List(ctx context.Context, filter *store.CampaignFilter) ([]*models.Campaign, error) {
	if filter == nil {
		filter = &store.CampaignFilter{}
	}

	// Build query
	query := `
		SELECT id, name, description, dm_id, settings, status, created_at, updated_at, deleted_at
		FROM campaigns
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if !filter.IncludeDeleted {
		query += " AND deleted_at IS NULL"
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, string(filter.Status))
		argIndex++
	}

	if filter.DMID != "" {
		query += fmt.Sprintf(" AND dm_id = $%d", argIndex)
		args = append(args, filter.DMID)
		argIndex++
	}

	// Order by created_at desc
	query += " ORDER BY created_at DESC"

	// Apply pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list campaigns: %w", err)
	}
	defer rows.Close()

	var campaigns []*models.Campaign
	for rows.Next() {
		campaign, err := scanCampaignFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan campaign: %w", err)
		}
		campaigns = append(campaigns, campaign)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating campaigns: %w", err)
	}

	return campaigns, nil
}

// Update updates a campaign
func (s *CampaignStore) Update(ctx context.Context, campaign *models.Campaign) error {
	campaign.UpdatedAt = time.Now()

	// Marshal settings to JSONB
	settingsJSON, err := json.Marshal(campaign.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		UPDATE campaigns
		SET name = $1, description = $2, dm_id = $3, settings = $4, status = $5, updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL
	`

	result, err := s.pool.Exec(ctx, query,
		campaign.Name,
		nullString(campaign.Description),
		campaign.DMID,
		settingsJSON,
		string(campaign.Status),
		campaign.UpdatedAt,
		campaign.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update campaign: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// Delete soft deletes a campaign
func (s *CampaignStore) Delete(ctx context.Context, id string) error {
	now := time.Now()

	query := `
		UPDATE campaigns
		SET deleted_at = $1, status = $2, updated_at = $3
		WHERE id = $4 AND deleted_at IS NULL
	`

	result, err := s.pool.Exec(ctx, query, now, string(models.CampaignStatusArchived), now, id)
	if err != nil {
		return fmt.Errorf("failed to delete campaign: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// HardDelete permanently deletes a campaign
func (s *CampaignStore) HardDelete(ctx context.Context, id string) error {
	query := `DELETE FROM campaigns WHERE id = $1`

	result, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to hard delete campaign: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// Count counts campaigns with optional filter
func (s *CampaignStore) Count(ctx context.Context, filter *store.CampaignFilter) (int64, error) {
	if filter == nil {
		filter = &store.CampaignFilter{}
	}

	query := "SELECT COUNT(*) FROM campaigns WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if !filter.IncludeDeleted {
		query += " AND deleted_at IS NULL"
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, string(filter.Status))
		argIndex++
	}

	if filter.DMID != "" {
		query += fmt.Sprintf(" AND dm_id = $%d", argIndex)
		args = append(args, filter.DMID)
		argIndex++
	}

	var count int64
	err := s.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count campaigns: %w", err)
	}

	return count, nil
}

// scanCampaign scans a single campaign using the provided query
func (s *CampaignStore) scanCampaign(ctx context.Context, query string, args ...interface{}) (*models.Campaign, error) {
	row := s.pool.QueryRow(ctx, query, args...)
	return scanCampaignFromRow(row)
}

// scanCampaignFromRow scans a campaign from a row
func scanCampaignFromRow(row pgx.Row) (*models.Campaign, error) {
	var (
		id          string
		name        string
		description sql.NullString
		dmID        string
		settingsJSON []byte
		status      string
		createdAt   time.Time
		updatedAt   time.Time
		deletedAt   sql.NullTime
	)

	err := row.Scan(
		&id,
		&name,
		&description,
		&dmID,
		&settingsJSON,
		&status,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan campaign: %w", err)
	}

	// Unmarshal settings
	var settings models.CampaignSettings
	if len(settingsJSON) > 0 {
		if err := json.Unmarshal(settingsJSON, &settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}
	}

	campaign := &models.Campaign{
		ID:          id,
		Name:        name,
		Description: description.String,
		DMID:        dmID,
		Settings:    &settings,
		Status:      models.CampaignStatus(status),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	if deletedAt.Valid {
		campaign.DeletedAt = &deletedAt.Time
	}

	return campaign, nil
}

// nullString returns a sql.NullString for a string
func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// Store errors
var (
	// ErrNotFound record not found error
	ErrNotFound = errors.New("record not found")
)
