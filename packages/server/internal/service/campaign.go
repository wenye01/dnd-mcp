// Package service provides business logic layer implementations
package service

import (
	"context"
	"fmt"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/store"
	"github.com/google/uuid"
)

// CampaignStore defines the interface for campaign data operations
type CampaignStore interface {
	Create(ctx context.Context, campaign *models.Campaign) error
	Get(ctx context.Context, id string) (*models.Campaign, error)
	GetByIDAndDM(ctx context.Context, id, dmID string) (*models.Campaign, error)
	List(ctx context.Context, filter *store.CampaignFilter) ([]*models.Campaign, error)
	Update(ctx context.Context, campaign *models.Campaign) error
	Delete(ctx context.Context, id string) error
	HardDelete(ctx context.Context, id string) error
	Count(ctx context.Context, filter *store.CampaignFilter) (int64, error)
}

// GameStateStore defines the interface for game state data operations
type GameStateStore interface {
	Create(ctx context.Context, gameState *models.GameState) error
	Get(ctx context.Context, campaignID string) (*models.GameState, error)
	Update(ctx context.Context, gameState *models.GameState) error
	Delete(ctx context.Context, campaignID string) error
}

// CampaignService provides campaign business logic
type CampaignService struct {
	campaignStore CampaignStore
	gameStateStore GameStateStore
}

// NewCampaignService creates a new campaign service
func NewCampaignService(campaignStore CampaignStore, gameStateStore GameStateStore) *CampaignService {
	return &CampaignService{
		campaignStore:  campaignStore,
		gameStateStore: gameStateStore,
	}
}

// CreateCampaignRequest represents a campaign creation request
type CreateCampaignRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	DMID        string                 `json:"dm_id"`
	Settings    *CampaignSettingsInput `json:"settings"`
}

// CampaignSettingsInput represents campaign settings input
type CampaignSettingsInput struct {
	MaxPlayers    int                    `json:"max_players"`
	StartLevel    int                    `json:"start_level"`
	Ruleset       string                 `json:"ruleset"`
	HouseRules    map[string]interface{} `json:"house_rules"`
	ContextWindow int                    `json:"context_window"`
}

// UpdateCampaignRequest represents a campaign update request
type UpdateCampaignRequest struct {
	Name        *string                `json:"name"`
	Description *string                `json:"description"`
	Settings    *CampaignSettingsInput `json:"settings"`
	Status      *models.CampaignStatus `json:"status"`
}

// ListCampaignsRequest represents a campaign list request
type ListCampaignsRequest struct {
	Status         models.CampaignStatus `json:"status"`
	DMID           string                `json:"dm_id"`
	IncludeDeleted bool                  `json:"include_deleted"`
	Limit          int                   `json:"limit"`
	Offset         int                   `json:"offset"`
}

// CreateCampaign creates a new campaign with associated game state
func (s *CampaignService) CreateCampaign(ctx context.Context, req *CreateCampaignRequest) (*models.Campaign, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Create campaign model
	campaign := models.NewCampaign(req.Name, req.DMID, req.Description)

	// Apply custom settings if provided
	if req.Settings != nil {
		s.applySettings(campaign.Settings, req.Settings)
	}

	// Validate campaign
	if err := campaign.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Generate UUID
	campaign.ID = uuid.New().String()

	// Save campaign
	if err := s.campaignStore.Create(ctx, campaign); err != nil {
		return nil, fmt.Errorf("failed to create campaign: %w", err)
	}

	// Create associated game state
	gameState := models.NewGameState(campaign.ID)
	if err := s.gameStateStore.Create(ctx, gameState); err != nil {
		// Try to clean up the campaign if game state creation fails
		_ = s.campaignStore.HardDelete(ctx, campaign.ID)
		return nil, fmt.Errorf("failed to create game state: %w", err)
	}

	return campaign, nil
}

// GetCampaign retrieves a campaign by ID
func (s *CampaignService) GetCampaign(ctx context.Context, campaignID string) (*models.Campaign, error) {
	if campaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}

	campaign, err := s.campaignStore.Get(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	return campaign, nil
}

