package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/rules/dice"
	"github.com/dnd-mcp/server/internal/service"
	"github.com/dnd-mcp/server/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCharacterStoreForDice is a mock implementation of CharacterStore for dice tests
type MockCharacterStoreForDice struct {
	mock.Mock
}

func (m *MockCharacterStoreForDice) Create(ctx context.Context, character *models.Character) error {
	args := m.Called(ctx, character)
	return args.Error(0)
}

func (m *MockCharacterStoreForDice) Get(ctx context.Context, id string) (*models.Character, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Character), args.Error(1)
}

func (m *MockCharacterStoreForDice) GetByCampaignAndID(ctx context.Context, campaignID, id string) (*models.Character, error) {
	args := m.Called(ctx, campaignID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Character), args.Error(1)
}

func (m *MockCharacterStoreForDice) List(ctx context.Context, filter *store.CharacterFilter) ([]*models.Character, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Character), args.Error(1)
}

func (m *MockCharacterStoreForDice) Update(ctx context.Context, character *models.Character) error {
	args := m.Called(ctx, character)
	return args.Error(0)
}

func (m *MockCharacterStoreForDice) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCharacterStoreForDice) Count(ctx context.Context, filter *store.CharacterFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// MockRandomSourceForService is a mock for the dice random source
type MockRandomSourceForService struct {
	values []int
	index  int
}

func (m *MockRandomSourceForService) Intn(n int) int {
	if m.index >= len(m.values) {
		return 0
	}
	val := m.values[m.index] % n
	m.index++
	return val
}

func TestDiceService_RollDice(t *testing.T) {
	tests := []struct {
		name        string
		req         *service.RollDiceRequest
		mockValues  []int
		expectTotal int
		expectErr   bool
		errContains string
	}{
		{
			name:        "basic 1d20",
			req:         &service.RollDiceRequest{Formula: "1d20"},
			mockValues:  []int{9}, // roll 10
			expectTotal: 10,
			expectErr:   false,
		},
		{
			name:        "1d20+5",
			req:         &service.RollDiceRequest{Formula: "1d20+5"},
			mockValues:  []int{9}, // roll 10
			expectTotal: 15,
			expectErr:   false,
		},
		{
			name:        "empty formula",
			req:         &service.RollDiceRequest{Formula: ""},
			mockValues:  nil,
			expectErr:   true,
			errContains: "formula is required",
		},
		{
			name:        "invalid formula",
			req:         &service.RollDiceRequest{Formula: "invalid"},
			mockValues:  nil,
			expectErr:   true,
			errContains: "invalid formula",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockCharacterStoreForDice)
			mockRandom := &MockRandomSourceForService{values: tt.mockValues}
			roller := dice.NewRollerWithSource(mockRandom)
			svc := service.NewDiceServiceWithRoller(mockStore, roller)

			resp, err := svc.RollDice(context.Background(), tt.req)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.expectTotal, resp.Result.Total)
		})
	}
}

func TestDiceService_RollDice_Advantage(t *testing.T) {
	mockStore := new(MockCharacterStoreForDice)
	// Mock returns 4 and 14 -> rolls 5 and 15, keep 15 with advantage
	mockRandom := &MockRandomSourceForService{values: []int{4, 14}}
	roller := dice.NewRollerWithSource(mockRandom)
	svc := service.NewDiceServiceWithRoller(mockStore, roller)

	resp, err := svc.RollDice(context.Background(), &service.RollDiceRequest{
		Formula: "1d20 adv",
	})
	assert.NoError(t, err)
	assert.Equal(t, 15, resp.Result.Rolls[0])
	assert.Equal(t, 15, resp.Result.Total)
}

func TestDiceService_RollDice_Disadvantage(t *testing.T) {
	mockStore := new(MockCharacterStoreForDice)
	// Mock returns 14 and 4 -> rolls 15 and 5, keep 5 with disadvantage
	mockRandom := &MockRandomSourceForService{values: []int{14, 4}}
	roller := dice.NewRollerWithSource(mockRandom)
	svc := service.NewDiceServiceWithRoller(mockStore, roller)

	resp, err := svc.RollDice(context.Background(), &service.RollDiceRequest{
		Formula: "1d20 dis",
	})
	assert.NoError(t, err)
	assert.Equal(t, 5, resp.Result.Rolls[0])
	assert.Equal(t, 5, resp.Result.Total)
}

