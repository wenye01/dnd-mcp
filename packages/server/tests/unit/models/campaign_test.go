package models_test

import (
	"testing"
	"time"

	"github.com/dnd-mcp/server/internal/models"
)

func TestNewCampaignSettings(t *testing.T) {
	settings := models.NewCampaignSettings()

	if settings.MaxPlayers != 4 {
		t.Errorf("expected MaxPlayers to be 4, got %d", settings.MaxPlayers)
	}
	if settings.StartLevel != 1 {
		t.Errorf("expected StartLevel to be 1, got %d", settings.StartLevel)
	}
	if settings.Ruleset != "dnd5e" {
		t.Errorf("expected Ruleset to be 'dnd5e', got %s", settings.Ruleset)
	}
	if settings.ContextWindow != 20 {
		t.Errorf("expected ContextWindow to be 20, got %d", settings.ContextWindow)
	}
	if settings.HouseRules == nil {
		t.Error("expected HouseRules to be initialized")
	}
}

func TestCampaignSettings_Validate(t *testing.T) {
	tests := []struct {
		name     string
		settings *models.CampaignSettings
		wantErr  bool
		errField string
	}{
		{
			name:     "valid default settings",
			settings: models.NewCampaignSettings(),
			wantErr:  false,
		},
		{
			name: "max players too low",
			settings: &models.CampaignSettings{
				MaxPlayers:    0,
				StartLevel:    1,
				Ruleset:       "dnd5e",
				ContextWindow: 20,
			},
			wantErr:  true,
			errField: "max_players",
		},
		{
			name: "max players too high",
			settings: &models.CampaignSettings{
				MaxPlayers:    11,
				StartLevel:    1,
				Ruleset:       "dnd5e",
				ContextWindow: 20,
			},
			wantErr:  true,
			errField: "max_players",
		},
		{
			name: "start level too low",
			settings: &models.CampaignSettings{
				MaxPlayers:    4,
				StartLevel:    0,
				Ruleset:       "dnd5e",
				ContextWindow: 20,
			},
			wantErr:  true,
			errField: "start_level",
		},
		{
			name: "start level too high",
			settings: &models.CampaignSettings{
				MaxPlayers:    4,
				StartLevel:    21,
				Ruleset:       "dnd5e",
				ContextWindow: 20,
			},
			wantErr:  true,
			errField: "start_level",
		},
		{
			name: "context window zero",
			settings: &models.CampaignSettings{
				MaxPlayers:    4,
				StartLevel:    1,
				Ruleset:       "dnd5e",
				ContextWindow: 0,
			},
			wantErr:  true,
			errField: "context_window",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.settings.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if verr, ok := err.(*models.ValidationError); ok {
					if verr.Field != tt.errField {
						t.Errorf("expected error field '%s', got '%s'", tt.errField, verr.Field)
					}
				} else {
					t.Errorf("expected ValidationError, got %T", err)
				}
			}
		})
	}
}

func TestNewCampaign(t *testing.T) {
	campaign := models.NewCampaign("Test Campaign", "dm-001", "A test campaign")

	if campaign.Name != "Test Campaign" {
		t.Errorf("expected Name to be 'Test Campaign', got %s", campaign.Name)
	}
	if campaign.DMID != "dm-001" {
		t.Errorf("expected DMID to be 'dm-001', got %s", campaign.DMID)
	}
	if campaign.Description != "A test campaign" {
		t.Errorf("expected Description to be 'A test campaign', got %s", campaign.Description)
	}
	if campaign.Status != models.CampaignStatusActive {
		t.Errorf("expected Status to be active, got %s", campaign.Status)
	}
	if campaign.Settings == nil {
		t.Error("expected Settings to be initialized")
	}
	if campaign.DeletedAt != nil {
		t.Error("expected DeletedAt to be nil for new campaign")
	}
}

