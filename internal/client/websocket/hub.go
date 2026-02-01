package websocket

import (
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Hub WebSocket连接中心
type Hub struct {
	// 注册的客户端
	clients map[*Client]bool

	// 按session_id分组的客户端
	sessionClients map[uuid.UUID][]*Client

	// 广播通道
	broadcast chan *Message

	// 注册通道（导出）
	Register chan *Client

	// 注销通道（导出）
	Unregister chan *Client

	mu sync.RWMutex
}

// Client WebSocket客户端
type Client struct {
	ID        uuid.UUID
	SessionID uuid.UUID
	Conn      *websocket.Conn
	Send      chan []byte
	Hub       *Hub
}

// Message WebSocket消息
type Message struct {
	SessionID uuid.UUID
	Data      []byte
}

// NewHub 创建新的Hub
func NewHub() *Hub {
	return &Hub{
		clients:        make(map[*Client]bool),
		sessionClients: make(map[uuid.UUID][]*Client),
		broadcast:      make(chan *Message, 256),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
	}
}

// Run 运行Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToSession(message)
		}
	}
}

// registerClient 注册客户端
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true
	h.sessionClients[client.SessionID] = append(h.sessionClients[client.SessionID], client)

	log.Printf("Client %s registered for session %s", client.ID, client.SessionID)
}

// unregisterClient 注销客户端
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.Send)

		// 从session组中移除
		clients := h.sessionClients[client.SessionID]
		for i, c := range clients {
			if c == client {
				h.sessionClients[client.SessionID] = append(clients[:i], clients[i+1:]...)
				break
			}
		}

		log.Printf("Client %s unregistered from session %s", client.ID, client.SessionID)
	}
}

// broadcastToSession 向指定会话的所有客户端广播消息
func (h *Hub) broadcastToSession(message *Message) {
	h.mu.RLock()
	clients := h.sessionClients[message.SessionID]
	h.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.Send <- message.Data:
		default:
			// 发送失败,关闭连接
			h.Unregister <- client
		}
	}
}

// BroadcastToSession 向外部暴露的广播方法
func (h *Hub) BroadcastToSession(sessionID uuid.UUID, data []byte) {
	h.broadcast <- &Message{
		SessionID: sessionID,
		Data:      data,
	}
}

// WritePump 写入循环
func (c *Client) WritePump() {
	defer c.Conn.Close()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("Error writing to websocket: %v", err)
				return
			}
		}
	}
}

// ReadPump 读取循环
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
