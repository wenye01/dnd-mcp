// Package store_test 提供 Redis 存储集成测试
// 运行这些测试需要本地运行 Redis 服务
// 运行方法: go test ./tests/integration/store/... -tags=integration -v
package store_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dnd-mcp/client/internal/models"
	"github.com/dnd-mcp/client/internal/store"
	redisclient "github.com/dnd-mcp/client/internal/store/redis"
	"github.com/dnd-mcp/client/pkg/config"
)

var (
	redisClient redisclient.Client
	sessionStore store.SessionStore
	messageStore store.MessageStore
)

// TestMain 在所有测试前后执行
func TestMain(m *testing.M) {
	// 检查是否设置了集成测试标记
	// if os.Getenv("INTEGRATION_TEST") != "1" {
	// 	fmt.Println("跳过集成测试 (设置 INTEGRATION_TEST=1 来运行)")
	// 	os.Exit(0)
	// }

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 创建 Redis 客户端
	redisClient, err = redisclient.NewClient(&cfg.Redis)
	if err != nil {
		fmt.Printf("创建 Redis 客户端失败: %v\n", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	// 创建存储实例
	sessionStore = redisclient.NewSessionStore(redisClient)
	messageStore = redisclient.NewMessageStore(redisClient)

	// 运行测试
	code := m.Run()
	os.Exit(code)
}

func TestSessionCreateAndGet(t *testing.T) {
	ctx := context.Background()

	// 创建会话
	session := models.NewSession("测试会话", "user-123", "http://localhost:9000")
	session.ID = "test-session-1"
	session.WebSocketKey = "ws-test-key-1"

	err := sessionStore.Create(ctx, session)
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}

	// 获取会话
	retrieved, err := sessionStore.Get(ctx, session.ID)
	if err != nil {
		t.Fatalf("获取会话失败: %v", err)
	}

	// 验证数据
	if retrieved.ID != session.ID {
		t.Errorf("ID 不匹配: got %s, want %s", retrieved.ID, session.ID)
	}
	if retrieved.Name != session.Name {
		t.Errorf("Name 不匹配: got %s, want %s", retrieved.Name, session.Name)
	}
	if retrieved.CreatorID != session.CreatorID {
		t.Errorf("CreatorID 不匹配: got %s, want %s", retrieved.CreatorID, session.CreatorID)
	}

	// 清理
	_ = sessionStore.Delete(ctx, session.ID)
}

func TestSessionList(t *testing.T) {
	ctx := context.Background()

	// 创建多个会话
	sessions := []*models.Session{
		models.NewSession("会话1", "user-1", "http://localhost:9000"),
		models.NewSession("会话2", "user-2", "http://localhost:9001"),
		models.NewSession("会话3", "user-3", "http://localhost:9002"),
	}

	for i, session := range sessions {
		session.ID = fmt.Sprintf("test-list-session-%d", i)
		session.WebSocketKey = fmt.Sprintf("ws-key-%d", i)
		if err := sessionStore.Create(ctx, session); err != nil {
			t.Fatalf("创建会话 %d 失败: %v", i, err)
		}
		defer sessionStore.Delete(ctx, session.ID)
	}

	// 列出所有会话
	list, err := sessionStore.List(ctx)
	if err != nil {
		t.Fatalf("列出会话失败: %v", err)
	}

	// 验证至少包含我们创建的会话
	if len(list) < len(sessions) {
		t.Errorf("会话数量不足: got %d, want >= %d", len(list), len(sessions))
	}
}

func TestSessionUpdate(t *testing.T) {
	ctx := context.Background()

	// 创建会话
	session := models.NewSession("原始名称", "user-123", "http://localhost:9000")
	session.ID = "test-update-session"
	session.WebSocketKey = "ws-update-key"

	err := sessionStore.Create(ctx, session)
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}
	defer sessionStore.Delete(ctx, session.ID)

	// 更新会话
	session.Name = "更新后的名称"
	session.MaxPlayers = 6
	session.UpdateSettings(map[string]interface{}{
		"ruleset": "dnd5e",
	})

	err = sessionStore.Update(ctx, session)
	if err != nil {
		t.Fatalf("更新会话失败: %v", err)
	}

	// 获取并验证
	retrieved, err := sessionStore.Get(ctx, session.ID)
	if err != nil {
		t.Fatalf("获取更新后的会话失败: %v", err)
	}

	if retrieved.Name != "更新后的名称" {
		t.Errorf("Name 未更新: got %s, want 更新后的名称", retrieved.Name)
	}
	if retrieved.MaxPlayers != 6 {
		t.Errorf("MaxPlayers 未更新: got %d, want 6", retrieved.MaxPlayers)
	}
	if retrieved.Settings["ruleset"] != "dnd5e" {
		t.Errorf("Settings 未更新: got %v, want dnd5e", retrieved.Settings["ruleset"])
	}
}

