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

// CharacterStore character storage interface
type CharacterStore interface {
	// Create creates a new character
	Create(ctx context.Context, character *models.Character) error

	// Get retrieves a character by ID
	Get(ctx context.Context, id string) (*models.Character, error)

	// GetByCampaignAndID retrieves a character by campaign ID and character ID
	GetByCampaignAndID(ctx context.Context, campaignID, id string) (*models.Character, error)

	// List lists characters with optional filters
	List(ctx context.Context, filter *CharacterFilter) ([]*models.Character, error)

	// Update updates a character
	Update(ctx context.Context, character *models.Character) error

	// Delete deletes a character
	Delete(ctx context.Context, id string) error

	// Count counts characters with optional filter
	Count(ctx context.Context, filter *CharacterFilter) (int64, error)
}

// CharacterFilter character list filter
type CharacterFilter struct {
	// CampaignID filter by campaign ID (optional)
	CampaignID string

	// IsNPC filter by NPC status (optional, nil means all)
	IsNPC *bool

	// PlayerID filter by player ID (optional)
	PlayerID string

	// NPCType filter by NPC type (optional)
	NPCType models.NPCType

	// Limit max number of results
	Limit int

	// Offset pagination offset
	Offset int
}

// CombatStore combat storage interface
type CombatStore interface {
	// Create creates a new combat
	Create(ctx context.Context, combat *models.Combat) error

	// Get retrieves a combat by ID
	Get(ctx context.Context, id string) (*models.Combat, error)

	// GetByCampaign retrieves combats by campaign ID
	GetByCampaign(ctx context.Context, campaignID string) ([]*models.Combat, error)

	// GetActive retrieves the active combat for a campaign
	GetActive(ctx context.Context, campaignID string) (*models.Combat, error)

	// Update updates a combat
	Update(ctx context.Context, combat *models.Combat) error

	// Delete deletes a combat
	Delete(ctx context.Context, id string) error
}

// MapStore map storage interface
type MapStore interface {
	// Create creates a new map
	Create(ctx context.Context, gameMap *models.Map) error

	// Get retrieves a map by ID
	Get(ctx context.Context, id string) (*models.Map, error)

	// GetByCampaign retrieves maps by campaign ID
	GetByCampaign(ctx context.Context, campaignID string) ([]*models.Map, error)

	// Update updates a map
	Update(ctx context.Context, gameMap *models.Map) error

	// Delete deletes a map
	Delete(ctx context.Context, id string) error

	// GetWorldMap retrieves the world map for a campaign
	GetWorldMap(ctx context.Context, campaignID string) (*models.Map, error)

	// GetBattleMap retrieves a battle map by ID
	GetBattleMap(ctx context.Context, id string) (*models.Map, error)

	// GetByParent retrieves battle maps by parent location
	GetByParent(ctx context.Context, parentID string) ([]*models.Map, error)
}