func TestDiceService_RollCheck(t *testing.T) {
	// Create test character with STR 16 (+3 modifier)
	character := models.NewCharacter("campaign-1", "Test Character", false)
	character.ID = "char-1"
	character.Abilities = &models.Abilities{
		Strength:     16, // +3
		Dexterity:    14, // +2
		Constitution: 12, // +1
		Intelligence: 10, // 0
		Wisdom:       8,  // -1
		Charisma:     6,  // -2
	}
	character.Skills = map[string]int{
		"athletics": 5, // +3 STR + 2 proficiency
	}

	tests := []struct {
		name          string
		req           *service.RollCheckRequest
		mockValues    []int
		expectTotal   int
		expectSuccess bool
		expectErr     bool
		errContains   string
	}{
		{
			name:        "strength check",
			req:         &service.RollCheckRequest{CharacterID: "char-1", Ability: "strength"},
			mockValues:  []int{9}, // roll 10
			expectTotal: 13,       // 10 + 3 STR modifier
		},
		{
			name:        "strength check with athletics skill",
			req:         &service.RollCheckRequest{CharacterID: "char-1", Ability: "strength", Skill: "athletics"},
			mockValues:  []int{9}, // roll 10
			expectTotal: 15,       // 10 + 5 athletics bonus
		},
		{
			name:          "strength check against DC 15",
			req:           &service.RollCheckRequest{CharacterID: "char-1", Ability: "strength", DC: 15},
			mockValues:    []int{9}, // roll 10
			expectTotal:   13,
			expectSuccess: false, // 13 < 15
		},
		{
			name:          "strength check against DC 10",
			req:           &service.RollCheckRequest{CharacterID: "char-1", Ability: "strength", DC: 10},
			mockValues:    []int{9}, // roll 10
			expectTotal:   13,
			expectSuccess: true, // 13 >= 10
		},
		{
			name:        "empty character_id",
			req:         &service.RollCheckRequest{CharacterID: "", Ability: "strength"},
			mockValues:  nil,
			expectErr:   true,
			errContains: "character_id is required",
		},
		{
			name:        "empty ability",
			req:         &service.RollCheckRequest{CharacterID: "char-1", Ability: ""},
			mockValues:  nil,
			expectErr:   true,
			errContains: "ability is required",
		},
		{
			name:        "invalid ability",
			req:         &service.RollCheckRequest{CharacterID: "char-1", Ability: "invalid"},
			mockValues:  nil,
			expectErr:   true,
			errContains: "invalid ability",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockCharacterStoreForDice)
			mockStore.On("Get", mock.Anything, "char-1").Return(character, nil)

			mockRandom := &MockRandomSourceForService{values: tt.mockValues}
			roller := dice.NewRollerWithSource(mockRandom)
			svc := service.NewDiceServiceWithRoller(mockStore, roller)

			resp, err := svc.RollCheck(context.Background(), tt.req)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.expectTotal, resp.Result.DiceResult.Total)
			if tt.req.DC > 0 {
				assert.Equal(t, tt.expectSuccess, resp.Result.Success)
			}
		})
	}
}

