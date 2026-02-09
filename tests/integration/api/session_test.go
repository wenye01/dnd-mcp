// Package api_test 提供 API 层集成测试
package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dnd-mcp/client/internal/api/handler"
	"github.com/dnd-mcp/client/internal/models"
	"github.com/dnd-mcp/client/internal/service"
	"github.com/dnd-mcp/client/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// MockSessionService Mock Service 接口
type MockSessionService struct {
	mockCreateSession func(req *service.CreateSessionRequest) (*models.Session, error)
	mockGetSession    func(sessionID string) (*models.Session, error)
	mockListSessions  func(status string) ([]*models.Session, error)
	mockUpdateSession func(sessionID string, req *service.UpdateSessionRequest) (*models.Session, error)
	mockDeleteSession func(sessionID string) error
}

func (m *MockSessionService) CreateSession(ctx context.Context, req *service.CreateSessionRequest) (*models.Session, error) {
	return m.mockCreateSession(req)
}

func (m *MockSessionService) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	return m.mockGetSession(sessionID)
}

func (m *MockSessionService) ListSessions(ctx context.Context, status string) ([]*models.Session, error) {
	return m.mockListSessions(status)
}

func (m *MockSessionService) UpdateSession(ctx context.Context, sessionID string, req *service.UpdateSessionRequest) (*models.Session, error) {
	return m.mockUpdateSession(sessionID, req)
}

func (m *MockSessionService) DeleteSession(ctx context.Context, sessionID string) error {
	return m.mockDeleteSession(sessionID)
}

// setupTestRouter 设置测试路由
func setupTestRouter(sessionService *MockSessionService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 注册路由
	router.POST("/api/sessions", handler.CreateSession(sessionService))
	router.GET("/api/sessions", handler.ListSessions(sessionService))
	router.GET("/api/sessions/:id", handler.GetSession(sessionService))
	router.PATCH("/api/sessions/:id", handler.UpdateSession(sessionService))
	router.DELETE("/api/sessions/:id", handler.DeleteSession(sessionService))

	return router
}

// TestCreateSession_Handler_Success 测试创建会话成功
func TestCreateSession_Handler_Success(t *testing.T) {
	mockService := &MockSessionService{
		mockCreateSession: func(req *service.CreateSessionRequest) (*models.Session, error) {
			return &models.Session{
				ID:           "test-session-123",
				Name:         req.Name,
				CreatorID:    req.CreatorID,
				MCPServerURL: req.MCPServerURL,
				MaxPlayers:   req.MaxPlayers,
				WebSocketKey: "ws-test-key",
				Status:       "active",
			}, nil
		},
	}

	router := setupTestRouter(mockService)

	reqBody := `{
		"name": "测试会话",
		"creator_id": "user-123",
		"mcp_server_url": "http://localhost:9000",
		"max_players": 5
	}`

	req := httptest.NewRequest("POST", "/api/sessions", bytes.NewReader([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-session-123", response["id"])
	assert.Equal(t, "测试会话", response["name"])
	assert.Equal(t, "user-123", response["creator_id"])
}

// TestCreateSession_Handler_InvalidRequest 测试无效请求
func TestCreateSession_Handler_InvalidRequest(t *testing.T) {
	mockService := &MockSessionService{}
	router := setupTestRouter(mockService)

	tests := []struct {
		name       string
		reqBody    string
		wantStatus int
	}{
		{
			name:       "缺少必填字段",
			reqBody:    `{"name": "测试"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "空name",
			reqBody:    `{"name": "", "creator_id": "user", "mcp_server_url": "http://test.com"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "无效的JSON",
			reqBody:    `{invalid json}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/sessions", bytes.NewReader([]byte(tt.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response, "error")
		})
	}
}

// TestGetSession_Handler_Success 测试获取会话成功
func TestGetSession_Handler_Success(t *testing.T) {
	mockService := &MockSessionService{
		mockGetSession: func(sessionID string) (*models.Session, error) {
			return &models.Session{
				ID:           sessionID,
				Name:         "测试会话",
				CreatorID:    "user-123",
				MCPServerURL: "http://localhost:9000",
				Status:       "active",
			}, nil
		},
	}

	router := setupTestRouter(mockService)

	req := httptest.NewRequest("GET", "/api/sessions/session-123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "session-123", response["id"])
	assert.Equal(t, "测试会话", response["name"])
}

// TestGetSession_Handler_NotFound 测试会话不存在
func TestGetSession_Handler_NotFound(t *testing.T) {
	mockService := &MockSessionService{
		mockGetSession: func(sessionID string) (*models.Session, error) {
			return nil, errors.ErrSessionNotFound
		},
	}

	router := setupTestRouter(mockService)

	req := httptest.NewRequest("GET", "/api/sessions/non-existent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, "SESSION_NOT_FOUND", errorObj["code"])
}

// TestListSessions_Handler_Success 测试列出会话成功
func TestListSessions_Handler_Success(t *testing.T) {
	mockService := &MockSessionService{
		mockListSessions: func(status string) ([]*models.Session, error) {
			return []*models.Session{
				{
					ID:        "session-1",
					Name:      "会话1",
					CreatorID: "user-1",
					Status:    "active",
				},
				{
					ID:        "session-2",
					Name:      "会话2",
					CreatorID: "user-2",
					Status:    "active",
				},
			}, nil
		},
	}

	router := setupTestRouter(mockService)

	req := httptest.NewRequest("GET", "/api/sessions", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "session-1", response[0]["id"])
	assert.Equal(t, "session-2", response[1]["id"])
}

// TestUpdateSession_Handler_Success 测试更新会话成功
func TestUpdateSession_Handler_Success(t *testing.T) {
	mockService := &MockSessionService{
		mockGetSession: func(sessionID string) (*models.Session, error) {
			return &models.Session{
				ID:     sessionID,
				Name:   "旧名称",
				Status: "active",
			}, nil
		},
		mockUpdateSession: func(sessionID string, req *service.UpdateSessionRequest) (*models.Session, error) {
			return &models.Session{
				ID:     sessionID,
				Name:   *req.Name,
				Status: "active",
			}, nil
		},
	}

	router := setupTestRouter(mockService)

	reqBody := `{"name": "新名称"}`
	req := httptest.NewRequest("PATCH", "/api/sessions/session-123", bytes.NewReader([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "session-123", response["id"])
	assert.Equal(t, "新名称", response["name"])
}

// TestDeleteSession_Handler_Success 测试删除会话成功
func TestDeleteSession_Handler_Success(t *testing.T) {
	mockService := &MockSessionService{
		mockGetSession: func(sessionID string) (*models.Session, error) {
			return &models.Session{
				ID:     sessionID,
				Name:   "测试会话",
				Status: "active",
			}, nil
		},
		mockDeleteSession: func(sessionID string) error {
			return nil
		},
	}

	router := setupTestRouter(mockService)

	req := httptest.NewRequest("DELETE", "/api/sessions/session-123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

// TestDeleteSession_Handler_NotFound 测试删除不存在的会话
func TestDeleteSession_Handler_NotFound(t *testing.T) {
	mockService := &MockSessionService{
		mockGetSession: func(sessionID string) (*models.Session, error) {
			return nil, errors.ErrSessionNotFound
		},
		mockDeleteSession: func(sessionID string) error {
			return errors.ErrSessionNotFound
		},
	}

	router := setupTestRouter(mockService)

	req := httptest.NewRequest("DELETE", "/api/sessions/non-existent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, "SESSION_NOT_FOUND", errorObj["code"])
}