func TestCampaign_Validate(t *testing.T) {
	tests := []struct {
		name     string
		campaign *models.Campaign
		wantErr  bool
		errField string
	}{
		{
			name:     "valid campaign",
			campaign: models.NewCampaign("Test", "dm-001", "Description"),
			wantErr:  false,
		},
		{
			name: "empty name",
			campaign: &models.Campaign{
				Name:     "",
				DMID:     "dm-001",
				Settings: models.NewCampaignSettings(),
				Status:   models.CampaignStatusActive,
			},
			wantErr:  true,
			errField: "name",
		},
		{
			name: "empty dm id",
			campaign: &models.Campaign{
				Name:     "Test",
				DMID:     "",
				Settings: models.NewCampaignSettings(),
				Status:   models.CampaignStatusActive,
			},
			wantErr:  true,
			errField: "dm_id",
		},
		{
			name: "invalid settings",
			campaign: &models.Campaign{
				Name: "Test",
				DMID: "dm-001",
				Settings: &models.CampaignSettings{
					MaxPlayers:    0, // invalid
					StartLevel:    1,
					ContextWindow: 20,
				},
				Status: models.CampaignStatusActive,
			},
			wantErr:  true,
			errField: "max_players",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.campaign.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if verr, ok := err.(*models.ValidationError); ok {
					if verr.Field != tt.errField {
						t.Errorf("expected error field '%s', got '%s'", tt.errField, verr.Field)
					}
				} else {
					t.Errorf("expected ValidationError, got %T", err)
				}
			}
		})
	}
}

func TestCampaign_StatusMethods(t *testing.T) {
	campaign := models.NewCampaign("Test", "dm-001", "")

	// Initially active
	if !campaign.IsActive() {
		t.Error("expected campaign to be active")
	}

	// Pause
	campaign.Pause()
	if !campaign.IsPaused() {
		t.Error("expected campaign to be paused")
	}
	if campaign.IsActive() {
		t.Error("expected campaign to not be active")
	}

	// Resume
	campaign.Resume()
	if !campaign.IsActive() {
		t.Error("expected campaign to be active after resume")
	}

	// Finish
	campaign.Finish()
	if !campaign.IsFinished() {
		t.Error("expected campaign to be finished")
	}

	// Archive
	campaign.Archive()
	if !campaign.IsArchived() {
		t.Error("expected campaign to be archived")
	}
	if campaign.DeletedAt == nil {
		t.Error("expected DeletedAt to be set when archived")
	}
}

func TestCampaign_UpdateSettings(t *testing.T) {
	campaign := models.NewCampaign("Test", "dm-001", "")
	oldUpdatedAt := campaign.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	newSettings := &models.CampaignSettings{
		MaxPlayers:    6,
		StartLevel:    3,
		Ruleset:       "dnd5e",
		ContextWindow: 30,
		HouseRules:    map[string]interface{}{"custom_rule": true},
	}

	err := campaign.UpdateSettings(newSettings)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if campaign.Settings.MaxPlayers != 6 {
		t.Errorf("expected MaxPlayers to be 6, got %d", campaign.Settings.MaxPlayers)
	}
	if campaign.Settings.StartLevel != 3 {
		t.Errorf("expected StartLevel to be 3, got %d", campaign.Settings.StartLevel)
	}
	if !campaign.UpdatedAt.After(oldUpdatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}
}

func TestCampaign_UpdateSettings_Invalid(t *testing.T) {
	campaign := models.NewCampaign("Test", "dm-001", "")

	invalidSettings := &models.CampaignSettings{
		MaxPlayers:    0, // invalid
		StartLevel:    1,
		ContextWindow: 20,
	}

	err := campaign.UpdateSettings(invalidSettings)
	if err == nil {
		t.Error("expected error for invalid settings")
	}
}

func TestCampaign_UpdateDescription(t *testing.T) {
	campaign := models.NewCampaign("Test", "dm-001", "Old description")
	oldUpdatedAt := campaign.UpdatedAt

	time.Sleep(10 * time.Millisecond)

	campaign.UpdateDescription("New description")

	if campaign.Description != "New description" {
		t.Errorf("expected Description to be 'New description', got %s", campaign.Description)
	}
	if !campaign.UpdatedAt.After(oldUpdatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}
}

func TestValidationError(t *testing.T) {
	err := models.NewValidationError("field_name", "is invalid")

	if err.Field != "field_name" {
		t.Errorf("expected Field to be 'field_name', got %s", err.Field)
	}
	if err.Message != "is invalid" {
		t.Errorf("expected Message to be 'is invalid', got %s", err.Message)
	}
	if err.Error() != "field_name: is invalid" {
		t.Errorf("expected Error() to be 'field_name: is invalid', got %s", err.Error())
	}
}
