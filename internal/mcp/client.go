// Package mcp 提供 MCP Client 接口和实现
package mcp

import (
	"context"

	"github.com/dnd-mcp/client/pkg/config"
)

// MCPClient MCP 客户端接口
type MCPClient interface {
	// Initialize 初始化 MCP 连接(握手)
	Initialize(ctx context.Context, sessionID, serverURL string) error

	// CallTool 调用 MCP Server 工具
	CallTool(ctx context.Context, sessionID string, toolName string, arguments map[string]interface{}) (map[string]interface{}, error)

	// SubscribeEvents 订阅 MCP Server 事件
	SubscribeEvents(ctx context.Context, sessionID string, eventTypes []string) (<-chan Event, error)

	// Close 关闭 MCP 连接
	Close(ctx context.Context) error
}

// Event MCP Server 事件
type Event struct {
	Type      string                 `json:"type"` // state_changed, combat_updated, etc.
	SessionID string                 `json:"session_id"`
	Data      map[string]interface{} `json:"data"`
}

// NewClient 创建 MCP 客户端
func NewClient(cfg *config.MCPConfig) (MCPClient, error) {
	if cfg.ServerURL == "mock://" {
		return NewMockClient(), nil
	}

	return NewHTTPClient(cfg.ServerURL, cfg.Timeout), nil
}
