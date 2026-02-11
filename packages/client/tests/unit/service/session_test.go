// Package service_test 提供 Service 层单元测试
package service_test

import (
	"context"
	"testing"

	"github.com/dnd-mcp/client/internal/models"
	"github.com/dnd-mcp/client/internal/service"
	"github.com/dnd-mcp/client/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSessionRepository Mock Repository 接口
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session *models.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) Get(ctx context.Context, id string) (*models.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionRepository) List(ctx context.Context) ([]*models.Session, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Session), args.Error(1)
}

func (m *MockSessionRepository) Update(ctx context.Context, session *models.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// TestSessionService_CreateSession_Success 测试成功创建会话
func TestSessionService_CreateSession_Success(t *testing.T) {
	// 创建 Mock
	mockRepo := new(MockSessionRepository)
	sessionService := service.NewSessionService(mockRepo)

	// 设置 Mock 期望
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Session")).Return(nil)

	// 测试
	req := &service.CreateSessionRequest{
		Name:         "测试会话",
		CreatorID:    "user-123",
		MCPServerURL: "http://localhost:9000",
		MaxPlayers:   5,
	}

	session, err := sessionService.CreateSession(context.Background(), req)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "测试会话", session.Name)
	assert.Equal(t, "user-123", session.CreatorID)
	assert.Equal(t, "http://localhost:9000", session.MCPServerURL)
	assert.Equal(t, 5, session.MaxPlayers)
	assert.Equal(t, "active", session.Status)
	assert.NotEmpty(t, session.ID)
	assert.NotEmpty(t, session.WebSocketKey)

	// 验证 Mock 调用
	mockRepo.AssertExpectations(t)
}

// TestSessionService_CreateSession_InvalidMaxPlayers 测试无效的 max_players
func TestSessionService_CreateSession_InvalidMaxPlayers(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	sessionService := service.NewSessionService(mockRepo)

	tests := []struct {
		name       string
		maxPlayers int
		wantErr    error
	}{
		{"max_players 小于 1", 0, errors.ErrInvalidMaxPlayers},
		{"max_players 为负数", -1, errors.ErrInvalidMaxPlayers},
		{"max_players 大于 10", 11, errors.ErrInvalidMaxPlayers},
		{"max_players 等于 10", 10, nil}, // 边界值,应该成功
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &service.CreateSessionRequest{
				Name:         "测试会话",
				CreatorID:    "user-123",
				MCPServerURL: "http://localhost:9000",
				MaxPlayers:   tt.maxPlayers,
			}

			if tt.wantErr != nil {
				// 只为成功的测试设置 Mock 期望
				if tt.maxPlayers == 10 {
					mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Session")).Return(nil)
				}
			} else {
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Session")).Return(nil)
			}

			session, err := sessionService.CreateSession(context.Background(), req)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, session)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, session)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestSessionService_GetSession_Success 测试成功获取会话
func TestSessionService_GetSession_Success(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	sessionService := service.NewSessionService(mockRepo)

	// 设置 Mock 期望
	expectedSession := &models.Session{
		ID:           "session-123",
		Name:         "测试会话",
		CreatorID:    "user-123",
		MCPServerURL: "http://localhost:9000",
		MaxPlayers:   5,
		Status:       "active",
	}
	mockRepo.On("Get", mock.Anything, "session-123").Return(expectedSession, nil)

	// 测试
	session, err := sessionService.GetSession(context.Background(), "session-123")

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "session-123", session.ID)
	assert.Equal(t, "测试会话", session.Name)

	// 验证 Mock 调用
	mockRepo.AssertExpectations(t)
}

// TestSessionService_GetSession_NotFound 测试会话不存在
func TestSessionService_GetSession_NotFound(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	sessionService := service.NewSessionService(mockRepo)

	// 设置 Mock 期望
	mockRepo.On("Get", mock.Anything, "non-existent").Return(nil, errors.ErrSessionNotFound)

	// 测试
	session, err := sessionService.GetSession(context.Background(), "non-existent")

	// 断言
	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Equal(t, errors.ErrSessionNotFound, err)

	// 验证 Mock 调用
	mockRepo.AssertExpectations(t)
}