func TestDiceService_RollCheck_AdvantageDisadvantage(t *testing.T) {
	character := models.NewCharacter("campaign-1", "Test Character", false)
	character.ID = "char-1"
	character.Abilities = &models.Abilities{Strength: 16} // +3

	mockStore := new(MockCharacterStoreForDice)
	mockStore.On("Get", mock.Anything, "char-1").Return(character, nil)

	t.Run("advantage takes higher", func(t *testing.T) {
		// Mock returns 4 and 14 -> rolls 5 and 15, keep 15
		mockRandom := &MockRandomSourceForService{values: []int{4, 14}}
		roller := dice.NewRollerWithSource(mockRandom)
		svc := service.NewDiceServiceWithRoller(mockStore, roller)

		resp, err := svc.RollCheck(context.Background(), &service.RollCheckRequest{
			CharacterID: "char-1",
			Ability:     "strength",
			Advantage:   true,
		})
		assert.NoError(t, err)
		assert.Equal(t, 15, resp.Result.DiceResult.Rolls[0])
		assert.Equal(t, 18, resp.Result.DiceResult.Total) // 15 + 3
	})

	t.Run("disadvantage takes lower", func(t *testing.T) {
		// Mock returns 14 and 4 -> rolls 15 and 5, keep 5
		mockRandom := &MockRandomSourceForService{values: []int{14, 4}}
		roller := dice.NewRollerWithSource(mockRandom)
		svc := service.NewDiceServiceWithRoller(mockStore, roller)

		resp, err := svc.RollCheck(context.Background(), &service.RollCheckRequest{
			CharacterID:  "char-1",
			Ability:      "strength",
			Disadvantage: true,
		})
		assert.NoError(t, err)
		assert.Equal(t, 5, resp.Result.DiceResult.Rolls[0])
		assert.Equal(t, 8, resp.Result.DiceResult.Total) // 5 + 3
	})

	t.Run("advantage and disadvantage cancel out", func(t *testing.T) {
		// Normal roll when both are true
		mockRandom := &MockRandomSourceForService{values: []int{9}} // roll 10
		roller := dice.NewRollerWithSource(mockRandom)
		svc := service.NewDiceServiceWithRoller(mockStore, roller)

		resp, err := svc.RollCheck(context.Background(), &service.RollCheckRequest{
			CharacterID:  "char-1",
			Ability:      "strength",
			Advantage:    true,
			Disadvantage: true,
		})
		assert.NoError(t, err)
		assert.Equal(t, 10, resp.Result.DiceResult.Rolls[0])
		assert.Equal(t, 13, resp.Result.DiceResult.Total) // 10 + 3
	})
}

func TestDiceService_RollCheck_CharacterNotFound(t *testing.T) {
	mockStore := new(MockCharacterStoreForDice)
	mockStore.On("Get", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))

	mockRandom := &MockRandomSourceForService{values: []int{9}}
	roller := dice.NewRollerWithSource(mockRandom)
	svc := service.NewDiceServiceWithRoller(mockStore, roller)

	resp, err := svc.RollCheck(context.Background(), &service.RollCheckRequest{
		CharacterID: "nonexistent",
		Ability:     "strength",
	})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestDiceService_RollSave(t *testing.T) {
	// Create test character
	character := models.NewCharacter("campaign-1", "Test Character", false)
	character.ID = "char-1"
	character.Abilities = &models.Abilities{
		Strength:     16, // +3
		Dexterity:    14, // +2
		Constitution: 12, // +1
	}
	character.Saves = map[string]int{
		"strength": 5, // +3 STR + 2 proficiency
	}

	tests := []struct {
		name          string
		req           *service.RollSaveRequest
		mockValues    []int
		expectTotal   int
		expectSuccess bool
		expectErr     bool
		errContains   string
	}{
		{
			name:        "strength save with proficiency",
			req:         &service.RollSaveRequest{CharacterID: "char-1", Ability: "strength"},
			mockValues:  []int{9}, // roll 10
			expectTotal: 15,       // 10 + 5 save bonus
		},
		{
			name:        "dexterity save without proficiency",
			req:         &service.RollSaveRequest{CharacterID: "char-1", Ability: "dexterity"},
			mockValues:  []int{9}, // roll 10
			expectTotal: 12,       // 10 + 2 DEX modifier
		},
		{
			name:          "save against DC",
			req:           &service.RollSaveRequest{CharacterID: "char-1", Ability: "strength", DC: 12},
			mockValues:    []int{9}, // roll 10
			expectTotal:   15,
			expectSuccess: true,
		},
		{
			name:        "empty character_id",
			req:         &service.RollSaveRequest{CharacterID: "", Ability: "strength"},
			mockValues:  nil,
			expectErr:   true,
			errContains: "character_id is required",
		},
		{
			name:        "empty ability",
			req:         &service.RollSaveRequest{CharacterID: "char-1", Ability: ""},
			mockValues:  nil,
			expectErr:   true,
			errContains: "ability is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockCharacterStoreForDice)
			mockStore.On("Get", mock.Anything, "char-1").Return(character, nil)

			mockRandom := &MockRandomSourceForService{values: tt.mockValues}
			roller := dice.NewRollerWithSource(mockRandom)
			svc := service.NewDiceServiceWithRoller(mockStore, roller)

			resp, err := svc.RollSave(context.Background(), tt.req)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.expectTotal, resp.Result.DiceResult.Total)
			if tt.req.DC > 0 {
				assert.Equal(t, tt.expectSuccess, resp.Result.Success)
			}
		})
	}
}

