// Package service provides business logic layer implementations
package service

import (
	"context"
	"fmt"

	"github.com/dnd-mcp/server/internal/models"
)

// CharacterStoreForCondition defines the character store interface needed by condition service
type CharacterStoreForCondition interface {
	Get(ctx context.Context, id string) (*models.Character, error)
	Update(ctx context.Context, character *models.Character) error
}

// ConditionService provides condition effect business logic
// 规则参考: PHB 附录A - Conditions
type ConditionService struct {
	characterStore CharacterStoreForCondition
}

// NewConditionService creates a new condition service
func NewConditionService(characterStore CharacterStoreForCondition) *ConditionService {
	return &ConditionService{
		characterStore: characterStore,
	}
}

// ApplyConditionRequest 应用状态效果请求
type ApplyConditionRequest struct {
	CampaignID     string `json:"campaign_id"`     // 战役ID（用于验证）
	CharacterID    string `json:"character_id"`    // 角色ID
	ConditionType  string `json:"condition_type"`  // 状态类型
	Duration       int    `json:"duration"`        // 持续回合数，-1表示永久
	Source         string `json:"source"`          // 来源（可选，如"spider poison"）
	ExhaustionLevel int    `json:"exhaustion_level"` // 力竭等级（仅用于exhaustion状态）
}

// ApplyConditionResponse 应用状态效果响应
type ApplyConditionResponse struct {
	Character  *models.Character `json:"character"`
	Applied    bool              `json:"applied"`
	Conditions []models.Condition `json:"conditions"`
	Message    string            `json:"message"`
}

// ApplyCondition 应用状态效果到角色
// 规则参考: PHB 附录A - Conditions
func (s *ConditionService) ApplyCondition(ctx context.Context, req *ApplyConditionRequest) (*ApplyConditionResponse, error) {
	// 1. 参数验证
	if req.CharacterID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "character_id is required")
	}
	if req.ConditionType == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "condition_type is required")
	}

	// 2. 验证状态类型
	if !models.IsValidConditionType(req.ConditionType) {
		return nil, NewServiceError(ErrCodeInvalidInput, fmt.Sprintf("invalid condition type: %s", req.ConditionType))
	}

	// 3. 特殊处理力竭状态
	if req.ConditionType == models.ConditionExhaustion {
		if req.ExhaustionLevel < 1 || req.ExhaustionLevel > 6 {
			return nil, NewServiceError(ErrCodeInvalidInput, "exhaustion_level must be between 1 and 6")
		}
	}

	// 4. 获取角色
	character, err := s.characterStore.Get(ctx, req.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// 5. 检查角色是否免疫该状态
	if character.Traits != nil && character.Traits.HasConditionImmunity(req.ConditionType) {
		return &ApplyConditionResponse{
			Character:  character,
			Applied:    false,
			Conditions: character.Conditions,
			Message:    fmt.Sprintf("Character is immune to %s", req.ConditionType),
		}, nil
	}

	// 6. 检查是否已有该状态（力竭特殊处理：取最高等级）
	existingIndex := -1
	for i, cond := range character.Conditions {
		if cond.Type == req.ConditionType {
			existingIndex = i
			break
		}
	}

	if req.ConditionType == models.ConditionExhaustion {
		// 力竭状态：取最高等级
		if existingIndex >= 0 {
			// 已有力竭，更新为最高等级
			level := req.ExhaustionLevel
			if level <= 0 {
				level = 1 // 默认1级
			}

			// 从现有条件获取当前等级
			currentLevel := models.ExtractExhaustionLevel(character.Conditions[existingIndex].Source)
			if level > currentLevel {
				// 使用新的来源或保持现有来源（去掉旧的等级标记）
				source := req.Source
				if source == "" {
					// 从现有 source 中提取基础来源（去掉等级标记）
					source = character.Conditions[existingIndex].Source
					// 简单的去除等级标记的处理
					for i := len(source) - 1; i >= 0; i-- {
						if source[i] == '(' && i > 2 && source[i-1] == ' ' {
							source = source[:i-1]
							break
						}
					}
				}
				character.Conditions[existingIndex].Source = fmt.Sprintf("%s (Level %d)", source, level)
				character.Conditions[existingIndex].Duration = req.Duration
			}
		} else {
			// 新增力竭状态
			source := req.Source
			if source == "" {
				source = "exhaustion"
			}
			level := req.ExhaustionLevel
			if level <= 0 {
				level = 1
			}
			character.Conditions = append(character.Conditions, models.Condition{
				Type:     models.ConditionExhaustion,
				Duration: req.Duration,
				Source:   fmt.Sprintf("%s (Level %d)", source, level),
			})
		}
	} else {
		// 普通状态：使用 Character 的 AddCondition 方法
		duration := req.Duration
		if duration == 0 && req.ConditionType != models.ConditionExhaustion {
			duration = -1 // 默认永久
		}
		source := req.Source
		if source == "" {
			source = req.ConditionType
		}
		character.AddCondition(req.ConditionType, duration, source)
	}

	// 7. 保存角色
	if err := s.characterStore.Update(ctx, character); err != nil {
		return nil, fmt.Errorf("failed to update character: %w", err)
	}

	return &ApplyConditionResponse{
		Character:  character,
		Applied:    true,
		Conditions: character.Conditions,
		Message:    fmt.Sprintf("Applied %s to character", req.ConditionType),
	}, nil
}

