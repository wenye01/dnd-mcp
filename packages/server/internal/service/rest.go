// Package service provides business logic layer implementations
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/store"
)

// RestService provides rest-related business logic
// 规则参考: PHB 第8章 Resting
type RestService struct {
	characterStore CharacterStore
	gameStateStore store.GameStateStore
}

// NewRestService creates a new rest service
func NewRestService(characterStore CharacterStore, gameStateStore store.GameStateStore) *RestService {
	return &RestService{
		characterStore: characterStore,
		gameStateStore: gameStateStore,
	}
}

// ShortRestRequest 短休请求
type ShortRestRequest struct {
	CampaignID      string `json:"campaign_id"`      // 战役ID
	CharacterID     string `json:"character_id"`     // 角色ID
	HitDiceToSpend  int    `json:"hit_dice_to_spend"` // 要消耗的生命骰数量
}

// ShortRestResponse 短休响应
type ShortRestResponse struct {
	CharacterID       string `json:"character_id"`        // 角色ID
	CharacterName     string `json:"character_name"`      // 角色名称
	HitDiceSpent      int    `json:"hit_dice_spent"`      // 消耗的生命骰数量
	HPHealed          int    `json:"hp_healed"`           // 恢复的HP
	HPCurrent         int    `json:"hp_current"`          // 当前HP
	HPMax             int    `json:"hp_max"`              // 最大HP
	HitDiceRemaining  int    `json:"hit_dice_remaining"`  // 剩余生命骰
	ConstitutionMod   int    `json:"constitution_mod"`    // 体质修正值
}

// LongRestRequest 长休请求
type LongRestRequest struct {
	CampaignID  string `json:"campaign_id"`  // 战役ID
	CharacterID string `json:"character_id"` // 角色ID
}

// LongRestResponse 长休响应
type LongRestResponse struct {
	CharacterID         string    `json:"character_id"`          // 角色ID
	CharacterName       string    `json:"character_name"`        // 角色名称
	HPHealed            int       `json:"hp_healed"`             // 恢复的HP
	HPCurrent           int       `json:"hp_current"`            // 当前HP
	HPMax               int       `json:"hp_max"`                // 最大HP
	SpellSlotsRestored  bool      `json:"spell_slots_restored"`  // 是否恢复了法术位
	HitDiceRestored     int       `json:"hit_dice_restored"`     // 恢复的生命骰数量
	HitDiceRemaining    int       `json:"hit_dice_remaining"`    // 剩余生命骰
	FeaturesRestored    []string  `json:"features_restored"`     // 恢复的特性名称
	LastLongRest        time.Time `json:"last_long_rest"`        // 上次长休时间
	ConstitutionMod     int       `json:"constitution_mod"`      // 体质修正值
}

// ShortRest 执行短休
// 规则参考: PHB 第8章 - Short Rest
// 短休持续至少1小时，角色可以消耗生命骰恢复HP
func (s *RestService) ShortRest(ctx context.Context, req *ShortRestRequest) (*ShortRestResponse, error) {
	// 1. 参数验证
	if req.CampaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}
	if req.CharacterID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "character ID is required")
	}

	// 2. 获取角色
	character, err := s.characterStore.GetByCampaignAndID(ctx, req.CampaignID, req.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// 3. 计算体质修正
	conMod := 0
	if character.Abilities != nil {
		conMod = GetAbilityModifier(character.Abilities.Constitution)
	}

	// 4. 执行短休
	result, err := character.ShortRest(req.HitDiceToSpend, conMod)
	if err != nil {
		return nil, fmt.Errorf("short rest failed: %w", err)
	}

	// 5. 保存更新
	if err := s.characterStore.Update(ctx, character); err != nil {
		return nil, fmt.Errorf("failed to update character: %w", err)
	}

	// 6. 返回响应
	return &ShortRestResponse{
		CharacterID:      character.ID,
		CharacterName:    character.Name,
		HitDiceSpent:     result.HitDiceSpent,
		HPHealed:         result.HPHealed,
		HPCurrent:        character.HP.Current,
		HPMax:            character.HP.Max,
		HitDiceRemaining: result.HitDiceRemaining,
		ConstitutionMod:  conMod,
	}, nil
}