func TestDiceService_RollAttack(t *testing.T) {
	mockRandom := &MockRandomSourceForService{values: []int{9}} // roll 10
	roller := dice.NewRollerWithSource(mockRandom)
	svc := service.NewDiceServiceWithRoller(nil, roller)

	tests := []struct {
		name         string
		req          *service.RollAttackRequest
		expectHit    bool
		expectCrit   bool
		expectErr    bool
		errContains  string
	}{
		{
			name:       "hit AC 12 with +5",
			req:        &service.RollAttackRequest{AttackBonus: 5, TargetAC: 12},
			expectHit:  true,  // 10 + 5 = 15 >= 12
			expectCrit: false, // not natural 20
		},
		{
			name:       "miss AC 18 with +5",
			req:        &service.RollAttackRequest{AttackBonus: 5, TargetAC: 18},
			expectHit:  false, // 10 + 5 = 15 < 18
			expectCrit: false,
		},
		{
			name:       "negative target AC",
			req:        &service.RollAttackRequest{AttackBonus: 5, TargetAC: -1},
			expectErr:  true,
			errContains: "target_ac cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.RollAttack(context.Background(), tt.req)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectHit, resp.Hit)
			assert.Equal(t, tt.expectCrit, resp.Crit)
		})
	}
}

func TestDiceService_RollAttack_CriticalHit(t *testing.T) {
	// Natural 20
	mockRandom := &MockRandomSourceForService{values: []int{19}} // roll 20
	roller := dice.NewRollerWithSource(mockRandom)
	svc := service.NewDiceServiceWithRoller(nil, roller)

	resp, err := svc.RollAttack(context.Background(), &service.RollAttackRequest{
		AttackBonus: 5,
		TargetAC:    30, // Even against impossible AC, natural 20 hits
	})
	assert.NoError(t, err)
	assert.True(t, resp.Hit)
	assert.True(t, resp.Crit)
}

func TestDiceService_RollAttack_CriticalFumble(t *testing.T) {
	// Natural 1
	mockRandom := &MockRandomSourceForService{values: []int{0}} // roll 1
	roller := dice.NewRollerWithSource(mockRandom)
	svc := service.NewDiceServiceWithRoller(nil, roller)

	resp, err := svc.RollAttack(context.Background(), &service.RollAttackRequest{
		AttackBonus: 20, // Even with huge bonus
		TargetAC:    5,  // Even against trivial AC
	})
	assert.NoError(t, err)
	assert.False(t, resp.Hit)
	assert.False(t, resp.Crit)
}

func TestDiceService_RollDamage(t *testing.T) {
	mockRandom := &MockRandomSourceForService{values: []int{2, 3}} // rolls 3, 4
	roller := dice.NewRollerWithSource(mockRandom)
	svc := service.NewDiceServiceWithRoller(nil, roller)

	tests := []struct {
		name        string
		req         *service.RollDamageRequest
		expectTotal int
		expectErr   bool
		errContains string
	}{
		{
			name:        "2d6+3",
			req:         &service.RollDamageRequest{Formula: "2d6+3", Crit: false},
			expectTotal: 10, // 3 + 4 + 3
		},
		{
			name:        "empty formula",
			req:         &service.RollDamageRequest{Formula: "", Crit: false},
			expectErr:   true,
			errContains: "formula is required",
		},
		{
			name:        "invalid formula",
			req:         &service.RollDamageRequest{Formula: "invalid", Crit: false},
			expectErr:   true,
			errContains: "invalid formula",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.RollDamage(context.Background(), tt.req)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectTotal, resp.Result.Total)
		})
	}
}

