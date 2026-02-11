// Package dto 提供数据传输对象（Data Transfer Objects）
package dto

import (
	"time"

	"github.com/dnd-mcp/client/internal/models"
)

// SessionResponse 会话响应DTO
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

// MessageResponse 消息响应DTO
type MessageResponse struct {
	ID        string       `json:"id"`
	SessionID string       `json:"session_id"`
	Role      string       `json:"role"`
	Content   string       `json:"content"`
	ToolCalls []ToolCall   `json:"tool_calls,omitempty"`
	PlayerID  string       `json:"player_id,omitempty"`
	CreatedAt string       `json:"created_at"`
}

// ToolCall 工具调用DTO
type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ErrorResponse 错误响应DTO
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail 错误详情DTO
type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// SuccessResponse 成功响应DTO
type SuccessResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// FromSession 从模型转换为响应DTO
func FromSession(session *models.Session) SessionResponse {
	return SessionResponse{
		ID:           session.ID,
		Name:         session.Name,
		CreatorID:    session.CreatorID,
		MCPServerURL: session.MCPServerURL,
		WebSocketKey: session.WebSocketKey,
		MaxPlayers:   session.MaxPlayers,
		Settings:     session.Settings,
		CreatedAt:    session.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    session.UpdatedAt.UTC().Format(time.RFC3339),
		Status:       session.Status,
	}
}

// FromSessions 从模型切片转换为响应DTO切片
func FromSessions(sessions []*models.Session) []SessionResponse {
	responses := make([]SessionResponse, len(sessions))
	for i, session := range sessions {
		responses[i] = FromSession(session)
	}
	return responses
}

// FromMessage 从模型转换为响应DTO
func FromMessage(message *models.Message) MessageResponse {
	toolCalls := make([]ToolCall, len(message.ToolCalls))
	for i, tc := range message.ToolCalls {
		toolCalls[i] = ToolCall{
			ID:        tc.ID,
			Name:      tc.Name,
			Arguments: tc.Arguments,
		}
	}

	return MessageResponse{
		ID:        message.ID,
		SessionID: message.SessionID,
		Role:      message.Role,
		Content:   message.Content,
		ToolCalls: toolCalls,
		PlayerID:  message.PlayerID,
		CreatedAt: message.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// FromMessages 从模型切片转换为响应DTO切片
func FromMessages(messages []*models.Message) []MessageResponse {
	responses := make([]MessageResponse, len(messages))
	for i, message := range messages {
		responses[i] = FromMessage(message)
	}
	return responses
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code, message string, details interface{}) ErrorResponse {
	return ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(message string, data interface{}) SuccessResponse {
	return SuccessResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	}
}