func TestMessageCreateAndGet(t *testing.T) {
	ctx := context.Background()

	// 创建会话
	sessionID := "test-message-session"
	session := models.NewSession("测试会话", "user-123", "http://localhost:9000")
	session.ID = sessionID
	session.WebSocketKey = "ws-message-key"

	err := sessionStore.Create(ctx, session)
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}
	defer sessionStore.Delete(ctx, sessionID)

	// 创建消息
	message := models.NewUserMessage(sessionID, "你好,世界!", "player-123")
	message.ID = "test-message-1"

	err = messageStore.Create(ctx, message)
	if err != nil {
		t.Fatalf("创建消息失败: %v", err)
	}

	// 获取消息
	retrieved, err := messageStore.Get(ctx, sessionID, message.ID)
	if err != nil {
		t.Fatalf("获取消息失败: %v", err)
	}

	// 验证数据
	if retrieved.ID != message.ID {
		t.Errorf("ID 不匹配: got %s, want %s", retrieved.ID, message.ID)
	}
	if retrieved.Content != message.Content {
		t.Errorf("Content 不匹配: got %s, want %s", retrieved.Content, message.Content)
	}
	if retrieved.Role != "user" {
		t.Errorf("Role 错误: got %s, want user", retrieved.Role)
	}
}

func TestMessageList(t *testing.T) {
	ctx := context.Background()

	// 创建会话
	sessionID := "test-list-messages-session"
	session := models.NewSession("测试会话", "user-123", "http://localhost:9000")
	session.ID = sessionID
	session.WebSocketKey = "ws-list-messages-key"

	err := sessionStore.Create(ctx, session)
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}
	defer sessionStore.Delete(ctx, sessionID)

	// 创建多条消息
	messages := []string{
		"第一条消息",
		"第二条消息",
		"第三条消息",
		"第四条消息",
		"第五条消息",
	}

	for i, content := range messages {
		message := models.NewUserMessage(sessionID, content, "player-123")
		message.ID = fmt.Sprintf("test-list-msg-%d", i)
		// 添加延迟以确保时间戳不同
		time.Sleep(10 * time.Millisecond)
		if err := messageStore.Create(ctx, message); err != nil {
			t.Fatalf("创建消息 %d 失败: %v", i, err)
		}
	}

	// 获取消息列表
	list, err := messageStore.List(ctx, sessionID, 10)
	if err != nil {
		t.Fatalf("获取消息列表失败: %v", err)
	}

	// 验证数量
	if len(list) != len(messages) {
		t.Errorf("消息数量不匹配: got %d, want %d", len(list), len(messages))
	}

	// 验证顺序 (应该从旧到新)
	if list[0].Content != "第一条消息" {
		t.Errorf("第一条消息错误: got %s, want 第一条消息", list[0].Content)
	}
	if list[len(list)-1].Content != "第五条消息" {
		t.Errorf("最后一条消息错误: got %s, want 第五条消息", list[len(list)-1].Content)
	}
}

func TestMessageListWithLimit(t *testing.T) {
	ctx := context.Background()

	// 创建会话
	sessionID := "test-limit-messages-session"
	session := models.NewSession("测试会话", "user-123", "http://localhost:9000")
	session.ID = sessionID
	session.WebSocketKey = "ws-limit-key"

	err := sessionStore.Create(ctx, session)
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}
	defer sessionStore.Delete(ctx, sessionID)

	// 创建5条消息
	for i := 0; i < 5; i++ {
		message := models.NewUserMessage(sessionID, fmt.Sprintf("消息%d", i), "player-123")
		message.ID = fmt.Sprintf("test-limit-msg-%d", i)
		time.Sleep(10 * time.Millisecond)
		if err := messageStore.Create(ctx, message); err != nil {
			t.Fatalf("创建消息失败: %v", err)
		}
	}

	// 测试限制数量
	tests := []struct {
		name      string
		limit     int
		wantCount int
	}{
		{"获取3条", 3, 3},
		{"获取全部", -1, 5},
		{"获取0条(应该返回默认50条)", 0, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list, err := messageStore.List(ctx, sessionID, tt.limit)
			if err != nil {
				t.Fatalf("获取消息列表失败: %v", err)
			}
			if len(list) != tt.wantCount {
				t.Errorf("消息数量错误: got %d, want %d", len(list), tt.wantCount)
			}
		})
	}
}
