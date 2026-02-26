// Package service provides business logic layer implementations
package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/rules"
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

	// ============ 扩展字段 ============

	// 图片
	Image string `json:"image"`

	// 经验值
	Experience int `json:"experience"`

	// 熟练加值
	Proficiency int `json:"proficiency"`

	// 详细移动速度
	SpeedDetail *models.Speed `json:"speed_detail"`

	// 死亡豁免
	DeathSaves *models.DeathSaves `json:"death_saves"`

	// 详细技能
	SkillsDetail map[string]*models.Skill `json:"skills_detail"`

	// 详细豁免
	SavesDetail map[string]*models.Save `json:"saves_detail"`

	// 货币
	Currency *models.Currency `json:"currency"`

	// 装备槽位
	EquipmentSlots *models.EquipmentSlots `json:"equipment_slots"`

	// 详细背包物品
	InventoryItems []*models.InventoryItem `json:"inventory_items"`

	// 法术书
	Spellbook *models.Spellbook `json:"spellbook"`

	// 专长/特性
	Features []*models.Feature `json:"features"`

	// 传记
	Biography *models.Biography `json:"biography"`

	// 特性/抗性/语言
	Traits *models.Traits `json:"traits"`

	// 导入元数据
	ImportMeta *models.ImportMeta `json:"import_meta"`
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

	// ============ 扩展字段 ============

	Image         *string               `json:"image"`
	Experience    *int                  `json:"experience"`
	Proficiency   *int                  `json:"proficiency"`
	SpeedDetail   *models.Speed         `json:"speed_detail"`
	DeathSaves    *models.DeathSaves    `json:"death_saves"`
	SkillsDetail  map[string]*models.Skill `json:"skills_detail"`
	SavesDetail   map[string]*models.Save `json:"saves_detail"`
	Currency      *models.Currency      `json:"currency"`
	EquipmentSlots *models.EquipmentSlots `json:"equipment_slots"`
	InventoryItems []*models.InventoryItem `json:"inventory_items"`
	Spellbook     *models.Spellbook     `json:"spellbook"`
	Features      []*models.Feature     `json:"features"`
	Biography     *models.Biography     `json:"biography"`
	Traits        *models.Traits        `json:"traits"`
	ImportMeta    *models.ImportMeta    `json:"import_meta"`
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

	// 7. 设置扩展字段
	if req.Image != "" {
		character.Image = req.Image
	}
	if req.Experience > 0 {
		character.Experience = req.Experience
	}
	if req.Proficiency > 0 {
		character.Proficiency = req.Proficiency
	}
	if req.SpeedDetail != nil {
		character.SpeedDetail = req.SpeedDetail
	}
	if req.DeathSaves != nil {
		character.DeathSaves = req.DeathSaves
	}
	if req.SkillsDetail != nil {
		character.SkillsDetail = req.SkillsDetail
	}
	if req.SavesDetail != nil {
		character.SavesDetail = req.SavesDetail
	}
	if req.Currency != nil {
		character.Currency = req.Currency
	}
	if req.EquipmentSlots != nil {
		character.EquipmentSlots = req.EquipmentSlots
	}
	if req.InventoryItems != nil {
		character.InventoryItems = req.InventoryItems
	}
	if req.Spellbook != nil {
		character.Spellbook = req.Spellbook
	}
	if req.Features != nil {
		character.Features = req.Features
	}
	if req.Biography != nil {
		character.Biography = req.Biography
	}
	if req.Traits != nil {
		character.Traits = req.Traits
	}
	if req.ImportMeta != nil {
		character.ImportMeta = req.ImportMeta
	}

	// 8. 验证角色
	if err := character.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 9. 生成 UUID
	character.ID = uuid.New().String()

	// 10. 持久化
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

	// 更新扩展字段
	if req.Image != nil {
		character.Image = *req.Image
	}
	if req.Experience != nil {
		character.Experience = *req.Experience
	}
	if req.Proficiency != nil {
		character.Proficiency = *req.Proficiency
	}
	if req.SpeedDetail != nil {
		character.SpeedDetail = req.SpeedDetail
	}
	if req.DeathSaves != nil {
		character.DeathSaves = req.DeathSaves
	}
	if req.SkillsDetail != nil {
		character.SkillsDetail = req.SkillsDetail
	}
	if req.SavesDetail != nil {
		character.SavesDetail = req.SavesDetail
	}
	if req.Currency != nil {
		character.Currency = req.Currency
	}
	if req.EquipmentSlots != nil {
		character.EquipmentSlots = req.EquipmentSlots
	}
	if req.InventoryItems != nil {
		character.InventoryItems = req.InventoryItems
	}
	if req.Spellbook != nil {
		character.Spellbook = req.Spellbook
	}
	if req.Features != nil {
		character.Features = req.Features
	}
	if req.Biography != nil {
		character.Biography = req.Biography
	}
	if req.Traits != nil {
		character.Traits = req.Traits
	}
	if req.ImportMeta != nil {
		character.ImportMeta = req.ImportMeta
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
		character.UpdatedAt = time.Now()
	}

	// 增加最大 HP
	if req.MaxHPBonus > 0 {
		character.HP.Max += req.MaxHPBonus
		character.HP.Current += req.MaxHPBonus // 同时增加当前 HP
		character.UpdatedAt = time.Now()
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

// DeathSaveRequest 死亡豁免请求
type DeathSaveRequest struct {
	Roll int `json:"roll"` // 豁免骰结果 (1-20)
}

// DeathSaveResponse 死亡豁免响应
type DeathSaveResponse struct {
	Result      string `json:"result"`       // 结果: success, failure, critical_success, critical_failure
	IsStable    bool   `json:"is_stable"`    // 是否稳定
	IsDead      bool   `json:"is_dead"`      // 是否死亡
	Successes   int    `json:"successes"`    // 成功次数
	Failures    int    `json:"failures"`     // 失败次数
	HealedHP    int    `json:"healed_hp"`    // 恢复的 HP（大成功时）
}

// MakeDeathSave 执行死亡豁免
// 规则参考: PHB 第9章 - Dropping to 0 Hit Points / Death Saving Throws
func (s *CharacterService) MakeDeathSave(ctx context.Context, id string, req *DeathSaveRequest) (*DeathSaveResponse, error) {
	if id == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "character ID is required")
	}
	if req.Roll < 1 || req.Roll > 20 {
		return nil, NewServiceError(ErrCodeInvalidInput, "roll must be between 1 and 20")
	}

	// 获取角色
	character, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// 验证角色处于昏迷状态
	if !character.IsUnconscious() {
		return nil, NewServiceError(ErrCodeInvalidInput, "character is not unconscious")
	}

	// 确保死亡豁免结构存在
	deathSaves := character.GetDeathSaves()

	response := &DeathSaveResponse{
		Successes: deathSaves.Successes,
		Failures:  deathSaves.Failures,
	}

	// 执行豁免
	// 自然 20: 大成功，恢复 1 HP
	if req.Roll == 20 {
		deathSaves.Reset()
		if character.HP != nil {
			response.HealedHP = character.HP.Heal(1)
		}
		response.Result = "critical_success"
		response.Successes = 0
		response.Failures = 0
	} else if req.Roll == 1 {
		// 自然 1: 大失败，失败 +2
		deathSaves.AddFailure()
		deathSaves.AddFailure()
		response.Result = "critical_failure"
		response.Failures = deathSaves.Failures
	} else if req.Roll >= 10 {
		// >= 10: 成功
		deathSaves.AddSuccess()
		response.Result = "success"
		response.Successes = deathSaves.Successes
		response.IsStable = deathSaves.IsStable()
	} else {
		// < 10: 失败
		deathSaves.AddFailure()
		response.Result = "failure"
		response.Failures = deathSaves.Failures
	}

	// 检查死亡
	response.IsDead = deathSaves.IsDead()

	// 保存更新
	if err := s.store.Update(ctx, character); err != nil {
		return nil, fmt.Errorf("failed to update character: %w", err)
	}

	return response, nil
}

