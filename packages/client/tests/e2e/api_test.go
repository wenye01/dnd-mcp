// Package e2e 提供端到端测试，模拟真实HTTP调用
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL = "http://localhost:8080"
)

// TestE2E_FullUserFlow 完整的用户流程测试
func TestE2E_FullUserFlow(t *testing.T) {
	// 等待服务启动
	waitForServer(t)

	t.Run("1_健康检查", func(t *testing.T) {
		resp := request(t, "GET", "/api/system/health", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var health HealthResponse
		err := json.Unmarshal(respBody(resp), &health)
		require.NoError(t, err)
		assert.Equal(t, "healthy", health.Status)
		assert.NotEmpty(t, health.Components)
	})

	t.Run("2_创建会话", func(t *testing.T) {
		createReq := map[string]interface{}{
			"name":           "E2E测试会话",
			"creator_id":     "user-e2e-test",
			"mcp_server_url":  "http://localhost:9000",
			"max_players":    4,
		}

		resp := requestJSON(t, "POST", "/api/sessions", createReq)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var session SessionResponse
		err := json.Unmarshal(respBody(resp), &session)
		require.NoError(t, err)

		// 验证响应字段
		assert.NotEmpty(t, session.ID)
		assert.Equal(t, "E2E测试会话", session.Name)
		assert.Equal(t, "user-e2e-test", session.CreatorID)
		assert.Equal(t, "active", session.Status)
		assert.NotEmpty(t, session.CreatedAt)
		assert.NotEmpty(t, session.WebSocketKey)

		// 保存sessionID供后续测试使用
		setSessionID(t, session.ID)
	})

	t.Run("3_获取会话详情", func(t *testing.T) {
		sessionID := getSessionID(t)

		resp := request(t, "GET", fmt.Sprintf("/api/sessions/%s", sessionID), nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var session SessionResponse
		err := json.Unmarshal(respBody(resp), &session)
		require.NoError(t, err)

		assert.Equal(t, sessionID, session.ID)
		assert.Equal(t, "E2E测试会话", session.Name)
	})

	t.Run("4_列出所有会话", func(t *testing.T) {
		resp := request(t, "GET", "/api/sessions", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var sessions []SessionResponse
		err := json.Unmarshal(respBody(resp), &sessions)
		require.NoError(t, err)

		// 验证至少包含我们创建的会话
		assert.GreaterOrEqual(t, len(sessions), 1)

		// 查找我们的会话
		sessionID := getSessionID(t)
		found := false
		for _, s := range sessions {
			if s.ID == sessionID {
				found = true
				break
			}
		}
		assert.True(t, found, "创建的会话应该在列表中")
	})

	t.Run("5_发送用户消息", func(t *testing.T) {
		sessionID := getSessionID(t)

		msgReq := map[string]interface{}{
			"content":   "你好，这是一条测试消息",
			"player_id": "player-e2e-test",
		}

		resp := requestJSON(t, "POST", fmt.Sprintf("/api/sessions/%s/chat", sessionID), msgReq)

		// 可能返回200或500（取决于是否配置LLM）
		// 至少不应该返回404
		assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)

		if resp.StatusCode == http.StatusOK {
			var message MessageResponse
			err := json.Unmarshal(respBody(resp), &message)
			require.NoError(t, err)

			assert.NotEmpty(t, message.ID)
			assert.Equal(t, sessionID, message.SessionID)
			assert.NotEmpty(t, message.Role)
			assert.NotEmpty(t, message.Content)
		}
	})

	t.Run("6_获取消息历史", func(t *testing.T) {
		sessionID := getSessionID(t)

		resp := request(t, "GET", fmt.Sprintf("/api/sessions/%s/messages?limit=10", sessionID), nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var messages []MessageResponse
		err := json.Unmarshal(respBody(resp), &messages)
		require.NoError(t, err)

		// 至少应该有我们刚才发送的消息
		assert.GreaterOrEqual(t, len(messages), 1)
	})

	t.Run("7_更新会话", func(t *testing.T) {
		sessionID := getSessionID(t)

		updateReq := map[string]interface{}{
			"name": "E2E测试会话-已更新",
		}

		resp := requestJSON(t, "PATCH", fmt.Sprintf("/api/sessions/%s", sessionID), updateReq)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var session SessionResponse
		err := json.Unmarshal(respBody(resp), &session)
		require.NoError(t, err)

		assert.Equal(t, "E2E测试会话-已更新", session.Name)
	})

	t.Run("8_系统统计", func(t *testing.T) {
		resp := request(t, "GET", "/api/system/stats", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var stats StatsResponse
		err := json.Unmarshal(respBody(resp), &stats)
		require.NoError(t, err)

		assert.Greater(t, stats.Uptime, int64(0))
		assert.NotEmpty(t, stats.Version)
		assert.GreaterOrEqual(t, stats.RequestCount, int64(0))
	})

	t.Run("9_删除会话", func(t *testing.T) {
		sessionID := getSessionID(t)

		resp := request(t, "DELETE", fmt.Sprintf("/api/sessions/%s", sessionID), nil)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("10_验证会话已删除", func(t *testing.T) {
		sessionID := getSessionID(t)

		resp := request(t, "GET", fmt.Sprintf("/api/sessions/%s", sessionID), nil)
		// 应该返回404或特定的错误
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

// TestE2E_ErrorCases 测试各种错误情况
func TestE2E_ErrorCases(t *testing.T) {
	waitForServer(t)

	t.Run("获取不存在的会话", func(t *testing.T) {
		resp := request(t, "GET", "/api/sessions/non-existent-id", nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("创建会话缺少必需字段", func(t *testing.T) {
		invalidReq := map[string]interface{}{
			"name": "", // 空名称
		}

		resp := requestJSON(t, "POST", "/api/sessions", invalidReq)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("发送消息到不存在的会话", func(t *testing.T) {
		msgReq := map[string]interface{}{
			"content":   "测试消息",
			"player_id": "player-123",
		}

		resp := requestJSON(t, "POST", "/api/sessions/non-existent/chat", msgReq)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("无效的JSON请求体", func(t *testing.T) {
		req, err := http.NewRequest("POST", baseURL+"/api/sessions", bytes.NewBufferString("invalid json"))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()
	})
}

// TestE2E_Concurrency 并发测试
func TestE2E_Concurrency(t *testing.T) {
	waitForServer(t)

	t.Run("并发创建会话", func(t *testing.T) {
		const concurrency = 10
		results := make(chan int, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(index int) {
				createReq := map[string]interface{}{
					"name":           fmt.Sprintf("并发测试会话-%d", index),
					"creator_id":     fmt.Sprintf("user-%d", index),
					"mcp_server_url":  "http://localhost:9000",
					"max_players":    4,
				}

				resp := requestJSON(t, "POST", "/api/sessions", createReq)
				results <- resp.StatusCode
			}(i)
		}

		// 收集结果
		successCount := 0
		for i := 0; i < concurrency; i++ {
			if <-results == http.StatusCreated {
				successCount++
			}
		}

		// 所有并发请求都应该成功
		assert.Equal(t, concurrency, successCount)
	})
}

// waitForServer 等待服务器启动
func waitForServer(t *testing.T) {
	const maxRetries = 30
	const retryInterval = 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(baseURL + "/api/system/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(retryInterval)
	}

	t.Fatal("服务器未能在30秒内启动")
}

// request 发送HTTP请求
func request(t *testing.T, method, path string, body []byte) *http.Response {
	req, err := http.NewRequest(method, baseURL+path, bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)

	return resp
}

// requestJSON 发送JSON请求
func requestJSON(t *testing.T, method, path string, data interface{}) *http.Response {
	body, err := json.Marshal(data)
	require.NoError(t, err)

	return request(t, method, path, body)
}

// respBody 读取响应体
func respBody(resp *http.Response) []byte {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	return body
}

// 测试辅助类型
type HealthResponse struct {
	Status     string                   `json:"status"`
	Timestamp  string                   `json:"timestamp"`
	Components map[string]ComponentHealth `json:"components"`
}

type ComponentHealth struct {
	Status    string  `json:"status"`
	Message   string  `json:"message,omitempty"`
	LatencyMs float64 `json:"latency_ms,omitempty"`
}

type SessionResponse struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	CreatorID    string                 `json:"creator_id"`
	MCPServerURL string                 `json:"mcp_server_url"`
	WebSocketKey string                 `json:"websocket_key"`
	MaxPlayers   int                    `json:"max_players"`
	Settings     map[string]interface{} `json:"settings"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    string                 `json:"updated_at"`
	Status       string                 `json:"status"`
}

type MessageResponse struct {
	ID        string     `json:"id"`
	SessionID string     `json:"session_id"`
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	PlayerID  string     `json:"player_id,omitempty"`
	CreatedAt string     `json:"created_at"`
}

type StatsResponse struct {
	Uptime      int64                  `json:"uptime_seconds"`
	StartTime   string                 `json:"start_time"`
	Version     string                 `json:"version"`
	RequestCount int64                 `json:"request_count"`
	ErrorCount   int64                 `json:"error_count"`
	Components  map[string]interface{} `json:"components"`
}

// 测试上下文管理
var testSessionID string

func setSessionID(t *testing.T, id string) {
	testSessionID = id
	t.Logf("设置会话ID: %s", id)
}

func getSessionID(t *testing.T) string {
	if testSessionID == "" {
		t.Fatal("会话ID未设置，请先运行创建会话测试")
	}
	return testSessionID
}

// TestMain 测试主函数
func TestMain(m *testing.M) {
	// 可以在这里设置测试标志
	os.Exit(m.Run())
}
