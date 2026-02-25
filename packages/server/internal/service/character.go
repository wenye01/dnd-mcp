// Package service provides business logic layer implementations
package service

import (
	"context"
	"fmt"
	"math"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/store"
	"github.com/google/uuid"
)

// CharacterStore defines the interface for character data operations
type CharacterStore interface {
	Create(ctx context.Context, character *models.Character) error
	Get(ctx context.Context, id string) (*models.Character, error)
	GetByCampaignAndID(ctx context.Context, campaignID, id string) (*models.Character, error)
	List(ctx context.Context, filter *store.CharacterFilter) ([]*models.Character, error)
	Update(ctx context.Context, character *models.Character) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context, filter *store.CharacterFilter) (int64, error)
}

// CharacterService provides character business logic
// 规则参考: PHB 第7章 Ability Scores and Modifiers, 第9章 Combat
type CharacterService struct {
	store CharacterStore
}

// NewCharacterService creates a new character service
func NewCharacterService(store CharacterStore) *CharacterService {
	return &CharacterService{
		store: store,
	}
}

// CreateCharacterRequest represents a character creation request
type CreateCharacterRequest struct {
	CampaignID string            `json:"campaign_id"`
	Name       string            `json:"name"`
	IsNPC      bool              `json:"is_npc"`
	NPCType    models.NPCType    `json:"npc_type"`
	PlayerID   string            `json:"player_id"`

	// 基础属性
	Race       string            `json:"race"`
	Class      string            `json:"class"`
	Level      int               `json:"level"`
	Background string            `json:"background"`
	Alignment  string            `json:"alignment"`

	// 属性值（可选，未提供时使用标准数组）
	Abilities  *models.Abilities `json:"abilities"`

	// 战斗属性（可选）
	HP         *models.HP        `json:"hp"`
	AC         int               `json:"ac"`
	Speed      int               `json:"speed"`
	Initiative int               `json:"initiative"`

	// 技能和豁免
	Skills     map[string]int    `json:"skills"`
	Saves      map[string]int    `json:"saves"`
}

// UpdateCharacterRequest represents a character update request
type UpdateCharacterRequest struct {
	Name       *string            `json:"name"`
	Race       *string            `json:"race"`
	Class      *string            `json:"class"`
	Level      *int               `json:"level"`
	Background *string            `json:"background"`
	Alignment  *string            `json:"alignment"`
	Abilities  *models.Abilities  `json:"abilities"`
	HP         *models.HP         `json:"hp"`
	AC         *int               `json:"ac"`
	Speed      *int               `json:"speed"`
	Initiative *int               `json:"initiative"`
	Skills     map[string]int     `json:"skills"`
	Saves      map[string]int     `json:"saves"`
}

// HPChangeRequest represents an HP change request
type HPChangeRequest struct {
	Damage     int  `json:"damage"`
	Healing    int  `json:"healing"`
	TempHP     int  `json:"temp_hp"`
	MaxHPBonus int  `json:"max_hp_bonus"`
}

