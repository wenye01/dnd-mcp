// Package integration 提供集成测试，使用真实Redis
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dnd-mcp/client/internal/api/handler"
	"github.com/dnd-mcp/client/internal/monitor"
	"github.com/dnd-mcp/client/internal/service"
	"github.com/dnd-mcp/client/internal/store"
	"github.com/dnd-mcp/client/internal/store/redis"
	"github.com/dnd-mcp/client/pkg/config"
	"github.com/dnd-mcp/client/tests/testutil"
	"github.com/gin-gonic/gin"
	goRedis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPI_SessionCRUD 测试会话的完整CRUD操作
func TestAPI_SessionCRUD(t *testing.T) {
	// 检查是否设置了集成测试标志
	if testing.Short() {
		t.Skip("跳过集成测试（使用 -short 标志）")
	}

	// 设置测试环境
	testCtx := testutil.SetupIntegrationTest(t)
	defer testCtx.Cleanup()

	// 创建HTTP测试服务器
	testServer := setupTestServer(testCtx.RedisClient)
	defer testServer.Close()

	// 准备测试数据
	createSessionData := map[string]interface{}{
		"name":           "集成测试会话",
		"creator_id":     "integration-test-user",
		"mcp_server_url": "http://localhost:9000",
		"max_players":    6,
	}

	t.Run("1_创建会话", func(t *testing.T) {
		body := marshallJSON(t, createSessionData)

		req, err := http.NewRequest("POST", testServer.URL+"/api/sessions", body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var session map[string]interface{}
		err = unmarshalJSON(resp, &session)
		require.NoError(t, err)

		// 验证返回的数据
		assert.NotEmpty(t, session["id"])
		assert.Equal(t, "集成测试会话", session["name"])
		assert.Equal(t, "integration-test-user", session["creator_id"])
		assert.Equal(t, "active", session["status"])

		// 保存sessionID
		testCtx.SessionIDs = append(testCtx.SessionIDs, session["id"].(string))
		t.Logf("创建会话成功: %s", session["id"])
	})

	if len(testCtx.SessionIDs) == 0 {
		t.Fatal("未成功创建会话")
	}
	sessionID := testCtx.SessionIDs[0]

	t.Run("2_获取会话详情", func(t *testing.T) {
		req, err := http.NewRequest("GET", testServer.URL+"/api/sessions/"+sessionID, nil)
		require.NoError(t, err)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var session map[string]interface{}
		err = unmarshalJSON(resp, &session)
		require.NoError(t, err)

		assert.Equal(t, sessionID, session["id"])
		assert.Equal(t, "集成测试会话", session["name"])
	})

	t.Run("3_列出所有会话", func(t *testing.T) {
		req, err := http.NewRequest("GET", testServer.URL+"/api/sessions", nil)
		require.NoError(t, err)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var sessions []map[string]interface{}
		err = unmarshalJSON(resp, &sessions)
		require.NoError(t, err)

		// 验证至少包含一个会话
		assert.GreaterOrEqual(t, len(sessions), 1)

		// 查找我们创建的会话
		found := false
		for _, s := range sessions {
			if s["id"] == sessionID {
				found = true
				break
			}
		}
		assert.True(t, found, "应该能在列表中找到创建的会话")
	})

	t.Run("4_更新会话", func(t *testing.T) {
		updateData := map[string]interface{}{
			"name": "集成测试会话-已更新",
		}

		body := marshallJSON(t, updateData)
		req, err := http.NewRequest("PATCH", testServer.URL+"/api/sessions/"+sessionID, body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var session map[string]interface{}
		err = unmarshalJSON(resp, &session)
		require.NoError(t, err)

		assert.Equal(t, "集成测试会话-已更新", session["name"])
	})

	t.Run("5_删除会话", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", testServer.URL+"/api/sessions/"+sessionID, nil)
		require.NoError(t, err)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// 验证会话已删除
		req, err = http.NewRequest("GET", testServer.URL+"/api/sessions/"+sessionID, nil)
		require.NoError(t, err)

		resp, err = client.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

// TestAPI_HealthCheck 测试健康检查API
func TestAPI_HealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试（使用 -short 标志）")
	}

	testCtx := testutil.SetupIntegrationTest(t)
	defer testCtx.Cleanup()

	testServer := setupTestServer(testCtx.RedisClient)
	defer testServer.Close()

	t.Run("健康检查", func(t *testing.T) {
		req, err := http.NewRequest("GET", testServer.URL+"/api/system/health", nil)
		require.NoError(t, err)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var health map[string]interface{}
		err = unmarshalJSON(resp, &health)
		require.NoError(t, err)

		assert.Equal(t, "healthy", health["status"])
		assert.NotEmpty(t, health["components"])
	})
}

// TestAPI_SystemStats 测试系统统计API
func TestAPI_SystemStats(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试（使用 -short 标志）")
	}

	testCtx := testutil.SetupIntegrationTest(t)
	defer testCtx.Cleanup()

	testServer := setupTestServer(testCtx.RedisClient)
	defer testServer.Close()

	t.Run("系统统计", func(t *testing.T) {
		// 等待一点时间让统计数据初始化
		time.Sleep(100 * time.Millisecond)

		req, err := http.NewRequest("GET", testServer.URL+"/api/system/stats", nil)
		require.NoError(t, err)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var stats map[string]interface{}
		err = unmarshalJSON(resp, &stats)
		require.NoError(t, err)

		// uptime_seconds 可能是 float64 (JSON number)
		uptime, ok := stats["uptime_seconds"].(float64)
		require.True(t, ok, "uptime_seconds should be a number")
		assert.GreaterOrEqual(t, uptime, float64(0)) // 允许为0，因为刚启动
		assert.NotEmpty(t, stats["version"])
	})
}

// setupTestServer 设置测试服务器
func setupTestServer(redisClient *goRedis.Client) *httptest.Server {
	gin.SetMode(gin.TestMode)

	// 创建配置
	cfg := &config.RedisConfig{
		Host: "localhost:6379",
	}

	// 创建Redis客户端包装器
	redisClientWrapper, err := redis.NewClient(cfg)
	if err != nil {
		panic(fmt.Sprintf("创建Redis客户端失败: %v", err))
	}

	// 创建存储
	sessionStore := redis.NewSessionStore(redisClientWrapper)
	messageStore := redis.NewMessageStore(redisClientWrapper)

	// 创建服务
	sessionService := service.NewSessionService(sessionStore)

	// 创建监控器
	healthMonitor := monitor.NewHealthMonitor()
	healthMonitor.Register(monitor.NewRedisHealthChecker(redisClientWrapper))

	statsMonitor := monitor.NewStatsMonitor("v0.1.0-test")
	// 不注册收集器，避免接口不匹配问题

	// 创建处理器
	systemHandler := handler.NewSystemHandler(nil, healthMonitor, statsMonitor)

	// 设置路由
	router := setupTestRoutes(sessionService, messageStore, systemHandler)

	// 创建并返回测试服务器
	return httptest.NewServer(router)
}

// setupTestRoutes 设置测试路由
func setupTestRoutes(sessionService *service.SessionService, messageStore store.MessageStore, systemHandler *handler.SystemHandler) *gin.Engine {
	router := gin.New()

	// 会话路由
	sessions := router.Group("/api/sessions")
	{
		sessions.POST("", handler.CreateSession(sessionService))
		sessions.GET("", handler.ListSessions(sessionService))
		sessions.GET("/:id", handler.GetSession(sessionService))
		sessions.PATCH("/:id", handler.UpdateSession(sessionService))
		sessions.DELETE("/:id", handler.DeleteSession(sessionService))
	}

	// 系统路由
	system := router.Group("/api/system")
	{
		system.GET("/health", systemHandler.Health)
		system.GET("/stats", systemHandler.Stats)
	}

	return router
}

// createSessionViaAPI 通过API创建会话
func createSessionViaAPI(t *testing.T, testServer *httptest.Server, name string) string {
	t.Helper()

	createSessionData := map[string]interface{}{
		"name":           name,
		"creator_id":     "test-user",
		"mcp_server_url": "http://localhost:9000",
		"max_players":    4,
	}

	body := marshallJSON(t, createSessionData)

	req, err := http.NewRequest("POST", testServer.URL+"/api/sessions", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var session map[string]interface{}
	err = unmarshalJSON(resp, &session)
	require.NoError(t, err)

	return session["id"].(string)
}

// marshallJSON 辅助函数：序列化JSON
func marshallJSON(t *testing.T, data interface{}) *bytes.Buffer {
	t.Helper()

	jsonBytes, err := json.Marshal(data)
	require.NoError(t, err)

	return bytes.NewBuffer(jsonBytes)
}

// unmarshalJSON 辅助函数：反序列化JSON
func unmarshalJSON(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(v)
}
