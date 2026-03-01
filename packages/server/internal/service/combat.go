// Package service provides business logic layer implementations
package service

import (
	"context"
	"fmt"

	"github.com/dnd-mcp/server/internal/models"
	rulescombat "github.com/dnd-mcp/server/internal/rules/combat"
	"github.com/dnd-mcp/server/internal/rules/dice"
)

// CombatStore defines the interface for combat data operations
type CombatStore interface {
	Create(ctx context.Context, combat *models.Combat) error
	Get(ctx context.Context, id string) (*models.Combat, error)
	GetByCampaign(ctx context.Context, campaignID string) ([]*models.Combat, error)
	GetActive(ctx context.Context, campaignID string) (*models.Combat, error)
	Update(ctx context.Context, combat *models.Combat) error
	Delete(ctx context.Context, id string) error
}

// CampaignStoreForCombat defines the campaign store interface needed by combat service
type CampaignStoreForCombat interface {
	Get(ctx context.Context, id string) (*models.Campaign, error)
}

// CombatService provides combat business logic
// 规则参考: PHB 第9章 Combat
type CombatService struct {
	combatStore    CombatStore
	characterStore CharacterStore
	campaignStore  CampaignStoreForCombat
	diceService    *DiceService
	roller         *dice.Roller
}

// NewCombatService creates a new combat service
func NewCombatService(
	combatStore CombatStore,
	characterStore CharacterStore,
	campaignStore CampaignStoreForCombat,
	diceService *DiceService,
) *CombatService {
	return &CombatService{
		combatStore:    combatStore,
		characterStore: characterStore,
		campaignStore:  campaignStore,
		diceService:    diceService,
		roller:         dice.NewRoller(),
	}
}

// NewCombatServiceWithRoller creates a new combat service with a custom roller (for testing)
func NewCombatServiceWithRoller(
	combatStore CombatStore,
	characterStore CharacterStore,
	campaignStore CampaignStoreForCombat,
	diceService *DiceService,
	roller *dice.Roller,
) *CombatService {
	return &CombatService{
		combatStore:    combatStore,
		characterStore: characterStore,
		campaignStore:  campaignStore,
		diceService:    diceService,
		roller:         roller,
	}
}

// StartCombatRequest 开始战斗请求
type StartCombatRequest struct {
	CampaignID     string   `json:"campaign_id"`
	ParticipantIDs []string `json:"participant_ids"`
	MapID          string   `json:"map_id"`
}

// StartCombat 开始战斗
// 规则参考: PHB 第9章 - Combat Step-by-Step / Initiative
func (s *CombatService) StartCombat(ctx context.Context, req *StartCombatRequest) (*models.Combat, error) {
	// 1. 参数验证
	if req.CampaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}
	if len(req.ParticipantIDs) == 0 {
		return nil, NewServiceError(ErrCodeInvalidInput, "at least one participant is required")
	}

	// 2. 验证战役存在
	_, err := s.campaignStore.Get(ctx, req.CampaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	// 3. 检查是否已有活动战斗
	activeCombat, err := s.combatStore.GetActive(ctx, req.CampaignID)
	if err == nil && activeCombat != nil {
		return nil, NewServiceError(ErrCodeInvalidState, "campaign already has an active combat")
	}

	// 4. 验证参战角色存在并投先攻
	participants := make([]models.Participant, 0, len(req.ParticipantIDs))
	for _, id := range req.ParticipantIDs {
		character, err := s.characterStore.Get(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get character %s: %w", id, err)
		}

		// 投先攻：1d20 + 敏捷调整值
		// 规则参考: PHB 第9章 - Initiative
		initiative := rulescombat.GetInitiative(character, s.roller)

		participants = append(participants, models.Participant{
			CharacterID: id,
			Initiative:  initiative,
			HasActed:    false,
			Conditions:  make([]models.Condition, 0),
		})
	}

	// 5. 创建战斗记录
	newCombat := models.NewCombat(req.CampaignID, req.ParticipantIDs)
	newCombat.Participants = participants
	newCombat.MapID = req.MapID

	// 6. 按先攻值排序（高到低）
	// 规则参考: PHB 第9章 - Initiative
	newCombat.SortParticipantsByInitiative()

	// 7. 保存战斗
	if err := s.combatStore.Create(ctx, newCombat); err != nil {
		return nil, fmt.Errorf("failed to create combat: %w", err)
	}

	// 8. 记录战斗日志
	newCombat.AddLogEntry("", "combat_start", "", fmt.Sprintf("Combat started with %d participants", len(participants)))

	return newCombat, nil
}

