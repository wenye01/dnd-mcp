package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	gorillaws "github.com/gorilla/websocket"

	"github.com/dnd-mcp/client/internal/client/websocket"
)

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	Hub *websocket.Hub
}

// NewWebSocketHandler 创建WebSocket处理器
func NewWebSocketHandler(hub *websocket.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		Hub: hub,
	}
}

var upgrader = gorillaws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
}

// HandleWebSocket 处理WebSocket连接
// GET /api/sessions/:id/ws
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	id := c.Param("id")
	sessionID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// 创建客户端
	client := &websocket.Client{
		ID:        uuid.New(),
		SessionID: sessionID,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		Hub:       h.Hub,
	}

	// 注册客户端
	h.Hub.Register <- client

	// 启动读写循环
	go client.WritePump()
	go client.ReadPump()
}
