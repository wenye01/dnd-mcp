// Package store provides storage interface definitions
package store

import (
	"context"

	"github.com/dnd-mcp/server/internal/models"
)

// CampaignStore campaign storage interface
type CampaignStore interface {
	// Create creates a new campaign
	Create(ctx context.Context, campaign *models.Campaign) error

	// Get retrieves a campaign by ID
	Get(ctx context.Context, id string) (*models.Campaign, error)

	// GetByIDAndDM retrieves a campaign by ID and DM ID
	GetByIDAndDM(ctx context.Context, id, dmID string) (*models.Campaign, error)

	// List lists campaigns with optional filters
	List(ctx context.Context, filter *CampaignFilter) ([]*models.Campaign, error)

	// Update updates a campaign
	Update(ctx context.Context, campaign *models.Campaign) error

	// Delete soft deletes a campaign
	Delete(ctx context.Context, id string) error

	// HardDelete permanently deletes a campaign
	HardDelete(ctx context.Context, id string) error

	// Count counts campaigns with optional filter
	Count(ctx context.Context, filter *CampaignFilter) (int64, error)
}

// CampaignFilter campaign list filter
type CampaignFilter struct {
	// Status filter by status (optional)
	Status models.CampaignStatus

	// DMID filter by DM ID (optional)
	DMID string

	// IncludeDeleted include soft-deleted campaigns
	IncludeDeleted bool

	// Limit max number of results
	Limit int

	// Offset pagination offset
	Offset int
}

// GameStateStore game state storage interface
type GameStateStore interface {
	// Create creates a new game state
	Create(ctx context.Context, gameState *models.GameState) error

	// Get retrieves game state by campaign ID
	Get(ctx context.Context, campaignID string) (*models.GameState, error)

	// GetByID retrieves game state by its own ID
	GetByID(ctx context.Context, id string) (*models.GameState, error)

	// Update updates a game state
	Update(ctx context.Context, gameState *models.GameState) error

	// Delete deletes a game state
	Delete(ctx context.Context, campaignID string) error
}