// ListCharactersRequest represents a character list request
type ListCharactersRequest struct {
	CampaignID string         `json:"campaign_id"`
	IsNPC      *bool          `json:"is_npc"`
	PlayerID   string         `json:"player_id"`
	NPCType    models.NPCType `json:"npc_type"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
}

// CreateCharacter creates a new character
// 规则参考: PHB 第1章 Step-by-Step Characters
func (s *CharacterService) CreateCharacter(ctx context.Context, req *CreateCharacterRequest) (*models.Character, error) {
	// 1. 参数验证
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// 2. 创建角色模型
	character := models.NewCharacter(req.CampaignID, req.Name, req.IsNPC)

	// 3. 设置基础属性
	character.PlayerID = req.PlayerID
	character.NPCType = req.NPCType
	character.Race = req.Race
	character.Class = req.Class
	character.Background = req.Background
	character.Alignment = req.Alignment

	if req.Level > 0 {
		character.Level = req.Level
	}

	// 4. 设置属性值（如果提供）或使用默认值
	if req.Abilities != nil {
		if err := character.SetAbilities(req.Abilities); err != nil {
			return nil, fmt.Errorf("invalid abilities: %w", err)
		}
	}

	// 5. 计算并设置战斗属性
	s.calculateCombatStats(character, req)

	// 6. 设置技能和豁免
	if req.Skills != nil {
		character.Skills = req.Skills
	}
	if req.Saves != nil {
		character.Saves = req.Saves
	}

	// 7. 验证角色
	if err := character.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 8. 生成 UUID
	character.ID = uuid.New().String()

	// 9. 持久化
	if err := s.store.Create(ctx, character); err != nil {
		return nil, fmt.Errorf("failed to create character: %w", err)
	}

	return character, nil
}

// GetCharacter retrieves a character by ID
func (s *CharacterService) GetCharacter(ctx context.Context, id string) (*models.Character, error) {
	if id == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "character ID is required")
	}

	character, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	return character, nil
}

// GetCharacterByCampaign retrieves a character by campaign ID and character ID
func (s *CharacterService) GetCharacterByCampaign(ctx context.Context, campaignID, id string) (*models.Character, error) {
	if campaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}
	if id == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "character ID is required")
	}

	character, err := s.store.GetByCampaignAndID(ctx, campaignID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	return character, nil
}

// ListCharacters lists characters with optional filtering
func (s *CharacterService) ListCharacters(ctx context.Context, req *ListCharactersRequest) ([]*models.Character, error) {
	filter := &store.CharacterFilter{
		CampaignID: req.CampaignID,
		IsNPC:      req.IsNPC,
		PlayerID:   req.PlayerID,
		NPCType:    req.NPCType,
		Limit:      req.Limit,
		Offset:     req.Offset,
	}

	// Set default limit if not specified
	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	characters, err := s.store.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list characters: %w", err)
	}

	return characters, nil
}

// UpdateCharacter updates a character
func (s *CharacterService) UpdateCharacter(ctx context.Context, id string, req *UpdateCharacterRequest) (*models.Character, error) {
	if id == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "character ID is required")
	}

	// 获取现有角色
	character, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// 应用更新
	if req.Name != nil {
		if *req.Name == "" {
			return nil, NewServiceError(ErrCodeInvalidInput, "character name cannot be empty")
		}
		character.Name = *req.Name
	}

	if req.Race != nil {
		character.Race = *req.Race
	}

	if req.Class != nil {
		character.Class = *req.Class
	}

	if req.Level != nil {
		if err := character.SetLevel(*req.Level); err != nil {
			return nil, fmt.Errorf("invalid level: %w", err)
		}
	}

	if req.Background != nil {
		character.Background = *req.Background
	}

	if req.Alignment != nil {
		character.Alignment = *req.Alignment
	}

	if req.Abilities != nil {
		if err := character.SetAbilities(req.Abilities); err != nil {
			return nil, fmt.Errorf("invalid abilities: %w", err)
		}
	}

	if req.HP != nil {
		if err := character.SetHP(req.HP); err != nil {
			return nil, fmt.Errorf("invalid HP: %w", err)
		}
	}

	if req.AC != nil {
		if *req.AC < 0 {
			return nil, NewServiceError(ErrCodeInvalidInput, "AC cannot be negative")
		}
		character.AC = *req.AC
	}

	if req.Speed != nil {
		if *req.Speed < 0 {
			return nil, NewServiceError(ErrCodeInvalidInput, "speed cannot be negative")
		}
		character.Speed = *req.Speed
	}

	if req.Initiative != nil {
		character.Initiative = *req.Initiative
	}

	if req.Skills != nil {
		character.Skills = req.Skills
	}

	if req.Saves != nil {
		character.Saves = req.Saves
	}

	// 验证更新后的角色
	if err := character.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 保存更新
	if err := s.store.Update(ctx, character); err != nil {
		return nil, fmt.Errorf("failed to update character: %w", err)
	}

	return character, nil
}

// DeleteCharacter deletes a character
func (s *CharacterService) DeleteCharacter(ctx context.Context, id string) error {
	if id == "" {
		return NewServiceError(ErrCodeInvalidInput, "character ID is required")
	}

	// 验证角色存在
	_, err := s.store.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}

	// 删除角色
	if err := s.store.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete character: %w", err)
	}

	return nil
}

// ChangeHP changes a character's HP
// 规则参考: PHB 第9章 Damage and Healing
func (s *CharacterService) ChangeHP(ctx context.Context, id string, req *HPChangeRequest) (*models.Character, error) {
	if id == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "character ID is required")
	}

	// 获取角色
	character, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// 确保 HP 结构存在
	if character.HP == nil {
		character.HP = models.NewHP(1) // 默认最小 HP
	}

	// 应用伤害（先扣临时 HP，再扣当前 HP）
	if req.Damage > 0 {
		character.TakeDamage(req.Damage)
	}

	// 应用治疗
	if req.Healing > 0 {
		character.Heal(req.Healing)
	}

	// 添加临时 HP（不叠加，取较大值）
	if req.TempHP > 0 {
		character.HP.AddTempHP(req.TempHP)
		character.UpdatedAt = character.UpdatedAt // 触发更新时间
	}

	// 增加最大 HP
	if req.MaxHPBonus > 0 {
		character.HP.Max += req.MaxHPBonus
		character.HP.Current += req.MaxHPBonus // 同时增加当前 HP
		character.UpdatedAt = character.UpdatedAt
	}

	// 验证 HP
	if err := character.HP.Validate(); err != nil {
		return nil, fmt.Errorf("invalid HP after change: %w", err)
	}

	// 保存更新
	if err := s.store.Update(ctx, character); err != nil {
		return nil, fmt.Errorf("failed to update character: %w", err)
	}

	return character, nil
}

// AddCondition adds a condition to a character
// 规则参考: PHB 附录A Conditions
func (s *CharacterService) AddCondition(ctx context.Context, id string, conditionType string, duration int, source string) (*models.Character, error) {
	if id == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "character ID is required")
	}
	if conditionType == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "condition type is required")
	}

	// 获取角色
	character, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// 添加状态
	character.AddCondition(conditionType, duration, source)

	// 保存更新
	if err := s.store.Update(ctx, character); err != nil {
		return nil, fmt.Errorf("failed to update character: %w", err)
	}

	return character, nil
}

// RemoveCondition removes a condition from a character
func (s *CharacterService) RemoveCondition(ctx context.Context, id string, conditionType string) (*models.Character, error) {
	if id == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "character ID is required")
	}
	if conditionType == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "condition type is required")
	}

	// 获取角色
	character, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// 移除状态
	if !character.RemoveCondition(conditionType) {
		return nil, NewServiceError(ErrCodeNotFound, fmt.Sprintf("condition %s not found", conditionType))
	}

	// 保存更新
	if err := s.store.Update(ctx, character); err != nil {
		return nil, fmt.Errorf("failed to update character: %w", err)
	}

	return character, nil
}

// CountCharacters counts characters with optional filtering
func (s *CharacterService) CountCharacters(ctx context.Context, req *ListCharactersRequest) (int64, error) {
	filter := &store.CharacterFilter{
		CampaignID: req.CampaignID,
		IsNPC:      req.IsNPC,
		PlayerID:   req.PlayerID,
		NPCType:    req.NPCType,
	}

	count, err := s.store.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count characters: %w", err)
	}

	return count, nil
}

// GetAbilityModifier calculates the ability modifier for a given ability score
// 规则参考: PHB 第7章 Ability Scores and Modifiers
// 公式: modifier = floor((score - 10) / 2)
func GetAbilityModifier(score int) int {
	return int(math.Floor(float64(score-10) / 2))
}

// CalculateUnarmoredAC calculates AC without armor
// 规则参考: PHB 第5章 Armor and Shields / 第9章 Combat
// 公式: AC = 10 + Dexterity modifier
func CalculateUnarmoredAC(dexterity int) int {
	return 10 + GetAbilityModifier(dexterity)
}

// validateCreateRequest validates the create character request
func (s *CharacterService) validateCreateRequest(req *CreateCharacterRequest) error {
	if req.CampaignID == "" {
		return NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}
	if req.Name == "" {
		return NewServiceError(ErrCodeInvalidInput, "character name is required")
	}

	// 玩家角色必须有 player_id
	if !req.IsNPC && req.PlayerID == "" {
		return NewServiceError(ErrCodeInvalidInput, "player ID is required for player characters")
	}

	// 玩家角色不能设置 npc_type
	// 规则参考: D&D 5e 玩家角色与NPC是不同的概念
	if !req.IsNPC && req.NPCType != "" {
		return NewServiceError(ErrCodeInvalidInput, "player characters cannot have NPC type")
	}

	// NPC 类型验证
	if req.IsNPC && req.NPCType != "" {
		if req.NPCType != models.NPCTypeScripted && req.NPCType != models.NPCTypeGenerated {
			return NewServiceError(ErrCodeInvalidInput, fmt.Sprintf("invalid NPC type: %s", req.NPCType))
		}
	}

	return nil
}

// calculateCombatStats calculates and sets combat-related stats
// 规则参考: PHB 第5章 Armor, 第9章 Combat
func (s *CharacterService) calculateCombatStats(character *models.Character, req *CreateCharacterRequest) {
	// 设置 HP
	if req.HP != nil {
		character.HP = req.HP
	} else {
		// 如果没有提供 HP，基于体质计算默认 HP
		if character.Abilities != nil {
			conMod := GetAbilityModifier(character.Abilities.Constitution)
			// 1级角色基础 HP = 类基础 HP (假设为 8) + 体质修正
			baseHP := 8 + conMod
			if baseHP < 1 {
				baseHP = 1 // HP 至少为 1
			}
			character.HP = models.NewHP(baseHP)
		}
	}

	// 设置 AC
	if req.AC > 0 {
		character.AC = req.AC
	} else {
		// 如果没有提供 AC，计算无甲 AC
		if character.Abilities != nil {
			character.AC = CalculateUnarmoredAC(character.Abilities.Dexterity)
		}
	}

	// 设置速度
	if req.Speed > 0 {
		character.Speed = req.Speed
	}
	// 否则使用默认速度 30 英尺（在 NewCharacter 中已设置）

	// 设置先攻
	if req.Initiative != 0 {
		character.Initiative = req.Initiative
	} else {
		// 默认先攻 = 敏捷修正值
		if character.Abilities != nil {
			character.Initiative = GetAbilityModifier(character.Abilities.Dexterity)
		}
	}
}
