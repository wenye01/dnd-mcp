// Package mcp 提供 MCP 事件监听器
package mcp

import (
	"context"
	"log"

	"github.com/dnd-mcp/client/internal/ws"
)

// EventListener 事件监听器
type EventListener struct {
	mcpClient MCPClient
	wsHub     *ws.Hub
	eventChan <-chan Event
	sessionID string
}

// NewEventListener 创建事件监听器
func NewEventListener(mcpClient MCPClient, wsHub *ws.Hub) *EventListener {
	return &EventListener{
		mcpClient: mcpClient,
		wsHub:     wsHub,
	}
}

// Start 启动事件监听
func (l *EventListener) Start(ctx context.Context, sessionID string) error {
	l.sessionID = sessionID

	// 订阅所有事件类型
	eventTypes := []string{
		"state_changed",
		"combat_updated",
		"character_moved",
		"dice_rolled",
	}

	eventChan, err := l.mcpClient.SubscribeEvents(ctx, sessionID, eventTypes)
	if err != nil {
		return err
	}

	l.eventChan = eventChan

	// 启动事件处理循环
	go l.processEvents(ctx)

	return nil
}

// processEvents 处理事件
func (l *EventListener) processEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-l.eventChan:
			if !ok {
				return
			}

			// 转换为 WebSocket 事件并广播
			wsEvent := ws.Event{
				SessionID: l.sessionID,
				Type:      event.Type,
				Data:      event.Data,
			}

			l.wsHub.Broadcast <- wsEvent

			log.Printf("[MCP] 事件已广播: type=%s, session=%s", event.Type, l.sessionID)
		}
	}
}