// LongRest 执行长休
// 规则参考: PHB 第8章 - Long Rest
// 长休持续至少8小时，角色恢复全部HP、一半生命骰、所有法术位
func (s *RestService) LongRest(ctx context.Context, req *LongRestRequest) (*LongRestResponse, error) {
	// 1. 参数验证
	if req.CampaignID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}
	if req.CharacterID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "character ID is required")
	}

	// 2. 获取角色
	character, err := s.characterStore.GetByCampaignAndID(ctx, req.CampaignID, req.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// 3. 计算体质修正
	conMod := 0
	if character.Abilities != nil {
		conMod = GetAbilityModifier(character.Abilities.Constitution)
	}

	// 4. 执行长休
	result := character.LongRest(conMod)

	// 5. 获取并更新游戏状态（记录上次长休时间）
	if s.gameStateStore != nil {
		gameState, err := s.gameStateStore.Get(ctx, req.CampaignID)
		if err == nil && gameState != nil {
			// 游戏状态存在，更新时间（游戏时间前进8小时）
			if gameState.GameTime != nil {
				gameState.GameTime.AddHours(8)
				gameState.UpdatedAt = time.Now()
				_ = s.gameStateStore.Update(ctx, gameState) // 忽略更新错误，因为休息已经完成
			}
		}
	}

	// 6. 保存角色更新
	if err := s.characterStore.Update(ctx, character); err != nil {
		return nil, fmt.Errorf("failed to update character: %w", err)
	}

	// 7. 返回响应
	now := time.Now()
	return &LongRestResponse{
		CharacterID:        character.ID,
		CharacterName:      character.Name,
		HPHealed:           result.HPHealed,
		HPCurrent:          character.HP.Current,
		HPMax:              character.HP.Max,
		SpellSlotsRestored: result.SpellSlotsRestored,
		HitDiceRestored:    result.HitDiceRestored,
		HitDiceRemaining:   result.HitDiceRemaining,
		FeaturesRestored:   result.FeaturesRestored,
		LastLongRest:       now,
		ConstitutionMod:    conMod,
	}, nil
}

// PartyLongRest 执行队伍长休（所有玩家角色）
// 规则参考: PHB 第8章 - Long Rest
func (s *RestService) PartyLongRest(ctx context.Context, campaignID string) (*models.GameState, []*LongRestResponse, error) {
	// 1. 参数验证
	if campaignID == "" {
		return nil, nil, NewServiceError(ErrCodeInvalidInput, "campaign ID is required")
	}

	// 2. 获取所有玩家角色
	filter := &store.CharacterFilter{
		CampaignID: campaignID,
		IsNPC:      boolPtr(false), // 只获取玩家角色
	}
	characters, err := s.characterStore.List(ctx, filter)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list characters: %w", err)
	}

	// 3. 为每个角色执行长休
	responses := make([]*LongRestResponse, 0, len(characters))
	for _, character := range characters {
		req := &LongRestRequest{
			CampaignID:  campaignID,
			CharacterID: character.ID,
		}

		resp, err := s.LongRest(ctx, req)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to long rest for character %s: %w", character.ID, err)
		}

		responses = append(responses, resp)
	}

	// 4. 获取更新后的游戏状态
	var gameState *models.GameState
	if s.gameStateStore != nil {
		gameState, err = s.gameStateStore.Get(ctx, campaignID)
		if err != nil {
			return nil, responses, nil // 游戏状态不存在，返回结果即可
		}
	}

	return gameState, responses, nil
}

// boolPtr 返回bool指针的辅助函数
func boolPtr(b bool) *bool {
	return &b
}
