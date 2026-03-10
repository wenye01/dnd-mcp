// Package integration_test 测试 Client-Server 协同
package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/dnd-mcp/client/internal/server"
	"github.com/dnd-mcp/client/internal/service"
)

// TestContextBuilder_GetContext 测试 ContextBuilder 从 Server 获取上下文
func TestContextBuilder_GetContext(t *testing.T) {
	// 使用 Mock ServerClient
	mockClient := server.NewMockClient()
	contextBuilder := service.NewContextBuilder(mockClient, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试简化模式
	messages, err := contextBuilder.BuildContext(ctx, "campaign-1", "我想攻击哥布林")
	if err != nil {
		t.Fatalf("BuildContext 失败: %v", err)
	}

	// 验证返回的消息
	if len(messages) < 2 {
		t.Errorf("期望至少 2 条消息（system + user），得到 %d 条", len(messages))
	}

	// 验证第一条是 system 消息
	if messages[0].Role != "system" {
		t.Errorf("第一条消息应该是 system，得到 %s", messages[0].Role)
	}

	// 验证最后一条是 user 消息
	lastMsg := messages[len(messages)-1]
	if lastMsg.Role != "user" {
		t.Errorf("最后一条消息应该是 user，得到 %s", lastMsg.Role)
	}

	if lastMsg.Content != "我想攻击哥布林" {
		t.Errorf("用户消息内容不匹配: %s", lastMsg.Content)
	}
}

// TestContextBuilder_RawContext 测试 ContextBuilder 使用完整模式
func TestContextBuilder_RawContext(t *testing.T) {
	mockClient := server.NewMockClient()
	config := &service.ContextBuilderConfig{
		UseRawContext: true,
		MessageLimit:  20,
		IncludeCombat: true,
	}
	contextBuilder := service.NewContextBuilder(mockClient, config)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	messages, err := contextBuilder.BuildContext(ctx, "campaign-1", "我要使用长剑攻击")
	if err != nil {
		t.Fatalf("BuildContext 失败: %v", err)
	}

	if len(messages) < 2 {
		t.Errorf("期望至少 2 条消息，得到 %d 条", len(messages))
	}

	// 验证 system 消息
	if messages[0].Role != "system" {
		t.Errorf("第一条消息应该是 system，得到 %s", messages[0].Role)
	}
}

// TestChatService_SaveMessage 测试 ChatService 通过 Server 保存消息
func TestChatService_SaveMessage(t *testing.T) {
	mockClient := server.NewMockClient()

	// 验证初始状态
	if len(mockClient.GetMessages()) != 0 {
		t.Errorf("初始消息数应该为 0")
	}

	// 保存消息
	ctx := context.Background()
	msg := &server.Message{
		ID:         "msg-test-1",
		CampaignID: "campaign-1",
		Role:       server.MessageRoleUser,
		Content:    "Hello, world!",
		PlayerID:   "player-1",
		CreatedAt:  time.Now(),
	}

	err := mockClient.SaveMessage(ctx, "campaign-1", msg)
	if err != nil {
		t.Fatalf("SaveMessage 失败: %v", err)
	}

	// 验证消息被保存
	messages := mockClient.GetMessages()
	if len(messages) != 1 {
		t.Errorf("期望 1 条消息，得到 %d 条", len(messages))
	}

	if messages[0].Content != "Hello, world!" {
		t.Errorf("消息内容不匹配: %s", messages[0].Content)
	}
}

// TestServerClient_CallTool 测试工具调用
func TestServerClient_CallTool(t *testing.T) {
	mockClient := server.NewMockClient()
	ctx := context.Background()

	// 测试 roll_dice 工具
	result, err := mockClient.CallTool(ctx, "campaign-1", "roll_dice", map[string]any{
		"formula": "1d20+5",
	})
	if err != nil {
		t.Fatalf("CallTool 失败: %v", err)
	}

	if success, ok := result["success"]; !ok || success != true {
		t.Errorf("期望 success=true，得到 %v", result)
	}
}

// TestContextBuilder_ErrorHandling 测试错误处理
func TestContextBuilder_ErrorHandling(t *testing.T) {
	mockClient := server.NewMockClient()
	contextBuilder := service.NewContextBuilder(mockClient, nil)

	ctx := context.Background()

	// 设置返回错误
	mockClient.SetReturnError(true)

	// 测试错误情况
	_, err := contextBuilder.BuildContext(ctx, "campaign-1", "test message")
	if err == nil {
		t.Error("期望返回错误，但没有")
	}

	// 恢复正常
	mockClient.SetReturnError(false)

	// 测试正常情况
	_, err = contextBuilder.BuildContext(ctx, "campaign-1", "test message")
	if err != nil {
		t.Errorf("不应该返回错误: %v", err)
	}
}

// TestCampaignID_Mapping 测试 Campaign ID 映射
func TestCampaignID_Mapping(t *testing.T) {
	mockClient := server.NewMockClient()
	contextBuilder := service.NewContextBuilder(mockClient, nil)

	testCases := []string{
		"campaign-1",
		"test-campaign",
		"550e8400-e29b-41d4-a716-446655440000",
	}

	for _, campaignID := range testCases {
		ctx := context.Background()
		messages, err := contextBuilder.BuildContext(ctx, campaignID, "test")
		if err != nil {
			t.Errorf("campaignID %s 失败: %v", campaignID, err)
			continue
		}

		// 验证上下文正确返回
		if len(messages) < 2 {
			t.Errorf("campaignID %s: 期望至少 2 条消息", campaignID)
		}
	}
}

// TestMultipleMessages 测试多消息场景
func TestMultipleMessages(t *testing.T) {
	mockClient := server.NewMockClient()
	ctx := context.Background()

	// 保存多条消息
	for i := 0; i < 5; i++ {
		msg := &server.Message{
			ID:         string(rune('A' + i)),
			CampaignID: "campaign-1",
			Role:       server.MessageRoleUser,
			Content:    "Message " + string(rune('A'+i)),
			CreatedAt:  time.Now().Add(time.Duration(i) * time.Minute),
		}
		mockClient.SaveMessage(ctx, "campaign-1", msg)
	}

	// 验证消息数量
	messages := mockClient.GetMessages()
	if len(messages) != 5 {
		t.Errorf("期望 5 条消息，得到 %d 条", len(messages))
	}

	// 清除消息
	mockClient.ClearMessages()
	if len(mockClient.GetMessages()) != 0 {
		t.Errorf("清除后应该没有消息")
	}
}
