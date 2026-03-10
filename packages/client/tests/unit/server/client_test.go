// Package server_test 测试 Server Client
package server_test

import (
	"context"
	"testing"
	"time"

	"github.com/dnd-mcp/client/internal/server"
)

// 辅助函数：检查消息切片是否包含指定 ID
func containsMessage(messages []server.Message, id string) bool {
	for _, m := range messages {
		if m.ID == id {
			return true
		}
	}
	return false
}

func TestMockClient_GetContext(t *testing.T) {
	client := server.NewMockClient()
	ctx := context.Background()

	tests := []struct {
		name          string
		campaignID    string
		limit         int
		includeCombat bool
		wantErr       bool
	}{
		{
			name:          "成功获取上下文",
			campaignID:    "campaign-1",
			limit:         20,
			includeCombat: true,
			wantErr:       false,
		},
		{
			name:          "不包含战斗信息",
			campaignID:    "campaign-1",
			limit:         10,
			includeCombat: false,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			result, err := client.GetContext(ctx, tt.campaignID, tt.limit, tt.includeCombat)

			if tt.wantErr {
				if err == nil {
					t.Error("期望返回错误，但没有")
				}
				return
			}

			if err != nil {
				t.Errorf("GetContext() 错误 = %v", err)
				return
			}

			if result == nil {
				t.Error("GetContext() 返回 nil")
				return
			}

			if result.CampaignID != tt.campaignID {
				t.Errorf("CampaignID = %v, want %v", result.CampaignID, tt.campaignID)
			}

			if result.GameSummary == nil {
				t.Error("GameSummary 不应为 nil")
			}

			if len(result.Messages) == 0 {
				t.Error("Messages 不应为空")
			}
		})
	}
}

func TestMockClient_GetRawContext(t *testing.T) {
	client := server.NewMockClient()
	ctx := context.Background()

	tests := []struct {
		name       string
		campaignID string
		wantErr    bool
	}{
		{
			name:       "成功获取原始上下文",
			campaignID: "campaign-1",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			result, err := client.GetRawContext(ctx, tt.campaignID)

			if tt.wantErr {
				if err == nil {
					t.Error("期望返回错误，但没有")
				}
				return
			}

			if err != nil {
				t.Errorf("GetRawContext() 错误 = %v", err)
				return
			}

			if result == nil {
				t.Error("GetRawContext() 返回 nil")
				return
			}

			if result.CampaignID != tt.campaignID {
				t.Errorf("CampaignID = %v, want %v", result.CampaignID, tt.campaignID)
			}

			// 验证返回的数据结构
			if result.GameState == nil {
				t.Error("GameState 不应为 nil")
			}

			if len(result.Characters) == 0 {
				t.Error("Characters 不应为空")
			}

			if len(result.Messages) == 0 {
				t.Error("Messages 不应为空")
			}
		})
	}
}

func TestMockClient_SaveMessage(t *testing.T) {
	client := server.NewMockClient()
	ctx := context.Background()

	tests := []struct {
		name    string
		msg     *server.Message
		wantErr bool
	}{
		{
			name: "保存用户消息",
			msg: &server.Message{
				ID:        "msg-1",
				CampaignID: "campaign-1",
				Role:       "user",
				Content:    "Hello, world!",
				CreatedAt:  time.Now(),
			},
			wantErr: false,
		},
		{
			name: "保存助手消息",
			msg: &server.Message{
				ID:        "msg-2",
				CampaignID: "campaign-1",
				Role:       "assistant",
				Content:    "Hello! How can I help you?",
				CreatedAt:  time.Now(),
			},
			wantErr: false,
		},
		{
			name: "保存带工具调用的消息",
			msg: &server.Message{
				ID:        "msg-3",
				CampaignID: "campaign-1",
				Role:       "assistant",
				Content:    "",
				ToolCalls: []server.ToolCall{
					{
						ID:   "tc-1",
						Name: "roll_dice",
						Arguments: map[string]interface{}{
							"formula": "1d20+5",
						},
					},
				},
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.SaveMessage(ctx, tt.msg.CampaignID, tt.msg)

			if tt.wantErr {
				if err == nil {
					t.Error("期望返回错误，但没有")
				}
				return
			}

			if err != nil {
				t.Errorf("SaveMessage() 错误 = %v", err)
				return
			}

			// 验证消息被存储
			messages := client.GetMessages()
			found := false
			for _, m := range messages {
				if m.ID == tt.msg.ID {
					found = true
					break
				}
			}
			if !found {
				t.Error("消息未被存储")
			}
		})
	}
}

