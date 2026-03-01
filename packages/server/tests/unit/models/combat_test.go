package models_test

import (
	"testing"
	"time"

	"github.com/dnd-mcp/server/internal/models"
)

func TestNewCombat(t *testing.T) {
	participantIDs := []string{"char-001", "char-002", "char-003"}
	combat := models.NewCombat("campaign-001", participantIDs)

	if combat.ID == "" {
		t.Error("expected ID to be generated")
	}
	if combat.CampaignID != "campaign-001" {
		t.Errorf("expected CampaignID to be 'campaign-001', got %s", combat.CampaignID)
	}
	if combat.Status != models.CombatStatusActive {
		t.Errorf("expected Status to be 'active', got %s", combat.Status)
	}
	if combat.Round != 1 {
		t.Errorf("expected Round to be 1, got %d", combat.Round)
	}
	if combat.TurnIndex != 0 {
		t.Errorf("expected TurnIndex to be 0, got %d", combat.TurnIndex)
	}
	if len(combat.Participants) != 3 {
		t.Errorf("expected 3 participants, got %d", len(combat.Participants))
	}
	if len(combat.Log) != 0 {
		t.Errorf("expected empty log, got %d entries", len(combat.Log))
	}
	if combat.StartedAt.IsZero() {
		t.Error("expected StartedAt to be set")
	}
	if combat.EndedAt != nil {
		t.Error("expected EndedAt to be nil")
	}

	// 验证参战者初始化
	for i, p := range combat.Participants {
		if p.CharacterID != participantIDs[i] {
			t.Errorf("expected participant %d CharacterID to be '%s', got '%s'", i, participantIDs[i], p.CharacterID)
		}
		if p.Initiative != 0 {
			t.Errorf("expected participant %d Initiative to be 0, got %d", i, p.Initiative)
		}
		if p.HasActed != false {
			t.Errorf("expected participant %d HasActed to be false", i)
		}
	}
}

func TestNewCombat_EmptyParticipants(t *testing.T) {
	combat := models.NewCombat("campaign-001", []string{})

	if len(combat.Participants) != 0 {
		t.Errorf("expected 0 participants, got %d", len(combat.Participants))
	}
}

