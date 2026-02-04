// Package ws 提供 WebSocket Hub 连接管理
package ws

import (
	"sync"

	"github.com/google/uuid"
)

// Hub WebSocket 连接中心
type Hub struct {
	// 注册的连接
	Connections map[string]*Connection

	// 会话连接索引 (session_id -> connection_ids)
	SessionConnections map[string][]string

	// 广播通道
	Broadcast chan Event

	// 注册通道
	Register chan *Connection

	// 注销通道
	Unregister chan *Connection

	// 互斥锁
	mu sync.RWMutex
}

// NewHub 创建新 Hub
func NewHub() *Hub {
	hub := &Hub{
		Connections:        make(map[string]*Connection),
		SessionConnections: make(map[string][]string),
		Broadcast:          make(chan Event, 256),
		Register:           make(chan *Connection),
		Unregister:         make(chan *Connection),
	}

	// 启动 Hub
	go hub.Run()

	return hub
}

// Run 运行 Hub
func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.Register:
			h.register(conn)

		case conn := <-h.Unregister:
			h.unregister(conn)

		case event := <-h.Broadcast:
			h.broadcast(&event)
		}
	}
}

// register 注册连接
func (h *Hub) register(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 添加连接
	h.Connections[conn.ID] = conn

	// 添加会话索引
	h.SessionConnections[conn.SessionID] = append(
		h.SessionConnections[conn.SessionID],
		conn.ID,
	)
}

// unregister 注销连接
func (h *Hub) unregister(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 移除连接
	if _, ok := h.Connections[conn.ID]; ok {
		delete(h.Connections, conn.ID)
		close(conn.Send)
	}

	// 移除会话索引
	if connections, ok := h.SessionConnections[conn.SessionID]; ok {
		// 过滤掉当前连接
		newConnections := make([]string, 0, len(connections))
		for _, connID := range connections {
			if connID != conn.ID {
				newConnections = append(newConnections, connID)
			}
		}

		if len(newConnections) == 0 {
			delete(h.SessionConnections, conn.SessionID)
		} else {
			h.SessionConnections[conn.SessionID] = newConnections
		}
	}

	conn.Close()
}

// broadcast 广播事件
func (h *Hub) broadcast(event *Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// 获取会话的所有连接
	connections, ok := h.SessionConnections[event.SessionID]
	if !ok {
		return
	}

	// 转换为服务器消息
	serverMsg := event.ToServerMessage()
	serverMsg.Data = event.Data

	// 发送给订阅了该事件类型的连接
	for _, connID := range connections {
		conn, ok := h.Connections[connID]
		if !ok {
			continue
		}

		// 检查订阅
		if !conn.IsSubscribed(event.Type) {
			continue
		}

		// 发送消息
		select {
		case conn.Send <- serverMsg:
		default:
			// 发送缓冲区已满，关闭连接
			h.Unregister <- conn
		}
	}
}

// GetConnection 获取连接
func (h *Hub) GetConnection(connID string) (*Connection, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conn, ok := h.Connections[connID]
	return conn, ok
}

// GetSessionConnections 获取会话的所有连接
func (h *Hub) GetSessionConnections(sessionID string) []*Connection {
	h.mu.RLock()
	defer h.mu.RUnlock()

	connIDs, ok := h.SessionConnections[sessionID]
	if !ok {
		return nil
	}

	connections := make([]*Connection, 0, len(connIDs))
	for _, connID := range connIDs {
		conn, ok := h.Connections[connID]
		if ok {
			connections = append(connections, conn)
		}
	}

	return connections
}

// BroadcastToSession 广播到会话的所有连接
func (h *Hub) BroadcastToSession(sessionID string, event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	connIDs, ok := h.SessionConnections[sessionID]
	if !ok {
		return
	}

	// 转换为服务器消息
	serverMsg := event.ToServerMessage()
	serverMsg.Data = event.Data

	for _, connID := range connIDs {
		conn, ok := h.Connections[connID]
		if !ok {
			continue
		}

		select {
		case conn.Send <- serverMsg:
		default:
			// 发送缓冲区已满
		}
	}
}

// Shutdown 关闭 Hub
func (h *Hub) Shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 关闭所有连接
	for _, conn := range h.Connections {
		conn.Close()
	}

	// 清空连接
	h.Connections = make(map[string]*Connection)
	h.SessionConnections = make(map[string][]string)
}

// ConnectionCount 获取连接数
func (h *Hub) ConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.Connections)
}

// SessionConnectionCount 获取会话连接数
func (h *Hub) SessionConnectionCount(sessionID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	connections, ok := h.SessionConnections[sessionID]
	if !ok {
		return 0
	}

	return len(connections)
}

// GenerateConnectionID 生成连接ID
func GenerateConnectionID() string {
	return "conn-" + uuid.New().String()
}