// GetCombatState 获取战斗状态
func (s *CombatService) GetCombatState(ctx context.Context, combatID string) (*models.Combat, error) {
	if combatID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "combat ID is required")
	}

	combat, err := s.combatStore.Get(ctx, combatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get combat: %w", err)
	}

	return combat, nil
}

// GetActiveCombat 获取战役的活动战斗
func (s *CombatService) GetActiveCombat(ctx context.Context, campaignID string) (*models.Combat, error) {
	if campaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}

	combat, err := s.combatStore.GetActive(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active combat: %w", err)
	}

	return combat, nil
}

// AttackRequest 攻击请求
type AttackRequest struct {
	CombatID     string `json:"combat_id"`
	AttackerID   string `json:"attacker_id"`
	TargetID     string `json:"target_id"`
	Advantage    bool   `json:"advantage"`
	Disadvantage bool   `json:"disadvantage"`
}

// AttackResponse 攻击响应
type AttackResponse struct {
	Result     *rulescombat.AttackResult `json:"result"`
	Combat     *models.Combat            `json:"combat"`
	TargetDead bool                      `json:"target_dead"`
}

// Attack 执行攻击
// 规则参考: PHB 第9章 - Making an Attack
func (s *CombatService) Attack(ctx context.Context, req *AttackRequest) (*AttackResponse, error) {
	// 1. 参数验证
	if req.CombatID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "combat ID is required")
	}
	if req.AttackerID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "attacker ID is required")
	}
	if req.TargetID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "target ID is required")
	}

	// 2. 获取战斗
	combat, err := s.combatStore.Get(ctx, req.CombatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get combat: %w", err)
	}

	// 3. 验证战斗状态
	if !combat.IsActive() {
		return nil, NewServiceError(ErrCodeInvalidState, "combat is not active")
	}

	// 4. 验证当前是攻击者的回合
	currentParticipant := combat.GetCurrentParticipant()
	if currentParticipant == nil || currentParticipant.CharacterID != req.AttackerID {
		return nil, NewServiceError(ErrCodeInvalidState, "not attacker's turn")
	}

	// 5. 获取攻击者和目标角色
	attacker, err := s.characterStore.Get(ctx, req.AttackerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attacker: %w", err)
	}

	target, err := s.characterStore.Get(ctx, req.TargetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target: %w", err)
	}

	// 6. 检查目标是否已在战斗中
	targetParticipant := combat.GetParticipantByCharacterID(req.TargetID)
	if targetParticipant == nil {
		return nil, NewServiceError(ErrCodeInvalidInput, "target is not in combat")
	}

	// 7. 获取攻击者装备的武器
	weapon := s.getEquippedWeapon(attacker)

	// 8. 执行攻击检定
	// 规则参考: PHB 第9章 - Attack Rolls
	result := rulescombat.ResolveAttack(attacker, target, weapon, req.Advantage, req.Disadvantage, s.roller)

	// 9. 如果命中并造成伤害，更新目标 HP
	if result.Hit && result.Damage > 0 {
		// 更新角色 HP
		target.TakeDamage(result.Damage)
		if err := s.characterStore.Update(ctx, target); err != nil {
			return nil, fmt.Errorf("failed to update target: %w", err)
		}

		// 更新参战者临时 HP（用于战斗追踪）
		if target.HP != nil {
			result.TargetHP = &models.HP{
				Current: target.HP.Current,
				Max:     target.HP.Max,
				Temp:    target.HP.Temp,
			}
			result.TargetDown = target.HP.IsAtZero()
		}
	}

	// 10. 记录战斗日志
	action := "attack"
	if result.Crit {
		action = "critical_hit"
	}
	resultDesc := fmt.Sprintf("roll %d vs AC %d", result.AttackRoll.Total, result.TargetAC)
	if result.Hit {
		resultDesc += fmt.Sprintf(", hit for %d damage", result.Damage)
	} else {
		resultDesc += ", miss"
	}
	combat.AddLogEntry(req.AttackerID, action, req.TargetID, resultDesc)

	// 11. 保存战斗状态
	if err := s.combatStore.Update(ctx, combat); err != nil {
		return nil, fmt.Errorf("failed to update combat: %w", err)
	}

	return &AttackResponse{
		Result:     result,
		Combat:     combat,
		TargetDead: target.IsDead(),
	}, nil
}

