// Package server_test 提供 Server Client 单元测试
package server_test

import (
	"context"
	"testing"
	"time"

	"github.com/dnd-mcp/client/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockClient_GetContext(t *testing.T) {
	ctx := context.Background()
	client := server.NewMockClient()

	t.Run("成功获取压缩上下文", func(t *testing.T) {
		result, err := client.GetContext(ctx, "campaign-123", 10, false)

		require.NoError(t, err)
		assert.Equal(t, "campaign-123", result.CampaignID)
		assert.NotNil(t, result.GameSummary)
		assert.Equal(t, "Greenwood Inn", result.GameSummary.Location)
		assert.Equal(t, "Night, 1st of Mirtul", result.GameSummary.Time)
		assert.False(t, result.GameSummary.InCombat)
		assert.NotEmpty(t, result.GameSummary.Party)
		assert.NotEmpty(t, result.Messages)
		assert.Equal(t, 1, result.RawMessageCount)
		assert.Greater(t, result.TokenEstimate, 0)
		assert.False(t, result.CreatedAt.IsZero())
	})

	t.Run("返回错误", func(t *testing.T) {
		client.SetReturnError(true)

		result, err := client.GetContext(ctx, "campaign-123", 10, false)

		assert.Error(t, err)
		assert.Nil(t, result)

		client.SetReturnError(false)
	})
}

func TestMockClient_GetRawContext(t *testing.T) {
	ctx := context.Background()
	client := server.NewMockClient()

	t.Run("成功获取原始上下文", func(t *testing.T) {
		result, err := client.GetRawContext(ctx, "campaign-456")

		require.NoError(t, err)
		assert.Equal(t, "campaign-456", result.CampaignID)
		assert.NotNil(t, result.GameState)
		assert.Equal(t, "Greenwood Inn", result.GameState.Location)
		assert.NotEmpty(t, result.Characters)
		assert.NotNil(t, result.Combat)
		assert.NotNil(t, result.Map)
		assert.NotEmpty(t, result.Messages)
		assert.Equal(t, 1, result.MessageCount)

		// 验证角色数据
		char := result.Characters[0]
		assert.Equal(t, "char-1", char.ID)
		assert.Equal(t, "Aldric", char.Name)
		assert.Equal(t, "pc", char.Type)
		assert.Equal(t, 25, char.HP)
		assert.Equal(t, 30, char.MaxHP)
		assert.Equal(t, 16, char.AC)
	})

	t.Run("返回错误", func(t *testing.T) {
		client.SetReturnError(true)

		result, err := client.GetRawContext(ctx, "campaign-456")

		assert.Error(t, err)
		assert.Nil(t, result)

		client.SetReturnError(false)
	})
}

func TestMockClient_SaveMessage(t *testing.T) {
	ctx := context.Background()
	client := server.NewMockClient()

	t.Run("成功保存消息", func(t *testing.T) {
		msg := &server.Message{
			ID:         "msg-123",
			CampaignID: "campaign-789",
			Role:       server.MessageRoleUser,
			Content:    "Hello, world!",
			PlayerID:   "player-1",
			CreatedAt:  time.Now(),
		}

		err := client.SaveMessage(ctx, "campaign-789", msg)
		require.NoError(t, err)

		// 验证消息被存储
		messages := client.GetMessages()
		assert.Len(t, messages, 1)
		assert.Equal(t, "msg-123", messages[0].ID)
		assert.Equal(t, server.MessageRoleUser, messages[0].Role)
		assert.Equal(t, "Hello, world!", messages[0].Content)
	})

	t.Run("保存多条消息", func(t *testing.T) {
		client.ClearMessages()

		messages := []*server.Message{
			{ID: "msg-1", CampaignID: "c1", Role: server.MessageRoleUser, Content: "First", PlayerID: "p1", CreatedAt: time.Now()},
			{ID: "msg-2", CampaignID: "c1", Role: server.MessageRoleAssistant, Content: "Second", PlayerID: "", CreatedAt: time.Now()},
			{ID: "msg-3", CampaignID: "c1", Role: server.MessageRoleSystem, Content: "Third", PlayerID: "", CreatedAt: time.Now()},
		}

		for _, msg := range messages {
			err := client.SaveMessage(ctx, "campaign-789", msg)
			require.NoError(t, err)
		}

		saved := client.GetMessages()
		assert.Len(t, saved, 3)
	})

	t.Run("返回错误", func(t *testing.T) {
		client.ClearMessages()
		client.SetReturnError(true)

		msg := &server.Message{
			ID:         "msg-error",
			CampaignID: "campaign-789",
			Role:       server.MessageRoleUser,
			Content:    "Error test",
			PlayerID:   "player-1",
			CreatedAt:  time.Now(),
		}

		err := client.SaveMessage(ctx, "campaign-789", msg)
		assert.Error(t, err)

		// 验证消息未被存储
		saved := client.GetMessages()
		assert.Len(t, saved, 0)

		client.SetReturnError(false)
	})
}

