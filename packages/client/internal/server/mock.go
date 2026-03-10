// Package server 提供 Mock Server Client 实现
package server

import (
	"context"
	"time"
)

// MockClient Mock Server 客户端
type MockClient struct {
	// 存储的消息，用于测试验证
	Messages []Message
	// 是否返回错误
	ReturnError bool
}

// NewMockClient 创建 Mock 客户端
func NewMockClient() *MockClient {
	return &MockClient{
		Messages:    make([]Message, 0),
		ReturnError: false,
	}
}

// Initialize 执行 MCP 握手（Mock 版本直接返回成功）
func (m *MockClient) Initialize(ctx context.Context) error {
	if m.ReturnError {
		return &MockError{Msg: "mock initialize error"}
	}
	return nil
}

// GetContext 获取压缩后的上下文
func (m *MockClient) GetContext(ctx context.Context, campaignID string, limit int, includeCombat bool) (*Context, error) {
	if m.ReturnError {
		return nil, &MockError{Msg: "mock error"}
	}

	// 过滤出指定 campaign 的消息
	var campaignMessages []Message
	for _, msg := range m.Messages {
		if msg.CampaignID == campaignID {
			campaignMessages = append(campaignMessages, msg)
		}
	}

	// 如果没有保存的消息，返回默认欢迎消息
	if len(campaignMessages) == 0 {
		campaignMessages = []Message{
			{
				ID:         "msg-default",
				CampaignID: campaignID,
				Role:       MessageRoleSystem,
				Content:    "Welcome to the adventure!",
				CreatedAt:  time.Now().Add(-1 * time.Hour),
			},
		}
	}

	// 应用 limit
	if limit > 0 && len(campaignMessages) > limit {
		campaignMessages = campaignMessages[len(campaignMessages)-limit:]
	}

	return &Context{
		CampaignID: campaignID,
		GameSummary: &GameSummary{
			Time:      "Night, 1st of Mirtul",
			Location:  "Greenwood Inn",
			Weather:   "Clear",
			InCombat:  false,
			Party: []PartyMember{
				{ID: "char-1", Name: "Aldric", HP: "25/30", Class: "Fighter 1"},
			},
		},
		Messages:         campaignMessages,
		RawMessageCount:  len(campaignMessages),
		TokenEstimate:    len(campaignMessages) * 50,
		CreatedAt:        time.Now(),
	}, nil
}

// GetRawContext 获取原始上下文（完整模式）
func (m *MockClient) GetRawContext(ctx context.Context, campaignID string) (*RawContext, error) {
	if m.ReturnError {
		return nil, &MockError{Msg: "mock error"}
	}

	return &RawContext{
		CampaignID: campaignID,
		GameState: &GameState{
			Location:     "Greenwood Inn",
			GameTime:     "Night, 1st of Mirtul",
			LastRestTime: time.Now().Add(-2 * time.Hour),
			ShortRests:   1,
			LongRests:    0,
		},
		Characters: []*Character{
			{
				ID:         "char-1",
				Name:       "Aldric",
				Type:       "pc",
				HP:         25,
				MaxHP:      30,
				AC:         16,
				Initiative: 3,
				Speed:      30,
			},
		},
		Combat: &Combat{
			ID:     "combat-1",
			Active: false,
			Round:  0,
			Turn:   0,
		},
		Map: &Map{
			ID:     "map-1",
			Name:   "Tavern Common Room",
			Type:   "custom",
			Width:  40,
			Height: 30,
			Scale:  5,
		},
		Messages: []*Message{
			{
				ID:         "msg-1",
				CampaignID: campaignID,
				Role:       MessageRoleSystem,
				Content:    "Welcome to the adventure!",
				PlayerID:   "",
				ToolCalls:  nil,
				CreatedAt:  time.Now().Add(-1 * time.Hour),
			},
		},
		MessageCount: 1,
	}, nil
}

// SaveMessage 保存消息到 Server
func (m *MockClient) SaveMessage(ctx context.Context, campaignID string, msg *Message) error {
	if m.ReturnError {
		return &MockError{Msg: "mock error"}
	}

	// 存储消息供测试验证
	m.Messages = append(m.Messages, *msg)
	return nil
}

// CallTool 调用 Server MCP Tool
func (m *MockClient) CallTool(ctx context.Context, campaignID, toolName string, args map[string]any) (map[string]any, error) {
	if m.ReturnError {
		return nil, &MockError{Msg: "mock error"}
	}

	// 根据工具名称返回模拟结果
	switch toolName {
	case "roll_dice":
		return map[string]any{
			"success": true,
			"result": map[string]any{
				"formula":  args["formula"],
				"total":    18,
				"rolls":    []int{15},
				"modifier": 3,
			},
		}, nil

	case "resolve_attack":
		return map[string]any{
			"success": true,
			"result": map[string]any{
				"attacker": args["attacker"],
				"target":   args["target"],
				"hit":      true,
				"damage":   8,
			},
		}, nil

	default:
		return map[string]any{
			"success": true,
			"message": "tool executed: " + toolName,
		}, nil
	}
}

// Close 关闭连接
func (m *MockClient) Close(_ context.Context) error {
	return nil
}

// MockError Mock 错误类型
type MockError struct {
	Msg string
}

func (e *MockError) Error() string {
	return e.Msg
}

// SetReturnError 设置是否返回错误（用于测试错误场景）
func (m *MockClient) SetReturnError(returnError bool) {
	m.ReturnError = returnError
}

// GetMessages 获取存储的消息（用于测试验证）
func (m *MockClient) GetMessages() []Message {
	return m.Messages
}

// ClearMessages 清除存储的消息
func (m *MockClient) ClearMessages() {
	m.Messages = make([]Message, 0)
}