// GetCampaignByDM retrieves a campaign by ID and DM ID
func (s *CampaignService) GetCampaignByDM(ctx context.Context, campaignID, dmID string) (*models.Campaign, error) {
	if campaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}
	if dmID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "DM ID is required")
	}

	campaign, err := s.campaignStore.GetByIDAndDM(ctx, campaignID, dmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	return campaign, nil
}

// ListCampaigns lists campaigns with optional filtering
func (s *CampaignService) ListCampaigns(ctx context.Context, req *ListCampaignsRequest) ([]*models.Campaign, error) {
	filter := &store.CampaignFilter{
		Status:         req.Status,
		DMID:           req.DMID,
		IncludeDeleted: req.IncludeDeleted,
		Limit:          req.Limit,
		Offset:         req.Offset,
	}

	// Set default limit if not specified
	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	campaigns, err := s.campaignStore.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list campaigns: %w", err)
	}

	return campaigns, nil
}

// UpdateCampaign updates a campaign
func (s *CampaignService) UpdateCampaign(ctx context.Context, campaignID string, req *UpdateCampaignRequest) (*models.Campaign, error) {
	if campaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}

	// Get existing campaign
	campaign, err := s.campaignStore.Get(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	// Apply updates
	if req.Name != nil {
		if *req.Name == "" {
			return nil, NewServiceError(ErrCodeInvalidInput, "campaign name cannot be empty")
		}
		campaign.Name = *req.Name
	}

	if req.Description != nil {
		campaign.Description = *req.Description
	}

	if req.Settings != nil {
		s.applySettings(campaign.Settings, req.Settings)
		if err := campaign.Settings.Validate(); err != nil {
			return nil, fmt.Errorf("invalid settings: %w", err)
		}
	}

	if req.Status != nil {
		if err := s.validateStatusTransition(campaign.Status, *req.Status); err != nil {
			return nil, err
		}
		campaign.Status = *req.Status
	}

	// Validate updated campaign
	if err := campaign.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save updates
	if err := s.campaignStore.Update(ctx, campaign); err != nil {
		return nil, fmt.Errorf("failed to update campaign: %w", err)
	}

	return campaign, nil
}

// DeleteCampaign soft deletes a campaign
func (s *CampaignService) DeleteCampaign(ctx context.Context, campaignID string) error {
	if campaignID == "" {
		return NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}

	// Verify campaign exists
	_, err := s.campaignStore.Get(ctx, campaignID)
	if err != nil {
		return fmt.Errorf("failed to get campaign: %w", err)
	}

	// Delete campaign (soft delete)
	if err := s.campaignStore.Delete(ctx, campaignID); err != nil {
		return fmt.Errorf("failed to delete campaign: %w", err)
	}

	return nil
}

// GetCampaignSummary retrieves a campaign summary for context
func (s *CampaignService) GetCampaignSummary(ctx context.Context, campaignID string) (*GameSummary, error) {
	if campaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}

	// Get campaign
	campaign, err := s.campaignStore.Get(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	// Get game state
	gameState, err := s.gameStateStore.Get(ctx, campaignID)
	if err != nil {
		// Game state might not exist for older campaigns
		gameState = nil
	}

	// Build summary
	summary := &GameSummary{
		CampaignName: campaign.Name,
		CampaignID:   campaign.ID,
		Status:       campaign.Status,
	}

	if gameState != nil {
		summary.GameTime = gameState.GameTime
		summary.Weather = gameState.Weather
		if gameState.PartyPosition != nil {
			summary.CurrentLocation = fmt.Sprintf("(%d, %d)", gameState.PartyPosition.X, gameState.PartyPosition.Y)
		}
		summary.IsInCombat = gameState.IsInCombat()
	}

	return summary, nil
}

// GetGameState retrieves the game state for a campaign
func (s *CampaignService) GetGameState(ctx context.Context, campaignID string) (*models.GameState, error) {
	if campaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}

	gameState, err := s.gameStateStore.Get(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game state: %w", err)
	}

	return gameState, nil
}