func TestCombat_Validate(t *testing.T) {
	tests := []struct {
		name      string
		combat    *models.Combat
		wantErr   bool
		errField  string
	}{
		{
			name:      "valid combat",
			combat:    models.NewCombat("campaign-001", []string{"char-001"}),
			wantErr:   false,
		},
		{
			name: "empty campaign id",
			combat: &models.Combat{
				CampaignID:   "",
				Status:       models.CombatStatusActive,
				Round:        1,
				TurnIndex:    0,
				Participants: []models.Participant{{CharacterID: "char-001"}},
			},
			wantErr:  true,
			errField: "campaign_id",
		},
		{
			name: "round less than 1",
			combat: &models.Combat{
				CampaignID:   "campaign-001",
				Status:       models.CombatStatusActive,
				Round:        0,
				TurnIndex:    0,
				Participants: []models.Participant{{CharacterID: "char-001"}},
			},
			wantErr:  true,
			errField: "round",
		},
		{
			name: "negative turn index",
			combat: &models.Combat{
				CampaignID:   "campaign-001",
				Status:       models.CombatStatusActive,
				Round:        1,
				TurnIndex:    -1,
				Participants: []models.Participant{{CharacterID: "char-001"}},
			},
			wantErr:  true,
			errField: "turn_index",
		},
		{
			name: "empty participants",
			combat: &models.Combat{
				CampaignID:   "campaign-001",
				Status:       models.CombatStatusActive,
				Round:        1,
				TurnIndex:    0,
				Participants: []models.Participant{},
			},
			wantErr:  true,
			errField: "participants",
		},
		{
			name: "turn index exceeds participants",
			combat: &models.Combat{
				CampaignID:   "campaign-001",
				Status:       models.CombatStatusActive,
				Round:        1,
				TurnIndex:    5,
				Participants: []models.Participant{{CharacterID: "char-001"}},
			},
			wantErr:  true,
			errField: "turn_index",
		},
		{
			name: "invalid participant",
			combat: &models.Combat{
				CampaignID:   "campaign-001",
				Status:       models.CombatStatusActive,
				Round:        1,
				TurnIndex:    0,
				Participants: []models.Participant{{CharacterID: ""}},
			},
			wantErr:  true,
			errField: "participants",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.combat.Validate()
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

func TestCombat_AddLogEntry(t *testing.T) {
	combat := models.NewCombat("campaign-001", []string{"char-001", "char-002"})

	combat.AddLogEntry("char-001", "attack", "char-002", "hit for 8 damage")

	if len(combat.Log) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(combat.Log))
	}

	entry := combat.Log[0]
	if entry.Round != 1 {
		t.Errorf("expected Round to be 1, got %d", entry.Round)
	}
	if entry.ActorID != "char-001" {
		t.Errorf("expected ActorID to be 'char-001', got %s", entry.ActorID)
	}
	if entry.Action != "attack" {
		t.Errorf("expected Action to be 'attack', got %s", entry.Action)
	}
	if entry.TargetID != "char-002" {
		t.Errorf("expected TargetID to be 'char-002', got %s", entry.TargetID)
	}
	if entry.Result != "hit for 8 damage" {
		t.Errorf("expected Result to be 'hit for 8 damage', got %s", entry.Result)
	}
	if entry.Timestamp.IsZero() {
		t.Error("expected Timestamp to be set")
	}
}

func TestCombat_GetCurrentParticipant(t *testing.T) {
	combat := models.NewCombat("campaign-001", []string{"char-001", "char-002", "char-003"})

	// 初始状态，应该是第一个参战者
	current := combat.GetCurrentParticipant()
	if current == nil {
		t.Fatal("expected current participant, got nil")
	}
	if current.CharacterID != "char-001" {
		t.Errorf("expected CharacterID to be 'char-001', got %s", current.CharacterID)
	}

	// 推进一个回合
	combat.TurnIndex = 1
	current = combat.GetCurrentParticipant()
	if current == nil {
		t.Fatal("expected current participant, got nil")
	}
	if current.CharacterID != "char-002" {
		t.Errorf("expected CharacterID to be 'char-002', got %s", current.CharacterID)
	}
}

func TestCombat_GetCurrentParticipant_Empty(t *testing.T) {
	combat := models.NewCombat("campaign-001", []string{})

	current := combat.GetCurrentParticipant()
	if current != nil {
		t.Errorf("expected nil for empty participants, got %+v", current)
	}
}

func TestCombat_GetCurrentParticipant_IndexOutOfRange(t *testing.T) {
	combat := models.NewCombat("campaign-001", []string{"char-001"})
	combat.TurnIndex = 10 // 超出范围

	current := combat.GetCurrentParticipant()
	if current != nil {
		t.Errorf("expected nil for out of range index, got %+v", current)
	}
}

func TestCombat_AdvanceTurn(t *testing.T) {
	combat := models.NewCombat("campaign-001", []string{"char-001", "char-002", "char-003"})

	// 初始状态
	if combat.TurnIndex != 0 {
		t.Errorf("expected TurnIndex to be 0, got %d", combat.TurnIndex)
	}
	if combat.Participants[0].HasActed != false {
		t.Error("expected first participant to not have acted")
	}

	// 推进第一个回合
	isNewRound := combat.AdvanceTurn()
	if isNewRound {
		t.Error("expected not to be a new round")
	}
	if combat.TurnIndex != 1 {
		t.Errorf("expected TurnIndex to be 1, got %d", combat.TurnIndex)
	}
	if combat.Participants[0].HasActed != true {
		t.Error("expected first participant to have acted")
	}
	if combat.Round != 1 {
		t.Errorf("expected Round to still be 1, got %d", combat.Round)
	}

	// 推进第二个回合
	isNewRound = combat.AdvanceTurn()
	if isNewRound {
		t.Error("expected not to be a new round")
	}
	if combat.TurnIndex != 2 {
		t.Errorf("expected TurnIndex to be 2, got %d", combat.TurnIndex)
	}

	// 推进第三个回合（应该进入新回合）
	isNewRound = combat.AdvanceTurn()
	if !isNewRound {
		t.Error("expected to be a new round")
	}
	if combat.TurnIndex != 0 {
		t.Errorf("expected TurnIndex to be 0, got %d", combat.TurnIndex)
	}
	if combat.Round != 2 {
		t.Errorf("expected Round to be 2, got %d", combat.Round)
	}
	// 检查所有参战者的行动状态已重置
	for i, p := range combat.Participants {
		if p.HasActed != false {
			t.Errorf("expected participant %d HasActed to be reset to false", i)
		}
	}
}

func TestCombat_AdvanceTurn_EmptyParticipants(t *testing.T) {
	combat := models.NewCombat("campaign-001", []string{})

	isNewRound := combat.AdvanceTurn()
	if isNewRound {
		t.Error("expected false for empty participants")
	}
}

func TestCombat_SortParticipantsByInitiative(t *testing.T) {
	combat := models.NewCombat("campaign-001", []string{"char-001", "char-002", "char-003"})

	// 设置先攻值
	combat.Participants[0].Initiative = 15
	combat.Participants[1].Initiative = 20
	combat.Participants[2].Initiative = 10

	combat.SortParticipantsByInitiative()

	// 验证排序结果（从高到低）
	if combat.Participants[0].CharacterID != "char-002" || combat.Participants[0].Initiative != 20 {
		t.Errorf("expected first to be char-002 with initiative 20, got %s with %d",
			combat.Participants[0].CharacterID, combat.Participants[0].Initiative)
	}
	if combat.Participants[1].CharacterID != "char-001" || combat.Participants[1].Initiative != 15 {
		t.Errorf("expected second to be char-001 with initiative 15, got %s with %d",
			combat.Participants[1].CharacterID, combat.Participants[1].Initiative)
	}
	if combat.Participants[2].CharacterID != "char-003" || combat.Participants[2].Initiative != 10 {
		t.Errorf("expected third to be char-003 with initiative 10, got %s with %d",
			combat.Participants[2].CharacterID, combat.Participants[2].Initiative)
	}
}

func TestCombat_SortParticipantsByInitiative_SameInitiative(t *testing.T) {
	combat := models.NewCombat("campaign-001", []string{"char-001", "char-002"})

	// 设置相同的先攻值
	combat.Participants[0].Initiative = 15
	combat.Participants[1].Initiative = 15

	combat.SortParticipantsByInitiative()

	// 相同先攻值时，排序顺序不稳定，但不应崩溃
	if len(combat.Participants) != 2 {
		t.Errorf("expected 2 participants, got %d", len(combat.Participants))
	}
}

func TestCombat_GetParticipantByCharacterID(t *testing.T) {
	combat := models.NewCombat("campaign-001", []string{"char-001", "char-002", "char-003"})

	// 查找存在的参战者
	p := combat.GetParticipantByCharacterID("char-002")
	if p == nil {
		t.Fatal("expected participant, got nil")
	}
	if p.CharacterID != "char-002" {
		t.Errorf("expected CharacterID to be 'char-002', got %s", p.CharacterID)
	}

	// 查找不存在的参战者
	p = combat.GetParticipantByCharacterID("char-999")
	if p != nil {
		t.Errorf("expected nil for non-existent participant, got %+v", p)
	}
}

func TestCombat_End(t *testing.T) {
	combat := models.NewCombat("campaign-001", []string{"char-001"})

	if combat.IsFinished() {
		t.Error("expected combat to not be finished initially")
	}
	if !combat.IsActive() {
		t.Error("expected combat to be active initially")
	}
	if combat.EndedAt != nil {
		t.Error("expected EndedAt to be nil initially")
	}

	// 稍微等待以确保时间戳不同
	time.Sleep(10 * time.Millisecond)

	combat.End()

	if !combat.IsFinished() {
		t.Error("expected combat to be finished after End()")
	}
	if combat.IsActive() {
		t.Error("expected combat to not be active after End()")
	}
	if combat.Status != models.CombatStatusFinished {
		t.Errorf("expected Status to be 'finished', got %s", combat.Status)
	}
	if combat.EndedAt == nil {
		t.Error("expected EndedAt to be set")
	}
	if combat.EndedAt.Before(combat.StartedAt) {
		t.Error("expected EndedAt to be after StartedAt")
	}
}

func TestCombat_IsFinished(t *testing.T) {
	combat := models.NewCombat("campaign-001", []string{"char-001"})

	if combat.IsFinished() {
		t.Error("expected combat to not be finished")
	}

	combat.Status = models.CombatStatusFinished
	if !combat.IsFinished() {
		t.Error("expected combat to be finished")
	}
}

func TestCombat_IsActive(t *testing.T) {
	combat := models.NewCombat("campaign-001", []string{"char-001"})

	if !combat.IsActive() {
		t.Error("expected combat to be active")
	}

	combat.Status = models.CombatStatusFinished
	if combat.IsActive() {
		t.Error("expected combat to not be active")
	}
}

// ============ Participant 测试 ============

func TestParticipant_Validate(t *testing.T) {
	tests := []struct {
		name        string
		participant *models.Participant
		wantErr     bool
		errField    string
	}{
		{
			name: "valid participant",
			participant: &models.Participant{
				CharacterID: "char-001",
				Initiative:  15,
				TempHP:      5,
			},
			wantErr: false,
		},
		{
			name: "empty character id",
			participant: &models.Participant{
				CharacterID: "",
				Initiative:  15,
			},
			wantErr:  true,
			errField: "character_id",
		},
		{
			name: "negative temp hp",
			participant: &models.Participant{
				CharacterID: "char-001",
				TempHP:      -5,
			},
			wantErr:  true,
			errField: "temp_hp",
		},
		{
			name: "invalid position",
			participant: &models.Participant{
				CharacterID: "char-001",
				Position:    &models.Position{X: -1, Y: 0},
			},
			wantErr:  true,
			errField: "x",
		},
		{
			name: "valid with position",
			participant: &models.Participant{
				CharacterID: "char-001",
				Position:    &models.Position{X: 10, Y: 20},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.participant.Validate()
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

func TestParticipant_SetPosition(t *testing.T) {
	p := &models.Participant{
		CharacterID: "char-001",
	}

	if p.Position != nil {
		t.Error("expected Position to be nil initially")
	}

	p.SetPosition(10, 20)

	if p.Position == nil {
		t.Fatal("expected Position to be set")
	}
	if p.Position.X != 10 {
		t.Errorf("expected X to be 10, got %d", p.Position.X)
	}
	if p.Position.Y != 20 {
		t.Errorf("expected Y to be 20, got %d", p.Position.Y)
	}
}

func TestParticipant_AddCondition(t *testing.T) {
	p := &models.Participant{
		CharacterID: "char-001",
		Conditions:  make([]models.Condition, 0),
	}

	p.AddCondition("poisoned", 3, "trap")

	if len(p.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(p.Conditions))
	}
	if p.Conditions[0].Type != "poisoned" {
		t.Errorf("expected Type to be 'poisoned', got %s", p.Conditions[0].Type)
	}
	if p.Conditions[0].Duration != 3 {
		t.Errorf("expected Duration to be 3, got %d", p.Conditions[0].Duration)
	}
	if p.Conditions[0].Source != "trap" {
		t.Errorf("expected Source to be 'trap', got %s", p.Conditions[0].Source)
	}
}

func TestParticipant_AddCondition_UpdateDuration(t *testing.T) {
	p := &models.Participant{
		CharacterID: "char-001",
		Conditions:  []models.Condition{{Type: "poisoned", Duration: 3, Source: "trap"}},
	}

	// 添加相同状态，持续时间更长
	p.AddCondition("poisoned", 5, "spell")

	if len(p.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(p.Conditions))
	}
	if p.Conditions[0].Duration != 5 {
		t.Errorf("expected Duration to be updated to 5, got %d", p.Conditions[0].Duration)
	}
}

func TestParticipant_AddCondition_Permanent(t *testing.T) {
	p := &models.Participant{
		CharacterID: "char-001",
		Conditions:  []models.Condition{{Type: "poisoned", Duration: 3, Source: "trap"}},
	}

	// 添加相同状态，永久持续
	p.AddCondition("poisoned", -1, "curse")

	if len(p.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(p.Conditions))
	}
	if p.Conditions[0].Duration != -1 {
		t.Errorf("expected Duration to be -1 (permanent), got %d", p.Conditions[0].Duration)
	}
}

func TestParticipant_RemoveCondition(t *testing.T) {
	p := &models.Participant{
		CharacterID: "char-001",
		Conditions: []models.Condition{
			{Type: "poisoned", Duration: 3},
			{Type: "blinded", Duration: 2},
		},
	}

	removed := p.RemoveCondition("poisoned")
	if !removed {
		t.Error("expected condition to be removed")
	}
	if len(p.Conditions) != 1 {
		t.Errorf("expected 1 condition, got %d", len(p.Conditions))
	}
	if p.Conditions[0].Type != "blinded" {
		t.Errorf("expected remaining condition to be 'blinded', got %s", p.Conditions[0].Type)
	}

	// 移除不存在的状态
	removed = p.RemoveCondition("paralyzed")
	if removed {
		t.Error("expected false for non-existent condition")
	}
}

func TestParticipant_HasCondition(t *testing.T) {
	p := &models.Participant{
		CharacterID: "char-001",
		Conditions: []models.Condition{
			{Type: "poisoned", Duration: 3},
		},
	}

	if !p.HasCondition("poisoned") {
		t.Error("expected to have 'poisoned' condition")
	}
	if p.HasCondition("blinded") {
		t.Error("expected not to have 'blinded' condition")
	}
}

func TestParticipant_TickConditions(t *testing.T) {
	p := &models.Participant{
		CharacterID: "char-001",
		Conditions: []models.Condition{
			{Type: "poisoned", Duration: 1},  // 将过期
			{Type: "blinded", Duration: 3},   // 不会过期
			{Type: "cursed", Duration: -1},   // 永久，不会过期
		},
	}

	expired := p.TickConditions()

	if len(expired) != 1 {
		t.Errorf("expected 1 expired condition, got %d", len(expired))
	}
	if expired[0] != "poisoned" {
		t.Errorf("expected 'poisoned' to expire, got %s", expired[0])
	}
	if len(p.Conditions) != 2 {
		t.Errorf("expected 2 remaining conditions, got %d", len(p.Conditions))
	}
}