func TestMockClient_CallTool(t *testing.T) {
	ctx := context.Background()
	client := server.NewMockClient()

	t.Run("调用 roll_dice 工具", func(t *testing.T) {
		args := map[string]any{
			"formula": "1d20+3",
		}

		result, err := client.CallTool(ctx, "campaign-123", "roll_dice", args)

		require.NoError(t, err)
		assert.True(t, result["success"].(bool))
		assert.NotNil(t, result["result"])

		resultData := result["result"].(map[string]any)
		assert.Equal(t, "1d20+3", resultData["formula"])
		assert.Equal(t, int(18), resultData["total"])
	})

	t.Run("调用 resolve_attack 工具", func(t *testing.T) {
		args := map[string]any{
			"attacker": "char-1",
			"target":   "char-2",
		}

		result, err := client.CallTool(ctx, "campaign-123", "resolve_attack", args)

		require.NoError(t, err)
		assert.True(t, result["success"].(bool))
		assert.NotNil(t, result["result"])

		resultData := result["result"].(map[string]any)
		assert.Equal(t, "char-1", resultData["attacker"])
		assert.Equal(t, "char-2", resultData["target"])
		assert.True(t, resultData["hit"].(bool))
	})

	t.Run("调用未知工具", func(t *testing.T) {
		args := map[string]any{
			"param": "value",
		}

		result, err := client.CallTool(ctx, "campaign-123", "unknown_tool", args)

		require.NoError(t, err)
		assert.True(t, result["success"].(bool))
		assert.Contains(t, result["message"].(string), "unknown_tool")
	})

	t.Run("返回错误", func(t *testing.T) {
		client.SetReturnError(true)

		result, err := client.CallTool(ctx, "campaign-123", "roll_dice", nil)

		assert.Error(t, err)
		assert.Nil(t, result)

		client.SetReturnError(false)
	})
}

func TestMockClient_Close(t *testing.T) {
	ctx := context.Background()
	client := server.NewMockClient()

	err := client.Close(ctx)
	assert.NoError(t, err)
}

func TestMessage_ToolCalls(t *testing.T) {
	msg := &server.Message{
		ID:         "msg-with-tools",
		CampaignID: "c1",
		Role:       server.MessageRoleAssistant,
		Content:    "I'll roll for you",
		PlayerID:   "",
		ToolCalls: []server.ToolCall{
			{
				ID:   "tool-1",
				Name: "roll_dice",
				Arguments: map[string]any{
					"formula": "1d20+5",
				},
			},
		},
		CreatedAt: time.Now(),
	}

	assert.Equal(t, "msg-with-tools", msg.ID)
	assert.Len(t, msg.ToolCalls, 1)
	assert.Equal(t, "roll_dice", msg.ToolCalls[0].Name)
}

func TestGameSummary(t *testing.T) {
	t.Run("战斗中的摘要", func(t *testing.T) {
		summary := &server.GameSummary{
			Time:      "Night",
			Location:  "Dark Forest",
			Weather:   "Rainy",
			InCombat:  true,
			Party:     []server.PartyMember{{ID: "c1", Name: "Hero", HP: "10/20", Class: "Fighter"}},
			Combat:    &server.CombatSummary{Round: 2, TurnIndex: 1, Participants: []string{"Hero", "Goblin"}},
		}

		assert.True(t, summary.InCombat)
		assert.NotNil(t, summary.Combat)
		assert.Equal(t, 2, summary.Combat.Round)
	})

	t.Run("非战斗的摘要", func(t *testing.T) {
		summary := &server.GameSummary{
			Time:     "Evening",
			Location: "Tavern",
			Weather:  "Clear",
			InCombat: false,
			Party:    []server.PartyMember{{ID: "c1", Name: "Hero", HP: "20/20", Class: "Fighter"}},
		}

		assert.False(t, summary.InCombat)
		assert.Nil(t, summary.Combat)
	})
}