// RemoveConditionRequest 移除状态效果请求
type RemoveConditionRequest struct {
	CampaignID    string `json:"campaign_id"`    // 战役ID（用于验证）
	CharacterID   string `json:"character_id"`   // 角色ID
	ConditionType string `json:"condition_type"` // 状态类型
	RemoveAll     bool   `json:"remove_all"`     // 是否移除所有状态
}

// RemoveConditionResponse 移除状态效果响应
type RemoveConditionResponse struct {
	Character  *models.Character `json:"character"`
	Removed    bool              `json:"removed"`
	Conditions []models.Condition `json:"conditions"`
	Message    string            `json:"message"`
}

// RemoveCondition 移除角色的状态效果
// 规则参考: PHB 附录A - Conditions
func (s *ConditionService) RemoveCondition(ctx context.Context, req *RemoveConditionRequest) (*RemoveConditionResponse, error) {
	// 1. 参数验证
	if req.CharacterID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "character_id is required")
	}

	// 2. 获取角色
	character, err := s.characterStore.Get(ctx, req.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	removed := false
	var message string

	// 3. 移除状态
	if req.RemoveAll {
		// 移除所有状态
		if len(character.Conditions) > 0 {
			character.Conditions = make([]models.Condition, 0)
			removed = true
			message = "Removed all conditions"
		} else {
			message = "No conditions to remove"
		}
	} else if req.ConditionType != "" {
		// 移除特定状态
		removed = character.RemoveCondition(req.ConditionType)
		if removed {
			message = fmt.Sprintf("Removed %s", req.ConditionType)
		} else {
			message = fmt.Sprintf("Character does not have %s", req.ConditionType)
		}
	} else {
		return nil, NewServiceError(ErrCodeInvalidInput, "either condition_type or remove_all must be specified")
	}

	// 4. 保存角色
	if removed {
		if err := s.characterStore.Update(ctx, character); err != nil {
			return nil, fmt.Errorf("failed to update character: %w", err)
		}
	}

	return &RemoveConditionResponse{
		Character:  character,
		Removed:    removed,
		Conditions: character.Conditions,
		Message:    message,
	}, nil
}

// GetConditionEffectsRequest 获取状态效果请求
type GetConditionEffectsRequest struct {
	CampaignID  string `json:"campaign_id"`  // 战役ID（用于验证）
	CharacterID string `json:"character_id"` // 角色ID
}

// ConditionEffectInfo 状态效果信息
type ConditionEffectInfo struct {
	Condition     string              `json:"condition"`      // 状态类型
	Duration      int                 `json:"duration"`       // 剩余持续时间
	Source        string              `json:"source"`         // 来源
	Effects       []string            `json:"effects"`        // 效果描述
	Disadvantages []string            `json:"disadvantages"`  // 劣势检定列表
	Advantages    []string            `json:"advantages"`     // 优势检定列表
	Immunities    []string            `json:"immunities"`     // 免疫列表
	OtherEffects  map[string]string   `json:"other_effects"`  // 其他效果
}

