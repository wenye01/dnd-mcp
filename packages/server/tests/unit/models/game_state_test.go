package models_test

import (
	"testing"
	"time"

	"github.com/dnd-mcp/server/internal/models"
)

func TestNewGameTime(t *testing.T) {
	gt := models.NewGameTime()

	if gt.Year != 1 {
		t.Errorf("expected Year to be 1, got %d", gt.Year)
	}
	if gt.Month != 1 {
		t.Errorf("expected Month to be 1, got %d", gt.Month)
	}
	if gt.Day != 1 {
		t.Errorf("expected Day to be 1, got %d", gt.Day)
	}
	if gt.Hour != 8 {
		t.Errorf("expected Hour to be 8, got %d", gt.Hour)
	}
	if gt.Minute != 0 {
		t.Errorf("expected Minute to be 0, got %d", gt.Minute)
	}
	if gt.Phase != models.TimePhaseMorning {
		t.Errorf("expected Phase to be 'morning', got %s", gt.Phase)
	}
}

func TestGameTime_Validate(t *testing.T) {
	tests := []struct {
		name     string
		gameTime *models.GameTime
		wantErr  bool
		errField string
	}{
		{
			name:     "valid default time",
			gameTime: models.NewGameTime(),
			wantErr:  false,
		},
		{
			name: "negative year",
			gameTime: &models.GameTime{
				Year: -1, Month: 1, Day: 1, Hour: 0, Minute: 0,
			},
			wantErr:  true,
			errField: "year",
		},
		{
			name: "month too low",
			gameTime: &models.GameTime{
				Year: 1, Month: 0, Day: 1, Hour: 0, Minute: 0,
			},
			wantErr:  true,
			errField: "month",
		},
		{
			name: "month too high",
			gameTime: &models.GameTime{
				Year: 1, Month: 13, Day: 1, Hour: 0, Minute: 0,
			},
			wantErr:  true,
			errField: "month",
		},
		{
			name: "day too low",
			gameTime: &models.GameTime{
				Year: 1, Month: 1, Day: 0, Hour: 0, Minute: 0,
			},
			wantErr:  true,
			errField: "day",
		},
		{
			name: "day too high",
			gameTime: &models.GameTime{
				Year: 1, Month: 1, Day: 32, Hour: 0, Minute: 0,
			},
			wantErr:  true,
			errField: "day",
		},
		{
			name: "hour too high",
			gameTime: &models.GameTime{
				Year: 1, Month: 1, Day: 1, Hour: 24, Minute: 0,
			},
			wantErr:  true,
			errField: "hour",
		},
		{
			name: "minute too high",
			gameTime: &models.GameTime{
				Year: 1, Month: 1, Day: 1, Hour: 0, Minute: 60,
			},
			wantErr:  true,
			errField: "minute",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.gameTime.Validate()
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

func TestGameTime_AddMinutes(t *testing.T) {
	tests := []struct {
		name          string
		initial       *models.GameTime
		minutes       int
		expectedHour  int
		expectedMin   int
		expectedPhase models.TimePhase
	}{
		{
			name:          "add 30 minutes",
			initial:       &models.GameTime{Year: 1, Month: 1, Day: 1, Hour: 8, Minute: 0, Phase: models.TimePhaseMorning},
			minutes:       30,
			expectedHour:  8,
			expectedMin:   30,
			expectedPhase: models.TimePhaseMorning,
		},
		{
			name:          "add 60 minutes (roll over hour)",
			initial:       &models.GameTime{Year: 1, Month: 1, Day: 1, Hour: 8, Minute: 30, Phase: models.TimePhaseMorning},
			minutes:       60,
			expectedHour:  9,
			expectedMin:   30,
			expectedPhase: models.TimePhaseMorning,
		},
		{
			name:          "add 120 minutes (roll over multiple hours)",
			initial:       &models.GameTime{Year: 1, Month: 1, Day: 1, Hour: 10, Minute: 0, Phase: models.TimePhaseMorning},
			minutes:       120,
			expectedHour:  12,
			expectedMin:   0,
			expectedPhase: models.TimePhaseNoon,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initial.AddMinutes(tt.minutes)
			if tt.initial.Hour != tt.expectedHour {
				t.Errorf("expected Hour to be %d, got %d", tt.expectedHour, tt.initial.Hour)
			}
			if tt.initial.Minute != tt.expectedMin {
				t.Errorf("expected Minute to be %d, got %d", tt.expectedMin, tt.initial.Minute)
			}
			if tt.initial.Phase != tt.expectedPhase {
				t.Errorf("expected Phase to be %s, got %s", tt.expectedPhase, tt.initial.Phase)
			}
		})
	}
}

func TestGameTime_AddHours(t *testing.T) {
	tests := []struct {
		name         string
		initial      *models.GameTime
		hours        int
		expectedDay  int
		expectedHour int
	}{
		{
			name:         "add 5 hours",
			initial:      &models.GameTime{Year: 1, Month: 1, Day: 1, Hour: 8, Minute: 0, Phase: models.TimePhaseMorning},
			hours:        5,
			expectedDay:  1,
			expectedHour: 13,
		},
		{
			name:         "add 24 hours (roll over day)",
			initial:      &models.GameTime{Year: 1, Month: 1, Day: 1, Hour: 12, Minute: 0, Phase: models.TimePhaseNoon},
			hours:        24,
			expectedDay:  2,
			expectedHour: 12,
		},
		{
			name:         "add 48 hours (roll over 2 days)",
			initial:      &models.GameTime{Year: 1, Month: 1, Day: 1, Hour: 12, Minute: 0, Phase: models.TimePhaseNoon},
			hours:        48,
			expectedDay:  3,
			expectedHour: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initial.AddHours(tt.hours)
			if tt.initial.Day != tt.expectedDay {
				t.Errorf("expected Day to be %d, got %d", tt.expectedDay, tt.initial.Day)
			}
			if tt.initial.Hour != tt.expectedHour {
				t.Errorf("expected Hour to be %d, got %d", tt.expectedHour, tt.initial.Hour)
			}
		})
	}
}

func TestGameTime_PhaseUpdate(t *testing.T) {
	tests := []struct {
		hour         int
		expectedPhase models.TimePhase
	}{
		{5, models.TimePhaseDawn},
		{6, models.TimePhaseMorning},
		{8, models.TimePhaseMorning},
		{11, models.TimePhaseMorning},
		{12, models.TimePhaseNoon},
		{13, models.TimePhaseNoon},
		{14, models.TimePhaseAfternoon},
		{17, models.TimePhaseAfternoon},
		{18, models.TimePhaseDusk},
		{19, models.TimePhaseDusk},
		{20, models.TimePhaseNight},
		{23, models.TimePhaseNight},
		{0, models.TimePhaseNight},
		{4, models.TimePhaseNight},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			gt := &models.GameTime{
				Year: 1, Month: 1, Day: 1, Hour: tt.hour, Minute: 0,
			}
			gt.AddHours(0) // trigger normalize to update phase
			if gt.Phase != tt.expectedPhase {
				t.Errorf("hour %d: expected Phase to be %s, got %s", tt.hour, tt.expectedPhase, gt.Phase)
			}
		})
	}
}