// CastSpellRequest 施法请求
type CastSpellRequest struct {
	CombatID  string   `json:"combat_id"`
	CasterID  string   `json:"caster_id"`
	SpellID   string   `json:"spell_id"`
	SpellName string   `json:"spell_name"`
	TargetIDs []string `json:"target_ids"`
	Level     int      `json:"level"`
	Damage    string   `json:"damage"`     // 伤害公式（如 "2d6"）
	DamageType string  `json:"damage_type"` // 伤害类型
	IsHealing  bool    `json:"is_healing"`  // 是否为治疗法术
}

// CastSpellResponse 施法响应
type CastSpellResponse struct {
	Result *rulescombat.SpellResult `json:"result"`
	Combat *models.Combat           `json:"combat"`
}

// CastSpell 施放法术
// 规则参考: PHB 第10-11章 - Spellcasting
func (s *CombatService) CastSpell(ctx context.Context, req *CastSpellRequest) (*CastSpellResponse, error) {
	// 1. 参数验证
	if req.CombatID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "combat ID is required")
	}
	if req.CasterID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "caster ID is required")
	}
	if req.SpellID == "" && req.SpellName == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "spell ID or name is required")
	}
	if len(req.TargetIDs) == 0 {
		return nil, NewServiceError(ErrCodeInvalidInput, "at least one target is required")
	}

	// 2. 获取战斗
	combat, err := s.combatStore.Get(ctx, req.CombatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get combat: %w", err)
	}

	// 3. 验证战斗状态
	if !combat.IsActive() {
		return nil, NewServiceError(ErrCodeInvalidState, "combat is not active")
	}

	// 4. 验证当前是施法者的回合
	currentParticipant := combat.GetCurrentParticipant()
	if currentParticipant == nil || currentParticipant.CharacterID != req.CasterID {
		return nil, NewServiceError(ErrCodeInvalidState, "not caster's turn")
	}

	// 5. 获取施法者
	caster, err := s.characterStore.Get(ctx, req.CasterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get caster: %w", err)
	}

	// 6. 创建法术结果
	result := &rulescombat.SpellResult{
		SpellID:    req.SpellID,
		SpellName:  req.SpellName,
		Level:      req.Level,
		CasterID:   req.CasterID,
		TargetIDs:  req.TargetIDs,
		DamageType: req.DamageType,
		IsHealing:  req.IsHealing,
		Results:    make([]rulescombat.SpellTargetResult, 0),
	}

	// 7. 对每个目标应用法术效果
	for _, targetID := range req.TargetIDs {
		target, err := s.characterStore.Get(ctx, targetID)
		if err != nil {
			continue // 跳过无效目标
		}

		targetResult := rulescombat.SpellTargetResult{
			TargetID: targetID,
			Hit:      true, // 简化版：默认命中
		}

		// 投伤害/治疗骰
		if req.Damage != "" {
			formula, err := dice.ParseFormula(req.Damage)
			if err == nil {
				diceResult := s.roller.RollFormula(formula)
				targetResult.Damage = diceResult.Total

				// 添加施法属性调整值
				// 规则参考: PHB 第10章 - Spellcasting Ability
				abilityMod := s.getSpellcastingModifier(caster)
				targetResult.Damage += abilityMod

				// 应用伤害或治疗
				if req.IsHealing {
					target.Heal(targetResult.Damage)
				} else {
					target.TakeDamage(targetResult.Damage)
				}

				// 更新目标
				if err := s.characterStore.Update(ctx, target); err != nil {
					continue
				}

				// 记录 HP
				if target.HP != nil {
					hpCopy := *target.HP
					targetResult.TargetHP = &hpCopy
				}
			}
		}

		result.Results = append(result.Results, targetResult)
		if !req.IsHealing {
			result.Damage += targetResult.Damage
		} else {
			result.Damage += targetResult.Damage // 治疗量也记录在 Damage 字段
		}
	}

	// 8. 记录战斗日志
	action := "spell_cast"
	if req.IsHealing {
		action = "healing_spell"
	}
	resultDesc := fmt.Sprintf("cast %s", req.SpellName)
	if result.Damage > 0 {
		if req.IsHealing {
			resultDesc += fmt.Sprintf(", healed %d", result.Damage)
		} else {
			resultDesc += fmt.Sprintf(", dealt %d %s damage", result.Damage, req.DamageType)
		}
	}
	combat.AddLogEntry(req.CasterID, action, "", resultDesc)

	// 9. 保存战斗状态
	if err := s.combatStore.Update(ctx, combat); err != nil {
		return nil, fmt.Errorf("failed to update combat: %w", err)
	}

	return &CastSpellResponse{
		Result: result,
		Combat: combat,
	}, nil
}