// DamageWhileUnconsciousRequest 昏迷时受伤请求
type DamageWhileUnconsciousRequest struct {
	Damage int  `json:"damage"` // 伤害值
	IsCrit bool `json:"is_crit"` // 是否暴击
}

// DamageWhileUnconsciousResponse 昏迷时受伤响应
type DamageWhileUnconsciousResponse struct {
	IsDead    bool `json:"is_dead"`     // 是否死亡
	Failures  int  `json:"failures"`    // 当前失败次数
	InstaKill bool `json:"insta_kill"`  // 是否被立即击杀（伤害 >= MaxHP）
}

// TakeDamageWhileUnconscious 昏迷状态下受到伤害
// 规则参考: PHB 第9章 - Dropping to 0 Hit Points / Damage at 0 Hit Points
func (s *CharacterService) TakeDamageWhileUnconscious(ctx context.Context, id string, req *DamageWhileUnconsciousRequest) (*DamageWhileUnconsciousResponse, error) {
	if id == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "character ID is required")
	}
	if req.Damage <= 0 {
		return nil, NewServiceError(ErrCodeInvalidInput, "damage must be positive")
	}

	// 获取角色
	character, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// 验证角色处于昏迷状态
	if !character.IsUnconscious() {
		return nil, NewServiceError(ErrCodeInvalidInput, "character is not unconscious")
	}

	// 确保死亡豁免结构存在
	deathSaves := character.GetDeathSaves()

	response := &DamageWhileUnconsciousResponse{}

	// 检查是否造成立即死亡（伤害 >= MaxHP）
	if character.HP != nil && req.Damage >= character.HP.Max {
		deathSaves.Failures = 3
		response.IsDead = true
		response.InstaKill = true
		response.Failures = 3
	} else {
		// 暴击: 失败 +2
		if req.IsCrit {
			deathSaves.AddFailure()
			deathSaves.AddFailure()
		} else {
			// 普通伤害: 失败 +1
			deathSaves.AddFailure()
		}
		response.Failures = deathSaves.Failures
		response.IsDead = deathSaves.IsDead()
	}

	// 保存更新
	if err := s.store.Update(ctx, character); err != nil {
		return nil, fmt.Errorf("failed to update character: %w", err)
	}

	return response, nil
}

