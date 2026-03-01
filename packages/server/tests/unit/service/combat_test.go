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

// MockCombatStore is a mock implementation of CombatStore
type MockCombatStore struct {
	mock.Mock
	combatMap map[string]*models.Combat
}

func NewMockCombatStore() *MockCombatStore {
	return &MockCombatStore{
		combatMap: make(map[string]*models.Combat),
	}
}

func (m *MockCombatStore) Create(ctx context.Context, combat *models.Combat) error {
	args := m.Called(ctx, combat)
	if args.Error(0) == nil {
		m.combatMap[combat.ID] = combat
	}
	return args.Error(0)
}

func (m *MockCombatStore) Get(ctx context.Context, id string) (*models.Combat, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Combat), args.Error(1)
}

func (m *MockCombatStore) GetByCampaign(ctx context.Context, campaignID string) ([]*models.Combat, error) {
	args := m.Called(ctx, campaignID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Combat), args.Error(1)
}

func (m *MockCombatStore) GetActive(ctx context.Context, campaignID string) (*models.Combat, error) {
	args := m.Called(ctx, campaignID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Combat), args.Error(1)
}

func (m *MockCombatStore) Update(ctx context.Context, combat *models.Combat) error {
	args := m.Called(ctx, combat)
	if args.Error(0) == nil {
		m.combatMap[combat.ID] = combat
	}
	return args.Error(0)
}

func (m *MockCombatStore) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockCampaignStoreForCombat is a mock implementation of CampaignStoreForCombat
type MockCampaignStoreForCombat struct {
	mock.Mock
}

func (m *MockCampaignStoreForCombat) Get(ctx context.Context, id string) (*models.Campaign, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Campaign), args.Error(1)
}

// MockCharacterStoreForCombat is a mock implementation of CharacterStore for combat tests
type MockCharacterStoreForCombat struct {
	mock.Mock
}

func (m *MockCharacterStoreForCombat) Create(ctx context.Context, character *models.Character) error {
	args := m.Called(ctx, character)
	return args.Error(0)
}

func (m *MockCharacterStoreForCombat) Get(ctx context.Context, id string) (*models.Character, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Character), args.Error(1)
}

func (m *MockCharacterStoreForCombat) GetByCampaignAndID(ctx context.Context, campaignID, id string) (*models.Character, error) {
	args := m.Called(ctx, campaignID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Character), args.Error(1)
}

func (m *MockCharacterStoreForCombat) List(ctx context.Context, filter *store.CharacterFilter) ([]*models.Character, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Character), args.Error(1)
}

func (m *MockCharacterStoreForCombat) Update(ctx context.Context, character *models.Character) error {
	args := m.Called(ctx, character)
	return args.Error(0)
}