// AdvanceTurnRequest 推进回合请求
type AdvanceTurnRequest struct {
	CombatID string `json:"combat_id"`
}

// AdvanceTurnResponse 推进回合响应
type AdvanceTurnResponse struct {
	Combat          *models.Combat `json:"combat"`
	NewRound        bool           `json:"new_round"`
	CurrentTurnName string         `json:"current_turn_name"`
}

// AdvanceTurn 推进回合
func (s *CombatService) AdvanceTurn(ctx context.Context, combatID string) (*AdvanceTurnResponse, error) {
	if combatID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "combat ID is required")
	}

	// 获取战斗
	combat, err := s.combatStore.Get(ctx, combatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get combat: %w", err)
	}

	// 验证战斗状态
	if !combat.IsActive() {
		return nil, NewServiceError(ErrCodeInvalidState, "combat is not active")
	}

	// 推进回合
	newRound := combat.AdvanceTurn()

	// 推进所有参战者状态持续时间
	for i := range combat.Participants {
		expired := combat.Participants[i].TickConditions()
		for _, cond := range expired {
			combat.AddLogEntry(combat.Participants[i].CharacterID, "condition_expired", "",
				fmt.Sprintf("condition %s expired", cond))
		}
	}

	// 记录日志
	if newRound {
		combat.AddLogEntry("", "new_round", "", fmt.Sprintf("Round %d begins", combat.Round))
	}

	// 获取当前行动者名称
	currentParticipant := combat.GetCurrentParticipant()
	currentTurnName := ""
	if currentParticipant != nil {
		char, err := s.characterStore.Get(ctx, currentParticipant.CharacterID)
		if err == nil {
			currentTurnName = char.Name
		}
	}

	// 保存战斗
	if err := s.combatStore.Update(ctx, combat); err != nil {
		return nil, fmt.Errorf("failed to update combat: %w", err)
	}

	return &AdvanceTurnResponse{
		Combat:          combat,
		NewRound:        newRound,
		CurrentTurnName: currentTurnName,
	}, nil
}

// EndCombatRequest 结束战斗请求
type EndCombatRequest struct {
	CombatID string `json:"combat_id"`
}

// EndCombat 结束战斗
func (s *CombatService) EndCombat(ctx context.Context, combatID string) (*models.Combat, error) {
	if combatID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "combat ID is required")
	}

	// 获取战斗
	combat, err := s.combatStore.Get(ctx, combatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get combat: %w", err)
	}

	// 验证战斗状态
	if !combat.IsActive() {
		return nil, NewServiceError(ErrCodeInvalidState, "combat is not active")
	}

	// 结束战斗
	combat.End()
	combat.AddLogEntry("", "combat_end", "", "Combat ended")

	// 保存战斗
	if err := s.combatStore.Update(ctx, combat); err != nil {
		return nil, fmt.Errorf("failed to update combat: %w", err)
	}

	return combat, nil
}