func TestMockClient_CallTool(t *testing.T) {
	client := server.NewMockClient()
	ctx := context.Background()

	tests := []struct {
		name     string
		toolName string
		args     map[string]any
		wantErr  bool
	}{
		{
			name:     "调用 roll_dice 工具",
			toolName: "roll_dice",
			args: map[string]any{
				"formula": "1d20+5",
			},
			wantErr: false,
		},
		{
			name:     "调用 resolve_attack 工具",
			toolName: "resolve_attack",
			args: map[string]any{
				"attacker": "char-1",
				"target":   "enemy-1",
			},
			wantErr: false,
		},
		{
			name:     "调用未知工具",
			toolName: "unknown_tool",
			args:     map[string]any{},
			wantErr:  false, // Mock 客户端不返回错误
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.CallTool(ctx, "campaign-1", tt.toolName, tt.args)

			if tt.wantErr {
				if err == nil {
					t.Error("期望返回错误，但没有")
				}
				return
			}

			if err != nil {
				t.Errorf("CallTool() 错误 = %v", err)
				return
			}

			if result == nil {
				t.Error("CallTool() 返回 nil")
				return
			}

			// 验证返回结果
			if success, ok := result["success"]; ok {
				if success != true {
					t.Error("工具调用应该成功")
				}
			}
		})
	}
}

func TestMockClient_ErrorHandling(t *testing.T) {
	client := server.NewMockClient()
	ctx := context.Background()

	// 设置返回错误
	client.SetReturnError(true)

	t.Run("GetContext 返回错误", func(t *testing.T) {
		_, err := client.GetContext(ctx, "campaign-1", 20, true)
		if err == nil {
			t.Error("期望返回错误，但没有")
		}
	})

	t.Run("GetRawContext 返回错误", func(t *testing.T) {
		_, err := client.GetRawContext(ctx, "campaign-1")
		if err == nil {
			t.Error("期望返回错误，但没有")
		}
	})

	t.Run("SaveMessage 返回错误", func(t *testing.T) {
		err := client.SaveMessage(ctx, "campaign-1", &server.Message{})
		if err == nil {
			t.Error("期望返回错误，但没有")
		}
	})

	t.Run("CallTool 返回错误", func(t *testing.T) {
		_, err := client.CallTool(ctx, "campaign-1", "test", nil)
		if err == nil {
			t.Error("期望返回错误，但没有")
		}
	})

	// 恢复正常
	client.SetReturnError(false)

	t.Run("恢复后正常工作", func(t *testing.T) {
		_, err := client.GetContext(ctx, "campaign-1", 20, true)
		if err != nil {
			t.Errorf("不应该返回错误: %v", err)
		}
	})
}

func TestMockClient_ClearMessages(t *testing.T) {
	client := server.NewMockClient()
	ctx := context.Background()

	// 保存一些消息
	_ = client.SaveMessage(ctx, "campaign-1", &server.Message{ID: "msg-1", CampaignID: "campaign-1"})
	_ = client.SaveMessage(ctx, "campaign-1", &server.Message{ID: "msg-2", CampaignID: "campaign-1"})

	if len(client.GetMessages()) != 2 {
		t.Errorf("期望 2 条消息，得到 %d", len(client.GetMessages()))
	}

	// 清除消息
	client.ClearMessages()

	if len(client.GetMessages()) != 0 {
		t.Errorf("期望 0 条消息，得到 %d", len(client.GetMessages()))
	}
}

func TestMockClient_Close(t *testing.T) {
	client := server.NewMockClient()
	ctx := context.Background()

	err := client.Close(ctx)
	if err != nil {
		t.Errorf("Close() 不应该返回错误: %v", err)
	}
}

func TestMockClient_Initialize(t *testing.T) {
	client := server.NewMockClient()
	ctx := context.Background()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "成功初始化",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.Initialize(ctx)

			if tt.wantErr {
				if err == nil {
					t.Error("期望返回错误，但没有")
				}
				return
			}

			if err != nil {
				t.Errorf("Initialize() 错误 = %v", err)
			}
		})
	}
}

func TestMockClient_Initialize_ErrorHandling(t *testing.T) {
	client := server.NewMockClient()
	ctx := context.Background()

	// 设置返回错误
	client.SetReturnError(true)

	err := client.Initialize(ctx)
	if err == nil {
		t.Error("期望返回错误，但没有")
	}

	// 恢复正常
	client.SetReturnError(false)

	err = client.Initialize(ctx)
	if err != nil {
		t.Errorf("不应该返回错误: %v", err)
	}
}