func (m *MockCharacterStoreForCombat) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCharacterStoreForCombat) Count(ctx context.Context, filter *store.CharacterFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// MockRandomSourceForCombat is a mock for the dice random source
type MockRandomSourceForCombat struct {
	values []int
	index  int
}

func (m *MockRandomSourceForCombat) Intn(n int) int {
	if m.index >= len(m.values) {
		return 0
	}
	val := m.values[m.index] % n
	m.index++
	return val
}

// Helper functions to create test characters
func createTestCharacter(id, name, campaignID string, hp, ac int) *models.Character {
	return &models.Character{
		ID:         id,
		CampaignID: campaignID,
		Name:       name,
		IsNPC:      false,
		PlayerID:   "player1",
		Level:      5,
		Abilities: &models.Abilities{
			Strength:     16, // +3
			Dexterity:    14, // +2
			Constitution: 14, // +2
			Intelligence: 10, // +0
			Wisdom:       12, // +1
			Charisma:     8,  // -1
		},
		HP:         models.NewHP(hp),
		AC:         ac,
		Speed:      30,
		Initiative: 2, // Dex mod
	}
}

func createTestWeapon() *models.EquipmentItem {
	return &models.EquipmentItem{
		ID:         "weapon1",
		Name:       "Longsword",
		Type:       models.EquipmentTypeWeapon,
		Damage:     "1d8",
		DamageType: "slashing",
		Properties: []string{"versatile"},
	}
}

// TestCombatService_StartCombat tests StartCombat
func TestCombatService_StartCombat(t *testing.T) {
	tests := []struct {
		name          string
		req           *service.StartCombatRequest
		mockSetup     func(*MockCombatStore, *MockCampaignStoreForCombat, *MockCharacterStoreForCombat)
		expectErr     bool
		errContains   string
	}{
		{
			name: "successful combat start",
			req: &service.StartCombatRequest{
				CampaignID:     "campaign1",
				ParticipantIDs: []string{"char1", "char2"},
			},
			mockSetup: func(cs *MockCombatStore, cas *MockCampaignStoreForCombat, chs *MockCharacterStoreForCombat) {
				cas.On("Get", mock.Anything, "campaign1").Return(&models.Campaign{ID: "campaign1"}, nil)
				cs.On("GetActive", mock.Anything, "campaign1").Return(nil, errors.New("not found"))
				chs.On("Get", mock.Anything, "char1").Return(createTestCharacter("char1", "Fighter", "campaign1", 45, 16), nil)
				chs.On("Get", mock.Anything, "char2").Return(createTestCharacter("char2", "Goblin", "campaign1", 7, 15), nil)
				cs.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			expectErr: false,
		},
		{
			name: "missing campaign ID",
			req: &service.StartCombatRequest{
				ParticipantIDs: []string{"char1"},
			},
			mockSetup:   func(cs *MockCombatStore, cas *MockCampaignStoreForCombat, chs *MockCharacterStoreForCombat) {},
			expectErr:   true,
			errContains: "campaign ID is required",
		},
		{
			name: "no participants",
			req: &service.StartCombatRequest{
				CampaignID:     "campaign1",
				ParticipantIDs: []string{},
			},
			mockSetup:   func(cs *MockCombatStore, cas *MockCampaignStoreForCombat, chs *MockCharacterStoreForCombat) {},
			expectErr:   true,
			errContains: "at least one participant",
		},
		{
			name: "campaign not found",
			req: &service.StartCombatRequest{
				CampaignID:     "campaign1",
				ParticipantIDs: []string{"char1"},
			},
			mockSetup: func(cs *MockCombatStore, cas *MockCampaignStoreForCombat, chs *MockCharacterStoreForCombat) {
				cas.On("Get", mock.Anything, "campaign1").Return(nil, errors.New("not found"))
			},
			expectErr:   true,
			errContains: "failed to get campaign",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCombatStore := NewMockCombatStore()
			mockCampaignStore := new(MockCampaignStoreForCombat)
			mockCharacterStore := new(MockCharacterStoreForCombat)
			mockDiceStore := new(MockCharacterStoreForDice)
			mockRandom := &MockRandomSourceForCombat{values: []int{9, 14}} // rolls 10 and 15
			roller := dice.NewRollerWithSource(mockRandom)
			diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

			tt.mockSetup(mockCombatStore, mockCampaignStore, mockCharacterStore)

			svc := service.NewCombatServiceWithRoller(
				mockCombatStore,
				mockCharacterStore,
				mockCampaignStore,
				diceSvc,
				roller,
			)

			combat, err := svc.StartCombat(context.Background(), tt.req)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, combat)
			assert.Equal(t, models.CombatStatusActive, combat.Status)
			assert.Equal(t, 1, combat.Round)
			assert.Len(t, combat.Participants, len(tt.req.ParticipantIDs))
		})
	}
}