// getEquippedWeapon 获取角色装备的武器
func (s *CombatService) getEquippedWeapon(character *models.Character) *models.EquipmentItem {
	if character.EquipmentSlots == nil {
		return nil
	}

	// 优先使用主手武器
	if character.EquipmentSlots.MainHand != nil {
		return character.EquipmentSlots.MainHand
	}

	// 检查副手是否为武器（不是盾牌）
	if character.EquipmentSlots.OffHand != nil &&
		character.EquipmentSlots.OffHand.Type == models.EquipmentTypeWeapon {
		return character.EquipmentSlots.OffHand
	}

	// 向后兼容：检查旧版 Equipment 列表
	for _, eq := range character.Equipment {
		if eq.Slot == "main_hand" || eq.Slot == "weapon" {
			return &models.EquipmentItem{
				ID:         eq.ID,
				Name:       eq.Name,
				Type:       models.EquipmentTypeWeapon,
				Damage:     eq.Damage,
				DamageType: eq.DamageType,
				MagicBonus: eq.Bonus, // Equipment.Bonus 对应 EquipmentItem.MagicBonus
			}
		}
	}

	return nil
}

// getSpellcastingModifier 获取施法属性调整值
// 规则参考: PHB 第10章 - Spellcasting Ability
func (s *CombatService) getSpellcastingModifier(character *models.Character) int {
	if character == nil || character.Abilities == nil {
		return 0
	}

	// 根据职业决定施法属性
	// 规则参考: PHB 第10章 - Spellcasting Ability
	class := character.Class
	switch class {
	case "Bard", "Sorcerer", "Warlock":
		return GetAbilityModifier(character.Abilities.Charisma)
	case "Cleric", "Druid", "Ranger":
		return GetAbilityModifier(character.Abilities.Wisdom)
	case "Wizard", "Artificer":
		return GetAbilityModifier(character.Abilities.Intelligence)
	case "Paladin":
		return GetAbilityModifier(character.Abilities.Charisma)
	default:
		// 默认使用魅力（适用于邪术师、吟游诗人等）
		return GetAbilityModifier(character.Abilities.Charisma)
	}
}

// CombatSummary 战斗统计
type CombatSummary struct {
	TotalRounds  int                  `json:"total_rounds"`
	TotalTurns   int                  `json:"total_turns"`
	Duration     string               `json:"duration"` // 使用字符串以便 JSON 序列化
	Participants []ParticipantSummary `json:"participants"`
}

// ParticipantSummary 参战者统计
type ParticipantSummary struct {
	CharacterID     string `json:"character_id"`
	CharacterName   string `json:"character_name"`
	DamageDealt     int    `json:"damage_dealt"`
	DamageTaken     int    `json:"damage_taken"`
	HealingReceived int    `json:"healing_received"`
	Kills           int    `json:"kills"`
	FinalHP         int    `json:"final_hp"`
	MaxHP           int    `json:"max_hp"`
	Survived        bool   `json:"survived"`
}

// EndCombatWithSummaryResponse 结束战斗响应（带统计）
type EndCombatWithSummaryResponse struct {
	Combat *models.Combat `json:"combat"`
	Summary *CombatSummary `json:"summary"`
}

// EndCombatWithSummary 结束战斗并生成战斗统计报告
func (s *CombatService) EndCombatWithSummary(ctx context.Context, combatID string) (*EndCombatWithSummaryResponse, error) {
	if combatID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "combat ID is required")
	}

	// 获取战斗
	combat, err := s.combatStore.Get(ctx, combatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get combat: %w", err)
	}

	// 验证战斗状态
	if !combat.IsActive() {
		return nil, NewServiceError(ErrCodeInvalidState, "combat is not active")
	}

	// 结束战斗
	combat.End()
	combat.AddLogEntry("", "combat_end", "", "Combat ended")

	// 生成战斗统计
	summary := s.generateCombatSummary(ctx, combat)

	// 保存战斗
	if err := s.combatStore.Update(ctx, combat); err != nil {
		return nil, fmt.Errorf("failed to update combat: %w", err)
	}

	return &EndCombatWithSummaryResponse{
		Combat:  combat,
		Summary: summary,
	}, nil
}