// StabilizeCharacter 稳定角色（通过 Medicine 检定或其他效果）
// 规则参考: PHB 第9章 - Dropping to 0 Hit Points / Stabilizing
func (s *CharacterService) StabilizeCharacter(ctx context.Context, id string) (*models.Character, error) {
	if id == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "character ID is required")
	}

	// 获取角色
	character, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// 验证角色处于昏迷状态
	if !character.IsUnconscious() {
		return nil, NewServiceError(ErrCodeInvalidInput, "character is not unconscious")
	}

	// 稳定角色
	deathSaves := character.GetDeathSaves()
	deathSaves.Successes = 3
	deathSaves.Failures = 0

	// 保存更新
	if err := s.store.Update(ctx, character); err != nil {
		return nil, fmt.Errorf("failed to update character: %w", err)
	}

	return character, nil
}

// EncumbranceResponse 负重响应
type EncumbranceResponse struct {
	Carried      float64 `json:"carried"`        // 当前负重（磅）
	Capacity     int     `json:"capacity"`       // 负重能力（磅）
	PushDragLift int     `json:"push_drag_lift"` // 推/拖/举起重量（磅）
	IsEncumbered bool    `json:"is_encumbered"`  // 是否超重
	SpeedPenalty int     `json:"speed_penalty"`  // 速度惩罚（变体规则）
	Level        string  `json:"level"`          // 超重等级
}

// GetEncumbrance 获取角色的负重状态
// 规则参考: PHB 第7章 - Lifting and Carrying
func (s *CharacterService) GetEncumbrance(ctx context.Context, id string, useVariantRules bool) (*EncumbranceResponse, error) {
	if id == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "character ID is required")
	}

	// 获取角色
	character, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	return s.calculateEncumbranceForCharacter(character, useVariantRules), nil
}

// calculateEncumbranceForCharacter 计算角色的负重状态
func (s *CharacterService) calculateEncumbranceForCharacter(character *models.Character, useVariantRules bool) *EncumbranceResponse {
	// 获取力量值
	strength := 10 // 默认力量值
	if character.Abilities != nil {
		strength = character.Abilities.Strength
	}

	// 确定体型（默认中型）
	size := models.SizeMedium
	// TODO: 根据种族确定体型，这里暂时使用默认值

	// 计算负重
	enc := rules.CalculateEncumbrance(
		strength,
		size,
		character.EquipmentSlots,
		character.InventoryItems,
		character.Equipment,
		character.Inventory,
		character.Currency,
	)

	// 计算超重等级
	level := rules.GetEncumbranceLevel(enc.Carried, strength, useVariantRules)

	// 计算速度惩罚
	baseSpeed := character.Speed
	if character.SpeedDetail != nil {
		baseSpeed = character.SpeedDetail.Walk
	}
	encumberedSpeed := rules.GetEncumberedSpeed(baseSpeed, level)
	speedPenalty := baseSpeed - encumberedSpeed

	// 转换超重等级为字符串
	levelStr := "none"
	switch level {
	case rules.EncumbranceLight:
		levelStr = "light"
	case rules.EncumbranceHeavy:
		levelStr = "heavy"
	case rules.EncumbranceOverCapacity:
		levelStr = "over_capacity"
	}

	return &EncumbranceResponse{
		Carried:      enc.Carried,
		Capacity:     enc.Capacity,
		PushDragLift: enc.PushDragLift,
		IsEncumbered: enc.IsEncumbered,
		SpeedPenalty: speedPenalty,
		Level:        levelStr,
	}
}

// GetCharacterWithEncumbrance 获取角色及其负重状态
// 规则参考: PHB 第7章 - Lifting and Carrying
func (s *CharacterService) GetCharacterWithEncumbrance(ctx context.Context, id string, useVariantRules bool) (*models.Character, *EncumbranceResponse, error) {
	if id == "" {
		return nil, nil, NewServiceError(ErrCodeInvalidInput, "character ID is required")
	}

	// 获取角色
	character, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get character: %w", err)
	}

	// 计算负重
	encResponse := s.calculateEncumbranceForCharacter(character, useVariantRules)

	return character, encResponse, nil
}
