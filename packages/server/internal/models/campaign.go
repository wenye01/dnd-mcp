// Package models 提供领域模型定义
package models

import (
	"time"
)

// CampaignStatus 战役状态
type CampaignStatus string

const (
	// CampaignStatusActive 进行中
	CampaignStatusActive CampaignStatus = "active"
	// CampaignStatusPaused 暂停
	CampaignStatusPaused CampaignStatus = "paused"
	// CampaignStatusFinished 已结束
	CampaignStatusFinished CampaignStatus = "finished"
	// CampaignStatusArchived 已归档
	CampaignStatusArchived CampaignStatus = "archived"
)

// CampaignSettings 战役设置
type CampaignSettings struct {
	MaxPlayers    int                    `json:"max_players"`     // 最大玩家数，默认4
	StartLevel    int                    `json:"start_level"`     // 起始等级，默认1
	Ruleset       string                 `json:"ruleset"`         // 规则集，默认 "dnd5e"
	HouseRules    map[string]interface{} `json:"house_rules"`     // 房规
	ContextWindow int                    `json:"context_window"`  // 上下文窗口大小，默认20
}

// NewCampaignSettings 创建默认战役设置
func NewCampaignSettings() *CampaignSettings {
	return &CampaignSettings{
		MaxPlayers:    4,
		StartLevel:    1,
		Ruleset:       "dnd5e",
		HouseRules:    make(map[string]interface{}),
		ContextWindow: 20,
	}
}

// Validate 验证战役设置
func (s *CampaignSettings) Validate() error {
	if s.MaxPlayers < 1 || s.MaxPlayers > 10 {
		return NewValidationError("max_players", "must be between 1 and 10")
	}
	if s.StartLevel < 1 || s.StartLevel > 20 {
		return NewValidationError("start_level", "must be between 1 and 20")
	}
	if s.ContextWindow < 1 {
		return NewValidationError("context_window", "must be at least 1")
	}
	return nil
}

// Campaign 战役实体
type Campaign struct {
	ID          string            `json:"id"`           // UUID
	Name        string            `json:"name"`         // 战役名称
	Description string            `json:"description"`  // 战役描述
	DMID        string            `json:"dm_id"`        // DM（地下城主）用户ID
	Settings    *CampaignSettings `json:"settings"`     // 战役设置
	Status      CampaignStatus    `json:"status"`       // 状态
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	DeletedAt   *time.Time        `json:"deleted_at,omitempty"` // 软删除时间
}

// NewCampaign 创建新战役
func NewCampaign(name, dmID, description string) *Campaign {
	now := time.Now()
	return &Campaign{
		Name:        name,
		Description: description,
		DMID:        dmID,
		Settings:    NewCampaignSettings(),
		Status:      CampaignStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Validate 验证战役数据
func (c *Campaign) Validate() error {
	if c.Name == "" {
		return NewValidationError("name", "cannot be empty")
	}
	if c.DMID == "" {
		return NewValidationError("dm_id", "cannot be empty")
	}
	if c.Settings != nil {
		if err := c.Settings.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// IsActive 检查战役是否活跃
func (c *Campaign) IsActive() bool {
	return c.Status == CampaignStatusActive
}

// IsPaused 检查战役是否暂停
func (c *Campaign) IsPaused() bool {
	return c.Status == CampaignStatusPaused
}

// IsFinished 检查战役是否已结束
func (c *Campaign) IsFinished() bool {
	return c.Status == CampaignStatusFinished
}

// IsArchived 检查战役是否已归档
func (c *Campaign) IsArchived() bool {
	return c.Status == CampaignStatusArchived
}

// Pause 暂停战役
func (c *Campaign) Pause() {
	c.Status = CampaignStatusPaused
	c.UpdatedAt = time.Now()
}

// Resume 恢复战役
func (c *Campaign) Resume() {
	c.Status = CampaignStatusActive
	c.UpdatedAt = time.Now()
}

// Finish 结束战役
func (c *Campaign) Finish() {
	c.Status = CampaignStatusFinished
	c.UpdatedAt = time.Now()
}

// Archive 归档战役
func (c *Campaign) Archive() {
	c.Status = CampaignStatusArchived
	now := time.Now()
	c.UpdatedAt = now
	c.DeletedAt = &now
}

// UpdateSettings 更新战役设置
func (c *Campaign) UpdateSettings(settings *CampaignSettings) error {
	if settings == nil {
		return nil
	}
	if err := settings.Validate(); err != nil {
		return err
	}
	c.Settings = settings
	c.UpdatedAt = time.Now()
	return nil
}

// UpdateDescription 更新战役描述
func (c *Campaign) UpdateDescription(description string) {
	c.Description = description
	c.UpdatedAt = time.Now()
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error 实现 error 接口
func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// NewValidationError 创建验证错误
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
