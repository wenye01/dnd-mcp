// Package mcp 提供 Mock MCP 实现
package mcp

import (
	"context"
	"fmt"
	"time"
)

// MockClient Mock MCP 客户端
type MockClient struct{}

// NewMockClient 创建 Mock 客户端
func NewMockClient() *MockClient {
	return &MockClient{}
}

// Initialize 初始化
func (m *MockClient) Initialize(ctx context.Context, sessionID, serverURL string) error {
	return nil
}

// CallTool 调用工具
func (m *MockClient) CallTool(ctx context.Context, sessionID string, toolName string, arguments map[string]interface{}) (map[string]interface{}, error) {
	// 根据工具名称返回模拟结果
	switch toolName {
	case "roll_dice":
		return map[string]interface{}{
			"success": true,
			"result": map[string]interface{}{
				"formula":  arguments["formula"],
				"total":    18,
				"rolls":    []int{15},
				"modifier": 3,
			},
		}, nil

	case "resolve_attack":
		return map[string]interface{}{
			"success": true,
			"result": map[string]interface{}{
				"attacker": arguments["attacker"],
				"target":   arguments["target"],
				"hit":      true,
				"damage":   8,
			},
		}, nil

	default:
		return map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("工具 %s 执行成功", toolName),
		}, nil
	}
}

// SubscribeEvents 订阅事件
func (m *MockClient) SubscribeEvents(ctx context.Context, sessionID string, eventTypes []string) (<-chan Event, error) {
	eventChan := make(chan Event, 10)

	go func() {
		defer close(eventChan)

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for i := 0; ; i++ {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// 模拟生成事件
				event := Event{
					Type:      eventTypes[i%len(eventTypes)],
					SessionID: sessionID,
					Data: map[string]interface{}{
						"mock":  true,
						"index": i,
					},
				}

				select {
				case eventChan <- event:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return eventChan, nil
}

// Close 关闭连接
func (m *MockClient) Close(ctx context.Context) error {
	return nil
}
