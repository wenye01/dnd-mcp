package store

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dnd-mcp/client/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPostgresStore_CreateSession 测试创建会话
func TestPostgresStore_CreateSession(t *testing.T) {
	// 使用环境变量 DATABASE_URL 或使用默认值
	db, err := setupTestDB()
	require.NoError(t, err, "Failed to setup test database")
	defer db.Close()

	store, err := NewPostgresStore(getTestDatabaseURL())
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	session := &models.Session{
		ID:           mustParseUUID("550e8400-e29b-41d4-a716-446655440001"),
		Version:      1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		GameTime:     "Morning",
		Location:     "地下城入口",
		CampaignName: "被遗忘的国度",
		State:        map[string]interface{}{},
	}

	err = store.CreateSession(ctx, session)
	assert.NoError(t, err, "Failed to create session")

	// 验证会话已创建
	fetched, err := store.GetSession(ctx, session.ID)
	assert.NoError(t, err)
	assert.Equal(t, session.ID, fetched.ID)
	assert.Equal(t, session.CampaignName, fetched.CampaignName)
	assert.Equal(t, session.Location, fetched.Location)
}

// TestPostgresStore_GetSession_NotFound 测试获取不存在的会话
func TestPostgresStore_GetSession_NotFound(t *testing.T) {
	store, err := setupTestStore()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	sessionID := mustParseUUID("550e8400-e29b-41d4-a716-446655449999")

	session, err := store.GetSession(ctx, sessionID)
	assert.Error(t, err)
	assert.Nil(t, session)
}

// TestPostgresStore_CreateMessage 测试创建消息
func TestPostgresStore_CreateMessage(t *testing.T) {
	store, err := setupTestStore()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// 首先创建一个会话
	sessionID := mustParseUUID("550e8400-e29b-41d4-a716-446655440002")
	session := &models.Session{
		ID:           sessionID,
		Version:      1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		GameTime:     "Morning",
		Location:     "测试地点",
		CampaignName: "测试战役",
		State:        map[string]interface{}{},
	}
	err = store.CreateSession(ctx, session)
	require.NoError(t, err)

	// 创建消息
	message := &models.Message{
		ID:        mustParseUUID("550e8400-e29b-41d4-a716-446655440003"),
		SessionID: sessionID,
		Role:      "user",
		Content:   "你好，地下城主",
		CreatedAt: time.Now(),
	}

	err = store.CreateMessage(ctx, message)
	assert.NoError(t, err, "Failed to create message")

	// 验证消息已创建
	messages, err := store.GetMessages(ctx, sessionID, 100, 0)
	assert.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, message.Content, messages[0].Content)
	assert.Equal(t, message.Role, messages[0].Role)
}

// TestPostgresStore_ListMessages_Empty 测试列出空消息列表
func TestPostgresStore_ListMessages_Empty(t *testing.T) {
	store, err := setupTestStore()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	sessionID := mustParseUUID("550e8400-e29b-41d4-a716-446655440004")

	messages, err := store.GetMessages(ctx, sessionID, 100, 0)
	assert.NoError(t, err)
	assert.Empty(t, messages)
}

// TestPostgresStore_ListMessages_Multiple 测试列出多条消息
func TestPostgresStore_ListMessages_Multiple(t *testing.T) {
	store, err := setupTestStore()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// 创建会话
	sessionID := mustParseUUID("550e8400-e29b-41d4-a716-446655440005")
	session := &models.Session{
		ID:           sessionID,
		Version:      1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		GameTime:     "Morning",
		Location:     "测试地点",
		CampaignName: "测试战役",
		State:        map[string]interface{}{},
	}
	err = store.CreateSession(ctx, session)
	require.NoError(t, err)

	// 创建多条消息
	messages := []*models.Message{
		{
			ID:        mustParseUUID("550e8400-e29b-41d4-a716-446655440006"),
			SessionID: sessionID,
			Role:      "user",
			Content:   "消息1",
			CreatedAt: time.Now(),
		},
		{
			ID:        mustParseUUID("550e8400-e29b-41d4-a716-446655440007"),
			SessionID: sessionID,
			Role:      "assistant",
			Content:   "消息2",
			CreatedAt: time.Now().Add(1 * time.Second),
		},
		{
			ID:        mustParseUUID("550e8400-e29b-41d4-a716-446655440008"),
			SessionID: sessionID,
			Role:      "user",
			Content:   "消息3",
			CreatedAt: time.Now().Add(2 * time.Second),
		},
	}

	for _, msg := range messages {
		err = store.CreateMessage(ctx, msg)
		require.NoError(t, err)
	}

	// 列出消息
	fetched, err := store.GetMessages(ctx, sessionID, 100, 0)
	assert.NoError(t, err)
	assert.Len(t, fetched, 3)
	assert.Equal(t, "消息1", fetched[0].Content)
	assert.Equal(t, "消息2", fetched[1].Content)
	assert.Equal(t, "消息3", fetched[2].Content)
}

// TestPostgresStore_DeleteSession_SoftDelete 测试软删除会话
func TestPostgresStore_DeleteSession_SoftDelete(t *testing.T) {
	store, err := setupTestStore()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// 创建会话
	sessionID := mustParseUUID("550e8400-e29b-41d4-a716-446655440009")
	session := &models.Session{
		ID:           sessionID,
		Version:      1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		GameTime:     "Morning",
		Location:     "测试地点",
		CampaignName: "测试战役",
		State:        map[string]interface{}{},
	}
	err = store.CreateSession(ctx, session)
	require.NoError(t, err)

	// 删除会话
	err = store.DeleteSession(ctx, sessionID)
	assert.NoError(t, err, "Failed to delete session")

	// 验证会话已被软删除
	fetched, err := store.GetSession(ctx, sessionID)
	assert.Error(t, err)
	assert.Nil(t, fetched)
}

// setupTestDB 设置测试数据库（使用真实数据库）
func setupTestDB() (*PostgresStore, error) {
	databaseURL := getTestDatabaseURL()
	return NewPostgresStore(databaseURL)
}

// setupTestStore 设置测试store并清理数据
func setupTestStore() (*PostgresStore, error) {
	store, err := setupTestDB()
	if err != nil {
		return nil, err
	}

	// 清理测试数据
	cleanupTestData(store)

	return store, nil
}

// cleanupTestData 清理测试数据
func cleanupTestData(store *PostgresStore) {
	ctx := context.Background()
	store.db.ExecContext(ctx, "DELETE FROM messages WHERE session_id IN (SELECT id FROM sessions WHERE campaign_name = '测试战役')")
	store.db.ExecContext(ctx, "DELETE FROM sessions WHERE campaign_name = '测试战役'")
}

// getTestDatabaseURL 获取测试数据库URL
func getTestDatabaseURL() string {
	// 从环境变量读取数据库配置
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	// 从环境变量读取密码，如果没有则使用默认值
	password := os.Getenv("TEST_DB_PASSWORD")
	if password == "" {
		password = "070831" // 默认密码
	}
	return fmt.Sprintf("postgres://postgres:%s@localhost:5432/dnd_mcp_test?sslmode=disable", password)
}

// mustParseUUID 解析UUID，如果失败则panic
func mustParseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return id
}