// TestSessionService_ListSessions_Success 测试成功列出所有会话
func TestSessionService_ListSessions_Success(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	sessionService := service.NewSessionService(mockRepo)

	// 设置 Mock 期望
	expectedSessions := []*models.Session{
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
	}
	mockRepo.On("List", mock.Anything).Return(expectedSessions, nil)

	// 测试
	sessions, err := sessionService.ListSessions(context.Background(), "")

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, sessions)
	assert.Len(t, sessions, 2)
	assert.Equal(t, "session-1", sessions[0].ID)
	assert.Equal(t, "session-2", sessions[1].ID)

	// 验证 Mock 调用
	mockRepo.AssertExpectations(t)
}

// TestSessionService_UpdateSession_Success 测试成功更新会话
func TestSessionService_UpdateSession_Success(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	sessionService := service.NewSessionService(mockRepo)

	// 设置 Mock 期望 - Get
	existingSession := &models.Session{
		ID:         "session-123",
		Name:       "旧名称",
		MaxPlayers: 4,
		Status:     "active",
	}
	mockRepo.On("Get", mock.Anything, "session-123").Return(existingSession, nil)

	// 设置 Mock 期望 - Update
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Session")).Return(nil)

	// 测试
	req := &service.UpdateSessionRequest{
		Name:       stringPtr("新名称"),
		MaxPlayers: intPtr(6),
	}

	session, err := sessionService.UpdateSession(context.Background(), "session-123", req)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "新名称", session.Name)
	assert.Equal(t, 6, session.MaxPlayers)

	// 验证 Mock 调用
	mockRepo.AssertExpectations(t)
}

// TestSessionService_UpdateSession_NotFound 测试更新不存在的会话
func TestSessionService_UpdateSession_NotFound(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	sessionService := service.NewSessionService(mockRepo)

	// 设置 Mock 期望
	mockRepo.On("Get", mock.Anything, "non-existent").Return(nil, errors.ErrSessionNotFound)

	// 测试
	req := &service.UpdateSessionRequest{
		Name: stringPtr("新名称"),
	}

	session, err := sessionService.UpdateSession(context.Background(), "non-existent", req)

	// 断言
	assert.Error(t, err)
	assert.Nil(t, session)

	// 验证 Mock 调用
	mockRepo.AssertExpectations(t)
}

// TestSessionService_DeleteSession_Success 测试成功删除会话
func TestSessionService_DeleteSession_Success(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	sessionService := service.NewSessionService(mockRepo)

	// 设置 Mock 期望 - Get
	existingSession := &models.Session{
		ID:     "session-123",
		Name:   "测试会话",
		Status: "active",
	}
	mockRepo.On("Get", mock.Anything, "session-123").Return(existingSession, nil)

	// 设置 Mock 期望 - Delete
	mockRepo.On("Delete", mock.Anything, "session-123").Return(nil)

	// 测试
	err := sessionService.DeleteSession(context.Background(), "session-123")

	// 断言
	assert.NoError(t, err)

	// 验证 Mock 调用
	mockRepo.AssertExpectations(t)
}

// TestSessionService_DeleteSession_NotFound 测试删除不存在的会话
func TestSessionService_DeleteSession_NotFound(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	sessionService := service.NewSessionService(mockRepo)

	// 设置 Mock 期望
	mockRepo.On("Get", mock.Anything, "non-existent").Return(nil, errors.ErrSessionNotFound)

	// 测试
	err := sessionService.DeleteSession(context.Background(), "non-existent")

	// 断言
	assert.Error(t, err)
	assert.Equal(t, errors.ErrSessionNotFound, err)

	// 验证 Mock 调用
	mockRepo.AssertExpectations(t)
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