// GetConditionEffectsResponse 获取状态效果响应
type GetConditionEffectsResponse struct {
	CharacterID      string                 `json:"character_id"`
	CharacterName    string                 `json:"character_name"`
	Conditions       []ConditionEffectInfo  `json:"conditions"`
	TotalDisadvantage []string              `json:"total_disadvantage"` // 所有劣势检定
	TotalAdvantage   []string               `json:"total_advantage"`    // 所有优势检定
	CannotAct        bool                   `json:"cannot_act"`         // 无法行动
	Incapped         bool                   `json:"incapped"`           // 无法施展法术
}

// GetConditionEffects 获取角色的所有状态效果及影响
// 规则参考: PHB 附录A - Conditions
func (s *ConditionService) GetConditionEffects(ctx context.Context, req *GetConditionEffectsRequest) (*GetConditionEffectsResponse, error) {
	// 1. 参数验证
	if req.CharacterID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "character_id is required")
	}

	// 2. 获取角色
	character, err := s.characterStore.Get(ctx, req.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// 3. 构建状态效果信息
	conditionInfos := make([]ConditionEffectInfo, 0)
	allDisadvantages := make(map[string]bool)
	allAdvantages := make(map[string]bool)
	cannotAct := false
	incapped := false

	for _, cond := range character.Conditions {
		effect := models.GetConditionEffect(cond.Type)
		info := ConditionEffectInfo{
			Condition:     cond.Type,
			Duration:      cond.Duration,
			Source:        cond.Source,
			Effects:       effect.Effects,
			Disadvantages: effect.Disadvantages,
			Advantages:    effect.Advantages,
			Immunities:    effect.Immunities,
			OtherEffects:  effect.OtherEffects,
		}

		// 汇总劣势
		for _, d := range effect.Disadvantages {
			allDisadvantages[d] = true
		}

		// 汇总优势
		for _, a := range effect.Advantages {
			allAdvantages[a] = true
		}

		// 检查特殊效果
		if cond.Type == models.ConditionIncapacitated ||
		   cond.Type == models.ConditionParalyzed ||
		   cond.Type == models.ConditionPetrified ||
		   cond.Type == models.ConditionStunned ||
		   cond.Type == models.ConditionUnconscious {
			cannotAct = true
		}

		// 检查法术失效
		if cond.Type == models.ConditionIncapacitated {
			incapped = true
		}

		conditionInfos = append(conditionInfos, info)
	}

	// 转换 map 到 slice
	totalDisadvantage := make([]string, 0, len(allDisadvantages))
	for d := range allDisadvantages {
		totalDisadvantage = append(totalDisadvantage, d)
	}

	totalAdvantage := make([]string, 0, len(allAdvantages))
	for a := range allAdvantages {
		totalAdvantage = append(totalAdvantage, a)
	}

	return &GetConditionEffectsResponse{
		CharacterID:       character.ID,
		CharacterName:     character.Name,
		Conditions:        conditionInfos,
		TotalDisadvantage: totalDisadvantage,
		TotalAdvantage:    totalAdvantage,
		CannotAct:         cannotAct,
		Incapped:          incapped,
	}, nil
}

// HasConditionRequest 检查状态请求
type HasConditionRequest struct {
	CharacterID   string `json:"character_id"`   // 角色ID
	ConditionType string `json:"condition_type"` // 状态类型
}

// HasConditionResponse 检查状态响应
type HasConditionResponse struct {
	HasCondition bool   `json:"has_condition"`
	Message      string `json:"message"`
}

// HasCondition 检查角色是否有特定状态
func (s *ConditionService) HasCondition(ctx context.Context, req *HasConditionRequest) (*HasConditionResponse, error) {
	if req.CharacterID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "character_id is required")
	}
	if req.ConditionType == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "condition_type is required")
	}

	character, err := s.characterStore.Get(ctx, req.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	has := character.HasCondition(req.ConditionType)
	message := fmt.Sprintf("Character %s have %s", map[bool]string{true: "does", false: "does not"}[has], req.ConditionType)

	return &HasConditionResponse{
		HasCondition: has,
		Message:      message,
	}, nil
}
