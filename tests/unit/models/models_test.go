// Package models_test 提供领域模型单元测试
package models_test

import (
	"testing"
	"time"

	"github.com/dnd-mcp/client/internal/models"
	"github.com/google/uuid"
)

func TestNewSession(t *testing.T) {
	name := "测试会话"
	creatorID := "user-123"
	mcpURL := "http://localhost:9000"

	session := models.NewSession(name, creatorID, mcpURL)

	// 验证基本字段
	if session.Name != name {
		t.Errorf("Name 错误: got %s, want %s", session.Name, name)
	}
	if session.CreatorID != creatorID {
		t.Errorf("CreatorID 错误: got %s, want %s", session.CreatorID, creatorID)
	}
	if session.MCPServerURL != mcpURL {
		t.Errorf("MCPServerURL 错误: got %s, want %s", session.MCPServerURL, mcpURL)
	}

	// 验证默认值
	if session.Status != "active" {
		t.Errorf("Status 默认值错误: got %s, want active", session.Status)
	}
	if session.MaxPlayers != 4 {
		t.Errorf("MaxPlayers 默认值错误: got %d, want 4", session.MaxPlayers)
	}

	// 验证时间
	if time.Since(session.CreatedAt) > time.Second {
		t.Error("CreatedAt 应该是最近的时间")
	}
	if time.Since(session.UpdatedAt) > time.Second {
		t.Error("UpdatedAt 应该是最近的时间")
	}

	// 验证 settings 已初始化
	if session.Settings == nil {
		t.Error("Settings 应该被初始化")
	}
}

func TestSessionIsActive(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{
			name:   "活跃会话",
			status: "active",
			want:   true,
		},
		{
			name:   "归档会话",
			status: "archived",
			want:   false,
		},
		{
			name:   "其他状态",
			status: "deleted",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &models.Session{
				ID:     uuid.New().String(),
				Status: tt.status,
			}
			if got := session.IsActive(); got != tt.want {
				t.Errorf("Session.IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSessionArchive(t *testing.T) {
	session := models.NewSession("测试会话", "user-123", "http://localhost:9000")
	if !session.IsActive() {
		t.Error("新会话应该是活跃状态")
	}

	session.Archive()

	if session.Status != "archived" {
		t.Errorf("Archive() 后状态应该是 archived, got %s", session.Status)
	}
}

func TestSessionUpdateSettings(t *testing.T) {
	session := models.NewSession("测试会话", "user-123", "http://localhost:9000")

	// 添加设置
	newSettings := map[string]interface{}{
		"ruleset":        "dnd5e",
		"starting_level": 5,
	}
	session.UpdateSettings(newSettings)

	// 验证设置
	if session.Settings["ruleset"] != "dnd5e" {
		t.Errorf("settings[\"ruleset\"] 错误: got %v, want dnd5e", session.Settings["ruleset"])
	}
	if session.Settings["starting_level"] != 5 {
		t.Errorf("settings[\"starting_level\"] 错误: got %v, want 5", session.Settings["starting_level"])
	}

	// 更新设置
	updatedSettings := map[string]interface{}{
		"starting_level": 10,
	}
	session.UpdateSettings(updatedSettings)

	// 验证原有设置保留,新设置更新
	if session.Settings["ruleset"] != "dnd5e" {
		t.Error("原有设置应该被保留")
	}
	if session.Settings["starting_level"] != 10 {
		t.Errorf("设置应该被更新: got %v, want 10", session.Settings["starting_level"])
	}
}

func TestNewMessage(t *testing.T) {
	sessionID := uuid.New().String()
	role := "user"
	content := "这是一条测试消息"

	message := models.NewMessage(sessionID, role, content)

	// 验证基本字段
	if message.SessionID != sessionID {
		t.Errorf("SessionID 错误: got %s, want %s", message.SessionID, sessionID)
	}
	if message.Role != role {
		t.Errorf("Role 错误: got %s, want %s", message.Role, role)
	}
	if message.Content != content {
		t.Errorf("Content 错误: got %s, want %s", message.Content, content)
	}

	// 验证 ID 已生成
	if message.ID == "" {
		t.Error("ID 应该被自动生成")
	}

	// 验证时间
	if time.Since(message.CreatedAt) > time.Second {
		t.Error("CreatedAt 应该是最近的时间")
	}
}

func TestNewUserMessage(t *testing.T) {
	sessionID := uuid.New().String()
	content := "用户消息"
	playerID := "player-123"

	message := models.NewUserMessage(sessionID, content, playerID)

	if message.Role != "user" {
		t.Errorf("Role 应该是 user, got %s", message.Role)
	}
	if message.PlayerID != playerID {
		t.Errorf("PlayerID 错误: got %s, want %s", message.PlayerID, playerID)
	}
}

func TestNewAssistantMessage(t *testing.T) {
	sessionID := uuid.New().String()
	content := "助手响应"

	message := models.NewAssistantMessage(sessionID, content)

	if message.Role != "assistant" {
		t.Errorf("Role 应该是 assistant, got %s", message.Role)
	}
}

func TestNewSystemMessage(t *testing.T) {
	sessionID := uuid.New().String()
	content := "系统通知"

	message := models.NewSystemMessage(sessionID, content)

	if message.Role != "system" {
		t.Errorf("Role 应该是 system, got %s", message.Role)
	}
}

func TestMessageHasToolCalls(t *testing.T) {
	message := models.NewMessage("session-1", "assistant", "响应")

	// 没有工具调用
	if message.HasToolCalls() {
		t.Error("新消息不应该有工具调用")
	}

	// 添加工具调用
	toolCall := models.ToolCall{
		ID:   "call-1",
		Name: "roll_dice",
		Arguments: map[string]interface{}{
			"formula": "1d20",
		},
	}
	message.AddToolCall(toolCall)

	// 有工具调用
	if !message.HasToolCalls() {
		t.Error("添加工具调用后应该返回 true")
	}
}

func TestMessageAddToolCall(t *testing.T) {
	message := models.NewMessage("session-1", "assistant", "响应")

	// 添加第一个工具调用
	toolCall1 := models.ToolCall{
		ID:        "call-1",
		Name:      "roll_dice",
		Arguments: map[string]interface{}{"formula": "1d20"},
	}
	message.AddToolCall(toolCall1)

	if len(message.ToolCalls) != 1 {
		t.Errorf("应该有 1 个工具调用, got %d", len(message.ToolCalls))
	}

	// 添加第二个工具调用
	toolCall2 := models.ToolCall{
		ID:        "call-2",
		Name:      "resolve_attack",
		Arguments: map[string]interface{}{"target": "goblin"},
	}
	message.AddToolCall(toolCall2)

	if len(message.ToolCalls) != 2 {
		t.Errorf("应该有 2 个工具调用, got %d", len(message.ToolCalls))
	}

	// 验证工具调用内容
	if message.ToolCalls[0].Name != "roll_dice" {
		t.Errorf("第一个工具调用名称错误: got %s, want roll_dice", message.ToolCalls[0].Name)
	}
	if message.ToolCalls[1].Name != "resolve_attack" {
		t.Errorf("第二个工具调用名称错误: got %s, want resolve_attack", message.ToolCalls[1].Name)
	}
}