// TestCombatService_StartCombat_AlreadyActive tests starting combat when one is already active
func TestCombatService_StartCombat_AlreadyActive(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	mockRandom := &MockRandomSourceForCombat{values: []int{9}}
	roller := dice.NewRollerWithSource(mockRandom)
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	// Setup mocks
	mockCampaignStore.On("Get", mock.Anything, "campaign1").Return(&models.Campaign{ID: "campaign1"}, nil)
	activeCombat := models.NewCombat("campaign1", []string{"char1"})
	mockCombatStore.On("GetActive", mock.Anything, "campaign1").Return(activeCombat, nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	_, err := svc.StartCombat(context.Background(), &service.StartCombatRequest{
		CampaignID:     "campaign1",
		ParticipantIDs: []string{"char2"},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already has an active combat")
}

// TestCombatService_GetCombatState tests GetCombatState
func TestCombatService_GetCombatState(t *testing.T) {
	tests := []struct {
		name        string
		combatID    string
		mockSetup   func(*MockCombatStore)
		expectErr   bool
		errContains string
	}{
		{
			name:     "successful get",
			combatID: "combat1",
			mockSetup: func(cs *MockCombatStore) {
				combat := models.NewCombat("campaign1", []string{"char1"})
				combat.ID = "combat1"
				cs.On("Get", mock.Anything, "combat1").Return(combat, nil)
			},
			expectErr: false,
		},
		{
			name:        "missing combat ID",
			combatID:    "",
			mockSetup:   func(cs *MockCombatStore) {},
			expectErr:   true,
			errContains: "combat ID is required",
		},
		{
			name:     "combat not found",
			combatID: "nonexistent",
			mockSetup: func(cs *MockCombatStore) {
				cs.On("Get", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectErr:   true,
			errContains: "failed to get combat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCombatStore := NewMockCombatStore()
			mockCampaignStore := new(MockCampaignStoreForCombat)
			mockCharacterStore := new(MockCharacterStoreForCombat)
			mockDiceStore := new(MockCharacterStoreForDice)
			roller := dice.NewRoller()
			diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

			tt.mockSetup(mockCombatStore)

			svc := service.NewCombatServiceWithRoller(
				mockCombatStore,
				mockCharacterStore,
				mockCampaignStore,
				diceSvc,
				roller,
			)

			combat, err := svc.GetCombatState(context.Background(), tt.combatID)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, combat)
		})
	}
}

// TestCombatService_GetActiveCombat tests GetActiveCombat
func TestCombatService_GetActiveCombat(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	roller := dice.NewRoller()
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	activeCombat := models.NewCombat("campaign1", []string{"char1"})
	mockCombatStore.On("GetActive", mock.Anything, "campaign1").Return(activeCombat, nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	combat, err := svc.GetActiveCombat(context.Background(), "campaign1")

	assert.NoError(t, err)
	assert.NotNil(t, combat)
	assert.Equal(t, models.CombatStatusActive, combat.Status)
}

// TestCombatService_Attack_Hit tests attack that hits
func TestCombatService_Attack_Hit(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	// Roll 10 for attack (10 + 5 = 15 >= AC 15, hit)
	// Roll 4 for damage (4 + 3 = 7 damage)
	mockRandom := &MockRandomSourceForCombat{values: []int{9, 3}}
	roller := dice.NewRollerWithSource(mockRandom)
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	// Create combat with attacker and target
	combat := models.NewCombat("campaign1", []string{"attacker", "target"})
	combat.ID = "combat1"
	combat.Participants = []models.Participant{
		{CharacterID: "attacker", Initiative: 20, HasActed: false},
		{CharacterID: "target", Initiative: 10, HasActed: false},
	}

	attacker := createTestCharacter("attacker", "Fighter", "campaign1", 45, 16)
	attacker.EquipmentSlots = &models.EquipmentSlots{
		MainHand: createTestWeapon(),
	}
	target := createTestCharacter("target", "Goblin", "campaign1", 7, 15)

	mockCombatStore.On("Get", mock.Anything, "combat1").Return(combat, nil)
	mockCharacterStore.On("Get", mock.Anything, "attacker").Return(attacker, nil)
	mockCharacterStore.On("Get", mock.Anything, "target").Return(target, nil)
	mockCharacterStore.On("Update", mock.Anything, mock.Anything).Return(nil)
	mockCombatStore.On("Update", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	resp, err := svc.Attack(context.Background(), &service.AttackRequest{
		CombatID:   "combat1",
		AttackerID: "attacker",
		TargetID:   "target",
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Result.Hit)
	assert.False(t, resp.Result.Crit)
	assert.Greater(t, resp.Result.Damage, 0)
}

// TestCombatService_Attack_Miss tests attack that misses
func TestCombatService_Attack_Miss(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	// Roll 4 for attack (4 + 5 = 9 < AC 15, miss)
	mockRandom := &MockRandomSourceForCombat{values: []int{3}}
	roller := dice.NewRollerWithSource(mockRandom)
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	combat := models.NewCombat("campaign1", []string{"attacker", "target"})
	combat.ID = "combat1"
	combat.Participants = []models.Participant{
		{CharacterID: "attacker", Initiative: 20, HasActed: false},
		{CharacterID: "target", Initiative: 10, HasActed: false},
	}

	attacker := createTestCharacter("attacker", "Fighter", "campaign1", 45, 16)
	attacker.EquipmentSlots = &models.EquipmentSlots{
		MainHand: createTestWeapon(),
	}
	target := createTestCharacter("target", "Goblin", "campaign1", 7, 15)

	mockCombatStore.On("Get", mock.Anything, "combat1").Return(combat, nil)
	mockCharacterStore.On("Get", mock.Anything, "attacker").Return(attacker, nil)
	mockCharacterStore.On("Get", mock.Anything, "target").Return(target, nil)
	mockCombatStore.On("Update", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	resp, err := svc.Attack(context.Background(), &service.AttackRequest{
		CombatID:   "combat1",
		AttackerID: "attacker",
		TargetID:   "target",
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Result.Hit)
	assert.False(t, resp.Result.Crit)
	assert.Equal(t, 0, resp.Result.Damage)
}

// TestCombatService_Attack_Crit tests critical hit
func TestCombatService_Attack_Crit(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	// Roll 19 for attack (natural 20, auto crit)
	// Roll 3, 5 for damage (crit: 2d8 = 8 + 3 = 11 damage)
	mockRandom := &MockRandomSourceForCombat{values: []int{19, 2, 4}}
	roller := dice.NewRollerWithSource(mockRandom)
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	combat := models.NewCombat("campaign1", []string{"attacker", "target"})
	combat.ID = "combat1"
	combat.Participants = []models.Participant{
		{CharacterID: "attacker", Initiative: 20, HasActed: false},
		{CharacterID: "target", Initiative: 10, HasActed: false},
	}

	attacker := createTestCharacter("attacker", "Fighter", "campaign1", 45, 16)
	attacker.EquipmentSlots = &models.EquipmentSlots{
		MainHand: createTestWeapon(),
	}
	target := createTestCharacter("target", "Goblin", "campaign1", 7, 15)

	mockCombatStore.On("Get", mock.Anything, "combat1").Return(combat, nil)
	mockCharacterStore.On("Get", mock.Anything, "attacker").Return(attacker, nil)
	mockCharacterStore.On("Get", mock.Anything, "target").Return(target, nil)
	mockCharacterStore.On("Update", mock.Anything, mock.Anything).Return(nil)
	mockCombatStore.On("Update", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	resp, err := svc.Attack(context.Background(), &service.AttackRequest{
		CombatID:   "combat1",
		AttackerID: "attacker",
		TargetID:   "target",
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Result.Hit)
	assert.True(t, resp.Result.Crit)
	assert.Greater(t, resp.Result.Damage, 0)
}

// TestCombatService_Attack_Fumble tests natural 1 (fumble)
func TestCombatService_Attack_Fumble(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	// Roll 0 for attack (natural 1, auto miss)
	mockRandom := &MockRandomSourceForCombat{values: []int{0}}
	roller := dice.NewRollerWithSource(mockRandom)
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	combat := models.NewCombat("campaign1", []string{"attacker", "target"})
	combat.ID = "combat1"
	combat.Participants = []models.Participant{
		{CharacterID: "attacker", Initiative: 20, HasActed: false},
		{CharacterID: "target", Initiative: 10, HasActed: false},
	}

	attacker := createTestCharacter("attacker", "Fighter", "campaign1", 45, 16)
	attacker.EquipmentSlots = &models.EquipmentSlots{
		MainHand: createTestWeapon(),
	}
	target := createTestCharacter("target", "Goblin", "campaign1", 7, 15)

	mockCombatStore.On("Get", mock.Anything, "combat1").Return(combat, nil)
	mockCharacterStore.On("Get", mock.Anything, "attacker").Return(attacker, nil)
	mockCharacterStore.On("Get", mock.Anything, "target").Return(target, nil)
	mockCombatStore.On("Update", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	resp, err := svc.Attack(context.Background(), &service.AttackRequest{
		CombatID:   "combat1",
		AttackerID: "attacker",
		TargetID:   "target",
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Result.Hit)
	assert.False(t, resp.Result.Crit)
	assert.Equal(t, 0, resp.Result.Damage)
}

// TestCombatService_CastSpell tests casting a spell
func TestCombatService_CastSpell(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	// Roll 3, 4 for damage (2d6 = 7 damage)
	mockRandom := &MockRandomSourceForCombat{values: []int{2, 3}}
	roller := dice.NewRollerWithSource(mockRandom)
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	combat := models.NewCombat("campaign1", []string{"caster", "target"})
	combat.ID = "combat1"
	combat.Participants = []models.Participant{
		{CharacterID: "caster", Initiative: 20, HasActed: false},
		{CharacterID: "target", Initiative: 10, HasActed: false},
	}

	caster := createTestCharacter("caster", "Wizard", "campaign1", 24, 12)
	caster.Class = "Wizard"
	target := createTestCharacter("target", "Goblin", "campaign1", 7, 15)

	mockCombatStore.On("Get", mock.Anything, "combat1").Return(combat, nil)
	mockCharacterStore.On("Get", mock.Anything, "caster").Return(caster, nil)
	mockCharacterStore.On("Get", mock.Anything, "target").Return(target, nil)
	mockCharacterStore.On("Update", mock.Anything, mock.Anything).Return(nil)
	mockCombatStore.On("Update", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	resp, err := svc.CastSpell(context.Background(), &service.CastSpellRequest{
		CombatID:   "combat1",
		CasterID:   "caster",
		SpellID:    "fireball",
		SpellName:  "Fireball",
		TargetIDs:  []string{"target"},
		Level:      3,
		Damage:     "2d6",
		DamageType: "fire",
		IsHealing:  false,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Result)
	assert.Equal(t, "Fireball", resp.Result.SpellName)
	assert.Len(t, resp.Result.Results, 1)
	assert.Greater(t, resp.Result.Damage, 0)
}

// TestCombatService_CastSpell_Healing tests casting a healing spell
func TestCombatService_CastSpell_Healing(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	// Roll 4 for healing (1d4+3 = 7 healing)
	mockRandom := &MockRandomSourceForCombat{values: []int{3}}
	roller := dice.NewRollerWithSource(mockRandom)
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	combat := models.NewCombat("campaign1", []string{"caster", "target"})
	combat.ID = "combat1"
	combat.Participants = []models.Participant{
		{CharacterID: "caster", Initiative: 20, HasActed: false},
		{CharacterID: "target", Initiative: 10, HasActed: false},
	}

	caster := createTestCharacter("caster", "Cleric", "campaign1", 30, 16)
	caster.Class = "Cleric"
	target := createTestCharacter("target", "Fighter", "campaign1", 20, 16) // damaged fighter
	target.HP.Current = 10 // has taken damage

	mockCombatStore.On("Get", mock.Anything, "combat1").Return(combat, nil)
	mockCharacterStore.On("Get", mock.Anything, "caster").Return(caster, nil)
	mockCharacterStore.On("Get", mock.Anything, "target").Return(target, nil)
	mockCharacterStore.On("Update", mock.Anything, mock.Anything).Return(nil)
	mockCombatStore.On("Update", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	resp, err := svc.CastSpell(context.Background(), &service.CastSpellRequest{
		CombatID:   "combat1",
		CasterID:   "caster",
		SpellID:    "cure_wounds",
		SpellName:  "Cure Wounds",
		TargetIDs:  []string{"target"},
		Level:      1,
		Damage:     "1d4",
		DamageType: "",
		IsHealing:  true,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Result.IsHealing)
	assert.Greater(t, resp.Result.Damage, 0)
}

// TestCombatService_AdvanceTurn tests advancing turn
func TestCombatService_AdvanceTurn(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	roller := dice.NewRoller()
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	combat := models.NewCombat("campaign1", []string{"char1", "char2"})
	combat.ID = "combat1"
	combat.Participants = []models.Participant{
		{CharacterID: "char1", Initiative: 20, HasActed: false},
		{CharacterID: "char2", Initiative: 10, HasActed: false},
	}

	char1 := createTestCharacter("char1", "Fighter", "campaign1", 45, 16)

	mockCombatStore.On("Get", mock.Anything, "combat1").Return(combat, nil)
	mockCharacterStore.On("Get", mock.Anything, "char2").Return(char1, nil)
	mockCombatStore.On("Update", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	resp, err := svc.AdvanceTurn(context.Background(), "combat1")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, resp.Combat.TurnIndex) // moved to second participant
	assert.False(t, resp.NewRound)
}

// TestCombatService_AdvanceTurn_NewRound tests advancing to new round
func TestCombatService_AdvanceTurn_NewRound(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	roller := dice.NewRoller()
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	combat := models.NewCombat("campaign1", []string{"char1", "char2"})
	combat.ID = "combat1"
	combat.Participants = []models.Participant{
		{CharacterID: "char1", Initiative: 20, HasActed: true},
		{CharacterID: "char2", Initiative: 10, HasActed: false},
	}
	combat.TurnIndex = 1 // Last participant's turn

	char1 := createTestCharacter("char1", "Fighter", "campaign1", 45, 16)

	mockCombatStore.On("Get", mock.Anything, "combat1").Return(combat, nil)
	mockCharacterStore.On("Get", mock.Anything, "char1").Return(char1, nil)
	mockCombatStore.On("Update", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	resp, err := svc.AdvanceTurn(context.Background(), "combat1")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.NewRound)
	assert.Equal(t, 2, resp.Combat.Round)
	assert.Equal(t, 0, resp.Combat.TurnIndex)
}

// TestCombatService_EndCombat tests ending combat
func TestCombatService_EndCombat(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	roller := dice.NewRoller()
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	combat := models.NewCombat("campaign1", []string{"char1"})
	combat.ID = "combat1"

	mockCombatStore.On("Get", mock.Anything, "combat1").Return(combat, nil)
	mockCombatStore.On("Update", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	updatedCombat, err := svc.EndCombat(context.Background(), "combat1")

	assert.NoError(t, err)
	assert.NotNil(t, updatedCombat)
	assert.Equal(t, models.CombatStatusFinished, updatedCombat.Status)
	assert.NotNil(t, updatedCombat.EndedAt)
}

// TestCombatService_EndCombatWithSummary tests ending combat with summary
func TestCombatService_EndCombatWithSummary(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	roller := dice.NewRoller()
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	combat := models.NewCombat("campaign1", []string{"char1", "char2"})
	combat.ID = "combat1"
	combat.Participants = []models.Participant{
		{CharacterID: "char1", Initiative: 20, HasActed: true},
		{CharacterID: "char2", Initiative: 10, HasActed: false},
	}
	combat.Round = 3
	// Add some combat log entries
	combat.AddLogEntry("char1", "attack", "char2", "roll 15 vs AC 15, hit for 8 damage")
	combat.AddLogEntry("char2", "attack", "char1", "roll 12 vs AC 16, miss")
	combat.AddLogEntry("char1", "critical_hit", "char2", "roll 20 vs AC 15, hit for 16 damage")

	char1 := createTestCharacter("char1", "Fighter", "campaign1", 45, 16)
	char2 := createTestCharacter("char2", "Goblin", "campaign1", 7, 15)

	mockCombatStore.On("Get", mock.Anything, "combat1").Return(combat, nil)
	mockCharacterStore.On("Get", mock.Anything, "char1").Return(char1, nil)
	mockCharacterStore.On("Get", mock.Anything, "char2").Return(char2, nil)
	mockCombatStore.On("Update", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	resp, err := svc.EndCombatWithSummary(context.Background(), "combat1")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Combat)
	assert.NotNil(t, resp.Summary)
	assert.Equal(t, models.CombatStatusFinished, resp.Combat.Status)
	assert.Equal(t, 3, resp.Summary.TotalRounds)
	assert.Len(t, resp.Summary.Participants, 2)

	// Verify participant summaries
	participantMap := make(map[string]service.ParticipantSummary)
	for _, p := range resp.Summary.Participants {
		participantMap[p.CharacterID] = p
	}

	// char1 dealt 24 damage (8 + 16)
	assert.Equal(t, 24, participantMap["char1"].DamageDealt)
	// char2 took 24 damage
	assert.Equal(t, 24, participantMap["char2"].DamageTaken)
}

// TestCombatService_EndCombatWithSummary_EmptyLog tests ending combat with empty log
func TestCombatService_EndCombatWithSummary_EmptyLog(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	roller := dice.NewRoller()
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	combat := models.NewCombat("campaign1", []string{"char1"})
	combat.ID = "combat1"

	char1 := createTestCharacter("char1", "Fighter", "campaign1", 45, 16)

	mockCombatStore.On("Get", mock.Anything, "combat1").Return(combat, nil)
	mockCharacterStore.On("Get", mock.Anything, "char1").Return(char1, nil)
	mockCombatStore.On("Update", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	resp, err := svc.EndCombatWithSummary(context.Background(), "combat1")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Summary)
	assert.Len(t, resp.Summary.Participants, 1)
	assert.Equal(t, 0, resp.Summary.Participants[0].DamageDealt)
	assert.Equal(t, 0, resp.Summary.Participants[0].DamageTaken)
}

// TestCombatService_EndCombatWithSummary_InvalidID tests ending combat with invalid ID
func TestCombatService_EndCombatWithSummary_InvalidID(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	roller := dice.NewRoller()
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	_, err := svc.EndCombatWithSummary(context.Background(), "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "combat ID is required")
}

// TestCombatService_EndCombatWithSummary_AlreadyEnded tests ending combat that is already finished
func TestCombatService_EndCombatWithSummary_AlreadyEnded(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	roller := dice.NewRoller()
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	combat := models.NewCombat("campaign1", []string{"char1"})
	combat.ID = "combat1"
	combat.End() // Already ended

	mockCombatStore.On("Get", mock.Anything, "combat1").Return(combat, nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	_, err := svc.EndCombatWithSummary(context.Background(), "combat1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "combat is not active")
}

// TestParseDamageFromResult tests the damage parsing function
func TestParseDamageFromResult(t *testing.T) {
	tests := []struct {
		name     string
		result   string
		expected int
	}{
		{"hit for 8 damage", "roll 15 vs AC 15, hit for 8 damage", 8},
		{"hit for 16 damage", "roll 20 vs AC 15, hit for 16 damage", 16},
		{"dealt 10 fire damage", "cast Fireball, dealt 10 fire damage", 10},
		{"no damage", "roll 12 vs AC 16, miss", 0},
		{"dealt 25 cold damage", "cast Ice Storm, dealt 25 cold damage", 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			damage := service.ParseDamageFromResult(tt.result)
			assert.Equal(t, tt.expected, damage)
		})
	}
}

// TestParseHealingFromResult tests the healing parsing function
func TestParseHealingFromResult(t *testing.T) {
	tests := []struct {
		name     string
		result   string
		expected int
	}{
		{"healed 7", "cast Cure Wounds, healed 7", 7},
		{"healed 15", "cast Healing Word, healed 15", 15},
		{"no healing", "cast Fireball, dealt 10 fire damage", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			healing := service.ParseHealingFromResult(tt.result)
			assert.Equal(t, tt.expected, healing)
		})
	}
}