// CountCampaigns counts campaigns with optional filtering
func (s *CampaignService) CountCampaigns(ctx context.Context, req *ListCampaignsRequest) (int64, error) {
	filter := &store.CampaignFilter{
		Status:         req.Status,
		DMID:           req.DMID,
		IncludeDeleted: req.IncludeDeleted,
	}

	count, err := s.campaignStore.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count campaigns: %w", err)
	}

	return count, nil
}

// validateCreateRequest validates the create campaign request
func (s *CampaignService) validateCreateRequest(req *CreateCampaignRequest) error {
	if req.Name == "" {
		return NewServiceError(ErrCodeInvalidInput, "campaign name is required")
	}
	if req.DMID == "" {
		return NewServiceError(ErrCodeInvalidInput, "DM ID is required")
	}
	return nil
}

// applySettings applies input settings to campaign settings
func (s *CampaignService) applySettings(settings *models.CampaignSettings, input *CampaignSettingsInput) {
	if input.MaxPlayers > 0 {
		settings.MaxPlayers = input.MaxPlayers
	}
	if input.StartLevel > 0 {
		settings.StartLevel = input.StartLevel
	}
	if input.Ruleset != "" {
		settings.Ruleset = input.Ruleset
	}
	if input.HouseRules != nil {
		settings.HouseRules = input.HouseRules
	}
	if input.ContextWindow > 0 {
		settings.ContextWindow = input.ContextWindow
	}
}

// validateStatusTransition validates status transitions
func (s *CampaignService) validateStatusTransition(from, to models.CampaignStatus) error {
	// Define valid transitions
	validTransitions := map[models.CampaignStatus][]models.CampaignStatus{
		models.CampaignStatusActive: {
			models.CampaignStatusPaused,
			models.CampaignStatusFinished,
			models.CampaignStatusArchived,
		},
		models.CampaignStatusPaused: {
			models.CampaignStatusActive,
			models.CampaignStatusFinished,
			models.CampaignStatusArchived,
		},
		models.CampaignStatusFinished: {
			models.CampaignStatusArchived,
		},
		models.CampaignStatusArchived: {}, // No transitions from archived
	}

	// Check if transition is valid
	allowed, exists := validTransitions[from]
	if !exists {
		return NewServiceError(ErrCodeInvalidState, fmt.Sprintf("unknown status: %s", from))
	}

	for _, status := range allowed {
		if status == to {
			return nil
		}
	}

	return NewServiceError(ErrCodeInvalidState, fmt.Sprintf("cannot transition from %s to %s", from, to))
}

// GameSummary represents a campaign summary for context
type GameSummary struct {
	CampaignID      string            `json:"campaign_id"`
	CampaignName    string            `json:"campaign_name"`
	Status          models.CampaignStatus `json:"status"`
	GameTime        *models.GameTime `json:"game_time,omitempty"`
	CurrentLocation string            `json:"current_location,omitempty"`
	Weather         string            `json:"weather,omitempty"`
	IsInCombat      bool              `json:"is_in_combat"`
	PartyMembers    []PartyMember     `json:"party_members,omitempty"`
	RecentEvents    []string          `json:"recent_events,omitempty"`
}

// PartyMember represents a party member summary
type PartyMember struct {
	CharacterID string `json:"character_id"`
	Name        string `json:"name"`
	Class       string `json:"class"`
	Level       int    `json:"level"`
}

// Service errors
const (
	ErrCodeInvalidInput = "INVALID_INPUT"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeInvalidState = "INVALID_STATE"
)

// ServiceError represents a service-level error
type ServiceError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *ServiceError) Error() string {
	return e.Message
}

// NewServiceError creates a new service error
func NewServiceError(code, message string) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
	}
}

// IsServiceError checks if an error is a service error
func IsServiceError(err error) bool {
	_, ok := err.(*ServiceError)
	return ok
}

// GetServiceError returns the service error if the error is one
func GetServiceError(err error) *ServiceError {
	if se, ok := err.(*ServiceError); ok {
		return se
	}
	return nil
}