func TestCharacter(t *testing.T) {
	char := &server.Character{
		ID:         "char-1",
		Name:       "Test Hero",
		Type:       "pc",
		HP:         20,
		MaxHP:      20,
		AC:         15,
		Initiative: 2,
		Speed:      30,
		Stats: server.CharacterStats{
			Strength:     16,
			Dexterity:    14,
			Constitution: 15,
			Intelligence: 10,
			Wisdom:       12,
			Charisma:     13,
		},
		Conditions: []string{},
	}

	assert.Equal(t, "char-1", char.ID)
	assert.Equal(t, "Test Hero", char.Name)
	assert.Equal(t, "pc", char.Type)
	assert.Equal(t, 20, char.HP)
	assert.Equal(t, 15, char.AC)
	assert.Equal(t, 16, char.Stats.Strength)
	assert.Empty(t, char.Conditions)
}

func TestCombat(t *testing.T) {
	combat := &server.Combat{
		ID:           "combat-1",
		Active:       true,
		Round:        2,
		Turn:         3,
		CurrentActor: "char-1",
		StartedAt:    time.Now(),
	}

	assert.True(t, combat.Active)
	assert.Equal(t, 2, combat.Round)
	assert.Equal(t, 3, combat.Turn)
	assert.Equal(t, "char-1", combat.CurrentActor)
}

func TestMap(t *testing.T) {
	mapData := &server.Map{
		ID:     "map-1",
		Name:   "Battle Arena",
		Type:   "custom",
		Width:  50,
		Height: 40,
		Scale:  5,
		Tokens: []server.Token{
			{
				ID:          "token-1",
				CharacterID: "char-1",
				X:           10,
				Y:           15,
				Width:       1,
				Height:      1,
				Visible:     true,
			},
		},
	}

	assert.Equal(t, "map-1", mapData.ID)
	assert.Equal(t, "Battle Arena", mapData.Name)
	assert.Equal(t, 50, mapData.Width)
	assert.Len(t, mapData.Tokens, 1)
	assert.Equal(t, "char-1", mapData.Tokens[0].CharacterID)
}

func TestMessageRole(t *testing.T) {
	assert.Equal(t, server.MessageRole("user"), server.MessageRoleUser)
	assert.Equal(t, server.MessageRole("assistant"), server.MessageRoleAssistant)
	assert.Equal(t, server.MessageRole("system"), server.MessageRoleSystem)
}

func TestToolResult(t *testing.T) {
	t.Run("成功结果", func(t *testing.T) {
		result := &server.ToolResult{
			Success: true,
			Data: map[string]any{
				"total": 18,
				"rolls": []int{15},
			},
			Error: "",
		}

		assert.True(t, result.Success)
		assert.Empty(t, result.Error)
		assert.NotNil(t, result.Data)
	})

	t.Run("失败结果", func(t *testing.T) {
		result := &server.ToolResult{
			Success: false,
			Data:    nil,
			Error:   "invalid formula",
		}

		assert.False(t, result.Success)
		assert.NotEmpty(t, result.Error)
		assert.Nil(t, result.Data)
	})
}

func TestPartyMember(t *testing.T) {
	member := &server.PartyMember{
		ID:    "char-1",
		Name:  "Aldric",
		HP:    "25/30",
		Class: "Fighter 1",
	}

	assert.Equal(t, "char-1", member.ID)
	assert.Equal(t, "Aldric", member.Name)
	assert.Equal(t, "25/30", member.HP)
	assert.Equal(t, "Fighter 1", member.Class)
}

func TestCombatSummary(t *testing.T) {
	summary := &server.CombatSummary{
		Round:        5,
		TurnIndex:    2,
		Participants: []string{"Aldric", "Goblin", "Orc"},
	}

	assert.Equal(t, 5, summary.Round)
	assert.Equal(t, 2, summary.TurnIndex)
	assert.Len(t, summary.Participants, 3)
}