func TestDiceService_RollDamage_Critical(t *testing.T) {
	// For crit: 2d6 becomes 4d6
	// Mock returns 0, 1, 2, 3 -> rolls 1, 2, 3, 4
	mockRandom := &MockRandomSourceForService{values: []int{0, 1, 2, 3}}
	roller := dice.NewRollerWithSource(mockRandom)
	svc := service.NewDiceServiceWithRoller(nil, roller)

	resp, err := svc.RollDamage(context.Background(), &service.RollDamageRequest{
		Formula: "2d6+3",
		Crit:    true,
	})
	assert.NoError(t, err)
	// 4d6 = 1+2+3+4 = 10, +3 modifier = 13
	assert.Equal(t, 13, resp.Result.Total)
	assert.Len(t, resp.Result.Rolls, 4)
}

func TestDiceService_RollCheck_DetailedSkill(t *testing.T) {
	// Create character with detailed skill
	character := models.NewCharacter("campaign-1", "Test Character", false)
	character.ID = "char-1"
	character.Abilities = &models.Abilities{Strength: 16} // +3
	character.Proficiency = 2
	character.SkillsDetail = map[string]*models.Skill{
		"athletics": {
			Ability:    "strength",
			Proficient: true,
		},
	}

	mockStore := new(MockCharacterStoreForDice)
	mockStore.On("Get", mock.Anything, "char-1").Return(character, nil)

	// Mock returns 9 -> roll 10
	mockRandom := &MockRandomSourceForService{values: []int{9}}
	roller := dice.NewRollerWithSource(mockRandom)
	svc := service.NewDiceServiceWithRoller(mockStore, roller)

	resp, err := svc.RollCheck(context.Background(), &service.RollCheckRequest{
		CharacterID: "char-1",
		Ability:     "strength",
		Skill:       "athletics",
	})
	assert.NoError(t, err)
	// 10 + 3 (STR) + 2 (proficiency) = 15
	assert.Equal(t, 15, resp.Result.DiceResult.Total)
}

func TestDiceService_RollSave_DetailedSave(t *testing.T) {
	// Create character with detailed save
	character := models.NewCharacter("campaign-1", "Test Character", false)
	character.ID = "char-1"
	character.Abilities = &models.Abilities{Constitution: 14} // +2
	character.Proficiency = 2
	character.SavesDetail = map[string]*models.Save{
		"constitution": {
			Proficient: true,
		},
	}

	mockStore := new(MockCharacterStoreForDice)
	mockStore.On("Get", mock.Anything, "char-1").Return(character, nil)

	// Mock returns 9 -> roll 10
	mockRandom := &MockRandomSourceForService{values: []int{9}}
	roller := dice.NewRollerWithSource(mockRandom)
	svc := service.NewDiceServiceWithRoller(mockStore, roller)

	resp, err := svc.RollSave(context.Background(), &service.RollSaveRequest{
		CharacterID: "char-1",
		Ability:     "constitution",
	})
	assert.NoError(t, err)
	// 10 + 2 (CON) + 2 (proficiency) = 14
	assert.Equal(t, 14, resp.Result.DiceResult.Total)
}

func TestGetSkillDCGuide(t *testing.T) {
	guide := service.GetSkillDCGuide()
	assert.Equal(t, 5, guide["very_easy"])
	assert.Equal(t, 10, guide["easy"])
	assert.Equal(t, 15, guide["medium"])
	assert.Equal(t, 20, guide["hard"])
	assert.Equal(t, 25, guide["very_hard"])
	assert.Equal(t, 30, guide["nearly_impossible"])
}