// generateCombatSummary 从战斗日志中生成战斗统计
func (s *CombatService) generateCombatSummary(ctx context.Context, combat *models.Combat) *CombatSummary {
	summary := &CombatSummary{
		TotalRounds:  combat.Round,
		Participants: make([]ParticipantSummary, 0),
	}

	// 计算战斗时长
	if combat.EndedAt != nil {
		duration := combat.EndedAt.Sub(combat.StartedAt)
		summary.Duration = duration.String()
	}

	// 计算总回合数（估算：每个回合 * 参战者数量）
	summary.TotalTurns = len(combat.Log)

	// 初始化参战者统计
	participantStats := make(map[string]*ParticipantSummary)
	for _, p := range combat.Participants {
		char, err := s.characterStore.Get(ctx, p.CharacterID)
		name := p.CharacterID
		finalHP := 0
		maxHP := 0
		survived := true

		if err == nil {
			name = char.Name
			if char.HP != nil {
				finalHP = char.HP.Current
				maxHP = char.HP.Max
			}
			survived = !char.IsDead()
		}

		participantStats[p.CharacterID] = &ParticipantSummary{
			CharacterID:     p.CharacterID,
			CharacterName:   name,
			DamageDealt:     0,
			DamageTaken:     0,
			HealingReceived: 0,
			Kills:           0,
			FinalHP:         finalHP,
			MaxHP:           maxHP,
			Survived:        survived,
		}
	}

	// 解析战斗日志
	for _, entry := range combat.Log {
		// 跳过系统日志（无行动者）
		if entry.ActorID == "" {
			continue
		}

		actorStats, actorExists := participantStats[entry.ActorID]
		if !actorExists {
			continue
		}

		switch entry.Action {
		case "attack", "critical_hit":
			// 解析伤害
			damage := ParseDamageFromResult(entry.Result)
			actorStats.DamageDealt += damage

			// 更新目标承受的伤害
			if entry.TargetID != "" {
				if targetStats, targetExists := participantStats[entry.TargetID]; targetExists {
					targetStats.DamageTaken += damage
				}
			}

		case "spell_cast":
			// 解析伤害（非法术）
			damage := ParseDamageFromResult(entry.Result)
			actorStats.DamageDealt += damage

			// 更新目标承受的伤害
			if entry.TargetID != "" {
				if targetStats, targetExists := participantStats[entry.TargetID]; targetExists {
					targetStats.DamageTaken += damage
				}
			}

		case "healing_spell":
			// 解析治疗
			healing := ParseHealingFromResult(entry.Result)
			actorStats.HealingReceived += healing
		}
	}

	// 转换为切片
	for _, stats := range participantStats {
		summary.Participants = append(summary.Participants, *stats)
	}

	return summary
}

// ParseDamageFromResult 从日志结果字符串中解析伤害值
// 导出以供测试使用
func ParseDamageFromResult(result string) int {
	// 简单解析：查找 "dealing X damage" 或 "hit for X damage" 模式
	// 这只是一个基本实现，实际可能需要更复杂的解析

	// 查找 "for %d damage" 模式
	for i := 0; i < len(result)-10; i++ {
		if result[i:i+4] == "for " {
			// 找到 "for "，尝试解析数字
			j := i + 4
			damage := 0
			for j < len(result) && result[j] >= '0' && result[j] <= '9' {
				damage = damage*10 + int(result[j]-'0')
				j++
			}
			if damage > 0 && j < len(result) && result[j:j+7] == " damage" {
				return damage
			}
		}
		// 查找 "dealt %d " 模式
		if i+6 <= len(result) && result[i:i+6] == "dealt " {
			j := i + 6
			damage := 0
			for j < len(result) && result[j] >= '0' && result[j] <= '9' {
				damage = damage*10 + int(result[j]-'0')
				j++
			}
			if damage > 0 {
				return damage
			}
		}
	}
	return 0
}

// ParseHealingFromResult 从日志结果字符串中解析治疗量
// 导出以供测试使用
func ParseHealingFromResult(result string) int {
	// 查找 "healed %d" 模式
	for i := 0; i < len(result)-7; i++ {
		if i+7 <= len(result) && result[i:i+7] == "healed " {
			j := i + 7
			healing := 0
			for j < len(result) && result[j] >= '0' && result[j] <= '9' {
				healing = healing*10 + int(result[j]-'0')
				j++
			}
			if healing > 0 {
				return healing
			}
		}
	}
	return 0
}