func TestPosition_Validate(t *testing.T) {
	tests := []struct {
		name     string
		position *models.Position
		wantErr  bool
		errField string
	}{
		{
			name:     "valid position",
			position: &models.Position{X: 10, Y: 20},
			wantErr:  false,
		},
		{
			name:     "zero position",
			position: &models.Position{X: 0, Y: 0},
			wantErr:  false,
		},
		{
			name:     "negative x",
			position: &models.Position{X: -1, Y: 0},
			wantErr:  true,
			errField: "x",
		},
		{
			name:     "negative y",
			position: &models.Position{X: 0, Y: -1},
			wantErr:  true,
			errField: "y",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.position.Validate()
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

func TestNewGameState(t *testing.T) {
	gs := models.NewGameState("campaign-001")

	if gs.ID != "campaign-001" {
		t.Errorf("expected ID to be 'campaign-001', got %s", gs.ID)
	}
	if gs.CampaignID != "campaign-001" {
		t.Errorf("expected CampaignID to be 'campaign-001', got %s", gs.CampaignID)
	}
	if gs.GameTime == nil {
		t.Error("expected GameTime to be initialized")
	}
	if gs.PartyPosition == nil {
		t.Error("expected PartyPosition to be initialized")
	}
	if gs.CurrentMapType != models.MapTypeWorld {
		t.Errorf("expected CurrentMapType to be 'world', got %s", gs.CurrentMapType)
	}
	if gs.Weather != "clear" {
		t.Errorf("expected Weather to be 'clear', got %s", gs.Weather)
	}
	if gs.ActiveCombatID != "" {
		t.Error("expected ActiveCombatID to be empty")
	}
}

func TestGameState_Validate(t *testing.T) {
	tests := []struct {
		name      string
		gameState *models.GameState
		wantErr   bool
		errField  string
	}{
		{
			name:      "valid game state",
			gameState: models.NewGameState("campaign-001"),
			wantErr:   false,
		},
		{
			name: "empty campaign id",
			gameState: &models.GameState{
				CampaignID:     "",
				GameTime:       models.NewGameTime(),
				PartyPosition:  &models.Position{X: 0, Y: 0},
				CurrentMapType: models.MapTypeWorld,
			},
			wantErr:  true,
			errField: "campaign_id",
		},
		{
			name: "invalid game time",
			gameState: &models.GameState{
				CampaignID: "campaign-001",
				GameTime:   &models.GameTime{Year: -1},
			},
			wantErr:  true,
			errField: "year",
		},
		{
			name: "invalid party position",
			gameState: &models.GameState{
				CampaignID:    "campaign-001",
				GameTime:      models.NewGameTime(),
				PartyPosition: &models.Position{X: -1, Y: 0},
			},
			wantErr:  true,
			errField: "x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.gameState.Validate()
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

func TestGameState_CombatMethods(t *testing.T) {
	gs := models.NewGameState("campaign-001")

	if gs.IsInCombat() {
		t.Error("expected not to be in combat initially")
	}

	gs.SetCombat("combat-001")
	if !gs.IsInCombat() {
		t.Error("expected to be in combat")
	}
	if gs.ActiveCombatID != "combat-001" {
		t.Errorf("expected ActiveCombatID to be 'combat-001', got %s", gs.ActiveCombatID)
	}

	gs.ClearCombat()
	if gs.IsInCombat() {
		t.Error("expected not to be in combat after clearing")
	}
}

func TestGameState_MapMethods(t *testing.T) {
	gs := models.NewGameState("campaign-001")

	if gs.IsInBattleMap() {
		t.Error("expected not to be in battle map initially")
	}

	gs.SetCurrentMap("map-001", models.MapTypeBattle)
	if !gs.IsInBattleMap() {
		t.Error("expected to be in battle map")
	}
	if gs.CurrentMapID != "map-001" {
		t.Errorf("expected CurrentMapID to be 'map-001', got %s", gs.CurrentMapID)
	}

	gs.SetCurrentMap("map-002", models.MapTypeWorld)
	if gs.IsInBattleMap() {
		t.Error("expected not to be in battle map after switching to world")
	}
}

func TestGameState_SetPartyPosition(t *testing.T) {
	gs := models.NewGameState("campaign-001")
	oldUpdatedAt := gs.UpdatedAt

	time.Sleep(10 * time.Millisecond)

	err := gs.SetPartyPosition(&models.Position{X: 10, Y: 20})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if gs.PartyPosition.X != 10 || gs.PartyPosition.Y != 20 {
		t.Errorf("expected position (10, 20), got (%d, %d)", gs.PartyPosition.X, gs.PartyPosition.Y)
	}
	if !gs.UpdatedAt.After(oldUpdatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}
}

func TestGameState_SetPartyPosition_Invalid(t *testing.T) {
	gs := models.NewGameState("campaign-001")

	err := gs.SetPartyPosition(&models.Position{X: -1, Y: 0})
	if err == nil {
		t.Error("expected error for invalid position")
	}
}

func TestGameState_SetWeather(t *testing.T) {
	gs := models.NewGameState("campaign-001")
	oldUpdatedAt := gs.UpdatedAt

	time.Sleep(10 * time.Millisecond)

	gs.SetWeather("rain")

	if gs.Weather != "rain" {
		t.Errorf("expected Weather to be 'rain', got %s", gs.Weather)
	}
	if !gs.UpdatedAt.After(oldUpdatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}
}

func TestGameState_AdvanceTime(t *testing.T) {
	gs := models.NewGameState("campaign-001")
	initialHour := gs.GameTime.Hour

	gs.AdvanceTime(5)

	if gs.GameTime.Hour != initialHour+5 {
		t.Errorf("expected Hour to be %d, got %d", initialHour+5, gs.GameTime.Hour)
	}
}

func TestGameState_AdvanceTimeMinutes(t *testing.T) {
	gs := models.NewGameState("campaign-001")
	initialMinute := gs.GameTime.Minute

	gs.AdvanceTimeMinutes(30)

	if gs.GameTime.Minute != initialMinute+30 {
		t.Errorf("expected Minute to be %d, got %d", initialMinute+30, gs.GameTime.Minute)
	}
}
