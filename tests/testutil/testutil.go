// Package testutil 提供测试工具函数
package testutil

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

// TestContext 测试上下文
type TestContext struct {
	RedisClient    *redis.Client
	PostgresConn   *pgx.Conn
	T              *testing.T
	Cleanup        func()
	SessionIDs      []string
	MessageIDs     []string
}

// SetupTestEnvironment 设置测试环境
func SetupTestEnvironment(t *testing.T) *TestContext {
	ctx := &TestContext{
		T:           t,
		SessionIDs:  make([]string, 0),
		MessageIDs:  make([]string, 0),
	}

	// 获取环境变量
	redisHost := getEnv("REDIS_HOST", "localhost:6379")
	postgresURL := getEnv("DATABASE_URL", "")

	// 创建Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisHost,
		DB:   1, // 使用DB 1进行测试，避免污染数据
	})

	// 测试Redis连接
	testCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := redisClient.Ping(testCtx).Err()
	if err != nil {
		cleanupRedis(redisClient)
		t.Fatalf("无法连接到Redis: %v", err)
	}
	t.Logf("✓ Redis连接成功: %s", redisHost)

	ctx.RedisClient = redisClient

	// 创建PostgreSQL连接（如果配置了）
	if postgresURL != "" {
		conn, err := pgx.Connect(testCtx, postgresURL)
		if err != nil {
			cleanupRedis(redisClient)
			t.Fatalf("无法连接到PostgreSQL: %v", err)
		}
		ctx.PostgresConn = conn
		t.Logf("✓ PostgreSQL连接成功")
	}

	// 设置清理函数
	cleanupFunc := func() {
		t.Log("清理测试环境...")
		ctx.cleanup()
	}

	ctx.Cleanup = cleanupFunc

	return ctx
}

// SetupIntegrationTest 设置集成测试环境
func SetupIntegrationTest(t *testing.T) *TestContext {
	t.Helper()
	return SetupTestEnvironment(t)
}

// cleanup 清理测试环境
func (ctx *TestContext) cleanup() {
	// 删除测试创建的所有会话
	if len(ctx.SessionIDs) > 0 && ctx.RedisClient != nil {
		testCtx := context.Background()
		for _, sessionID := range ctx.SessionIDs {
			// 删除会话数据
			sessionKey := fmt.Sprintf("session:%s", sessionID)
			ctx.RedisClient.Del(testCtx, sessionKey)
			ctx.RedisClient.SRem(testCtx, "sessions:all", sessionID)

			// 删除消息
			messageKey := fmt.Sprintf("msg:%s", sessionID)
			ctx.RedisClient.Del(testCtx, messageKey)
		}
		ctx.T.Logf("✓ 清理了 %d 个测试会话", len(ctx.SessionIDs))
	}

	// 关闭Redis连接
	if ctx.RedisClient != nil {
		cleanupRedis(ctx.RedisClient)
	}

	// 关闭PostgreSQL连接
	if ctx.PostgresConn != nil {
		ctx.PostgresConn.Close(context.Background())
	}
}

// cleanupRedis 清理Redis客户端
func cleanupRedis(client *redis.Client) {
	if client != nil {
		client.Close()
	}
}

// CreateTestSession 创建测试会话
func (ctx *TestContext) CreateTestSession(data map[string]interface{}) string {
	ctx.T.Helper()

	sessionID := fmt.Sprintf("test-session-%d", time.Now().UnixNano())

	// 设置默认值
	if data["name"] == nil {
		data["name"] = "Test Session"
	}
	if data["creator_id"] == nil {
		data["creator_id"] = "test-user"
	}
	if data["status"] == nil {
		data["status"] = "active"
	}

	// 保存到Redis
	testCtx := context.Background()
	sessionKey := fmt.Sprintf("session:%s", sessionID)

	pipe := ctx.RedisClient.Pipeline()
	pipe.HSet(testCtx, sessionKey, data)
	pipe.SAdd(testCtx, "sessions:all", sessionID)
	_, err := pipe.Exec(testCtx)
	require.NoError(ctx.T, err, "创建测试会话失败")

	ctx.SessionIDs = append(ctx.SessionIDs, sessionID)
	ctx.T.Logf("✓ 创建测试会话: %s", sessionID)

	return sessionID
}

// CreateTestMessage 创建测试消息
func (ctx *TestContext) CreateTestMessage(sessionID string, content string) string {
	ctx.T.Helper()

	messageID := fmt.Sprintf("test-msg-%d", time.Now().UnixNano())

	testCtx := context.Background()
	messageKey := fmt.Sprintf("msg:%s", sessionID)

	// 添加到有序集合
	score := float64(time.Now().UnixNano() / 1e6)
	messageData := map[string]interface{}{
		"id":         messageID,
		"session_id": sessionID,
		"role":       "user",
		"content":    content,
		"created_at": time.Now().UTC().Format(time.RFC3339),
	}

	pipe := ctx.RedisClient.Pipeline()
	pipe.ZAdd(testCtx, messageKey, redis.Z{Score: score, Member: messageID})
	pipe.HSet(testCtx, fmt.Sprintf("message:%s", messageID), messageData)
	_, err := pipe.Exec(testCtx)
	require.NoError(ctx.T, err, "创建测试消息失败")

	ctx.MessageIDs = append(ctx.MessageIDs, messageID)
	ctx.T.Logf("✓ 创建测试消息: %s", messageID)

	return messageID
}

// FlushRedis 清空Redis测试数据库
func (ctx *TestContext) FlushRedis() {
	ctx.T.Helper()

	testCtx := context.Background()
	err := ctx.RedisClient.FlushDB(testCtx).Err()
	require.NoError(ctx.T, err, "清空Redis失败")
	ctx.T.Log("✓ 清空Redis测试数据库")
}

// WaitForCondition 等待条件满足
func (ctx *TestContext) WaitForCondition(condition func() bool, timeout time.Duration, msg string) {
	ctx.T.Helper()

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		<-ticker.C
	}

	ctx.T.Fatalf("等待条件超时: %s", msg)
}

// AssertRedisHasKey 断言Redis包含指定key
func (ctx *TestContext) AssertRedisHasKey(key string) {
	ctx.T.Helper()

	testCtx := context.Background()
	exists, err := ctx.RedisClient.Exists(testCtx, key).Result()
	require.NoError(ctx.T, err)
	require.True(ctx.T, exists > 0, "Redis应该包含key: %s", key)
}

// AssertRedisHashField 断言Redis Hash包含指定字段
func (ctx *TestContext) AssertRedisHashField(key, field string) {
	ctx.T.Helper()

	testCtx := context.Background()
	exists, err := ctx.RedisClient.HExists(testCtx, key, field).Result()
	require.NoError(ctx.T, err)
	require.True(ctx.T, exists, "Redis Hash %s 应该包含字段: %s", key, field)
}

// getEnv 获取环境变量，带默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
