// Package server 提供 Server API 客户端接口定义
package server

import (
	"context"

	"github.com/dnd-mcp/client/pkg/config"
)

// ServerClient Server API 客户端接口
type ServerClient interface {
	// Initialize 执行 MCP 握手
	Initialize(ctx context.Context) error

	// GetContext 获取压缩后的上下文
	GetContext(ctx context.Context, campaignID string, limit int, includeCombat bool) (*Context, error)

	// GetRawContext 获取原始上下文（完整模式）
	GetRawContext(ctx context.Context, campaignID string) (*RawContext, error)

	// SaveMessage 保存消息到 Server
	SaveMessage(ctx context.Context, campaignID string, msg *Message) error

	// CallTool 调用 Server MCP Tool（复用现有 mcp.MCPClient）
	CallTool(ctx context.Context, campaignID, toolName string, args map[string]any) (map[string]any, error)

	// Close 关闭客户端连接
	Close(ctx context.Context) error
}

// NewClient 创建 Server 客户端
func NewClient(cfg *config.ServerConfig) (ServerClient, error) {
	if cfg.ServerURL == "mock://" {
		return NewMockClient(), nil
	}

	return NewHTTPClient(cfg.ServerURL, cfg.Timeout), nil
}
