// Package ws 提供 WebSocket 连接封装
package ws

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Connection WebSocket 连接封装
type Connection struct {
	// 连接ID
	ID string

	// 会话ID
	SessionID string

	// 玩家ID
	PlayerID string

	// WebSocket 连接
	WS *websocket.Conn

	// 发送缓冲区
	Send chan ServerMessage

	// 订阅的事件
	Subscriptions map[string]bool

	// Hub 引用
	Hub *Hub

	// 关闭信号
	closeOnce sync.Once
	closed    chan struct{}
}

// NewConnection 创建新连接
func NewConnection(id, sessionID, playerID string, ws *websocket.Conn, hub *Hub) *Connection {
	return &Connection{
		ID:            id,
		SessionID:     sessionID,
		PlayerID:      playerID,
		WS:            ws,
		Send:          make(chan ServerMessage, 256),
		Subscriptions: make(map[string]bool),
		Hub:           hub,
		closed:        make(chan struct{}),
	}
}

// ReadPump 读取 pump
func (c *Connection) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.WS.Close()
	}()

	c.WS.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.WS.SetPongHandler(func(string) error {
		c.WS.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.WS.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// 记录错误日志
			}
			break
		}

		// 解析客户端消息
		var clientMsg ClientMessage
		if err := json.Unmarshal(message, &clientMsg); err != nil {
			// 发送错误响应
			c.sendError("INVALID_MESSAGE", "Invalid message format")
			continue
		}

		// 处理消息
		c.handleMessage(&clientMsg)
	}
}

// WritePump 写入 pump
func (c *Connection) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.WS.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.WS.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub 关闭了连接
				c.WS.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 序列化消息
			data, err := json.Marshal(message)
			if err != nil {
				// 记录错误
				continue
			}

			if err := c.WS.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-ticker.C:
			c.WS.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.WS.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.closed:
			return
		}
	}
}

// Close 关闭连接
func (c *Connection) Close() {
	c.closeOnce.Do(func() {
		close(c.closed)
		close(c.Send)
	})
}

// handleMessage 处理客户端消息
func (c *Connection) handleMessage(msg *ClientMessage) {
	switch msg.Type {
	case "subscribe":
		c.handleSubscribe(msg)

	case "unsubscribe":
		c.handleUnsubscribe(msg)

	case "ping":
		c.handlePing(msg)

	default:
		c.sendError("UNKNOWN_MESSAGE_TYPE", "Unknown message type: "+msg.Type)
	}
}

// handleSubscribe 处理订阅
func (c *Connection) handleSubscribe(msg *ClientMessage) {
	// 解析订阅事件
	eventsData, ok := msg.Data["events"]
	if !ok {
		c.sendError("INVALID_SUBSCRIBE", "Missing events field")
		return
	}

	events, ok := eventsData.([]interface{})
	if !ok {
		c.sendError("INVALID_SUBSCRIBE", "Events must be an array")
		return
	}

	// 添加订阅
	for _, event := range events {
		if eventType, ok := event.(string); ok {
			c.Subscriptions[eventType] = true
		}
	}
}

// handleUnsubscribe 处理取消订阅
func (c *Connection) handleUnsubscribe(msg *ClientMessage) {
	// 解析取消订阅事件
	eventsData, ok := msg.Data["events"]
	if !ok {
		c.sendError("INVALID_UNSUBSCRIBE", "Missing events field")
		return
	}

	events, ok := eventsData.([]interface{})
	if !ok {
		c.sendError("INVALID_UNSUBSCRIBE", "Events must be an array")
		return
	}

	// 移除订阅
	for _, event := range events {
		if eventType, ok := event.(string); ok {
			delete(c.Subscriptions, eventType)
		}
	}
}

// handlePing 处理心跳
func (c *Connection) handlePing(msg *ClientMessage) {
	// 返回 pong
	pongMsg := ServerMessage{
		Type: "pong",
		Data: map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	select {
	case c.Send <- pongMsg:
	default:
		// 发送缓冲区已满
	}
}

// sendError 发送错误消息
func (c *Connection) sendError(code, message string) {
	errorMsg := ServerMessage{
		Type: "error",
		Data: map[string]interface{}{
			"code":      code,
			"message":   message,
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	select {
	case c.Send <- errorMsg:
	default:
		// 发送缓冲区已满
	}
}

// IsSubscribed 检查是否订阅了事件
func (c *Connection) IsSubscribed(eventType string) bool {
	subscribed, ok := c.Subscriptions[eventType]
	return ok && subscribed
}
