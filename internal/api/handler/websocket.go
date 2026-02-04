// Package handler 提供 WebSocket Handler
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/dnd-mcp/client/internal/store"
	"github.com/dnd-mcp/client/internal/ws"
)

// WebSocketUpgrader WebSocket 升级器
var WebSocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源（生产环境应限制）
	},
}

// WSHandler WebSocket 处理器
type WSHandler struct {
	hub          *ws.Hub
	sessionStore store.SessionStore
}

// NewWSHandler 创建 WebSocket 处理器
func NewWSHandler(hub *ws.Hub, sessionStore store.SessionStore) *WSHandler {
	return &WSHandler{
		hub:          hub,
		sessionStore: sessionStore,
	}
}

// HandleWebSocket 处理 WebSocket 连接
// GET /ws/sessions/:session-id?key={ws-key}
func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	// 获取参数
	sessionID := c.Param("id")
	key := c.Query("key")

	// 验证参数
	if sessionID == "" || key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "MISSING_PARAMETERS",
				"message": "session_id and key are required",
			},
		})
		return
	}

	// 验证会话是否存在
	session, err := h.sessionStore.Get(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "SESSION_NOT_FOUND",
				"message": "Session not found",
			},
		})
		return
	}

	// 验证 websocket_key
	if session.WebSocketKey != key {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "INVALID_KEY",
				"message": "Invalid WebSocket key",
			},
		})
		return
	}

	// 升级 HTTP 连接到 WebSocket
	wsConn, err := WebSocketUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "UPGRADE_FAILED",
				"message": "Failed to upgrade to WebSocket: " + err.Error(),
			},
		})
		return
	}

	// 创建连接对象
	connID := ws.GenerateConnectionID()
	playerID := c.Query("player_id")
	if playerID == "" {
		playerID = "anonymous"
	}

	conn := ws.NewConnection(connID, sessionID, playerID, wsConn, h.hub)

	// 注册连接
	h.hub.Register <- conn

	// 发送连接成功消息
	connectedMsg := ws.ServerMessage{
		Type: "connected",
		Data: map[string]interface{}{
			"session_id": sessionID,
			"connection_id": connID,
			"player_id":  playerID,
		},
	}
	conn.Send <- connectedMsg

	// 启动读写 goroutine
	go conn.WritePump()
	go conn.ReadPump()
}

// BroadcastTestMessage 广播测试消息（仅用于测试）
func (h *WSHandler) BroadcastTestMessage(c *gin.Context) {
	var req struct {
		SessionID string                 `json:"session_id" binding:"required"`
		Type      string                 `json:"type" binding:"required"`
		Data      map[string]interface{} `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request: " + err.Error(),
			},
		})
		return
	}

	// 创建事件
	event := ws.NewEvent(req.SessionID, req.Type, req.Data)

	// 广播事件
	h.hub.Broadcast <- *event

	c.JSON(http.StatusOK, gin.H{
		"message": "Event broadcasted",
		"event_id": event.ID,
	})
}

// GetConnectionsInfo 获取连接信息（仅用于测试）
func (h *WSHandler) GetConnectionsInfo(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "MISSING_SESSION_ID",
				"message": "session_id is required",
			},
		})
		return
	}

	connections := h.hub.GetSessionConnections(sessionID)
	if connections == nil {
		c.JSON(http.StatusOK, gin.H{
			"session_id": sessionID,
			"connections": []interface{}{},
			"count": 0,
		})
		return
	}

	connInfos := make([]map[string]interface{}, 0, len(connections))
	for _, conn := range connections {
		connInfo := map[string]interface{}{
			"connection_id": conn.ID,
			"session_id":   conn.SessionID,
			"player_id":    conn.PlayerID,
			"subscriptions": conn.Subscriptions,
		}
		connInfos = append(connInfos, connInfo)
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id":  sessionID,
		"connections": connInfos,
		"count":       len(connInfos),
	})
}

// BroadcastMessage 广播消息到会话
func (h *WSHandler) BroadcastMessage(c *gin.Context) {
	sessionID := c.Param("id")
	var req struct {
		Type string                 `json:"type" binding:"required"`
		Data map[string]interface{} `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request: " + err.Error(),
			},
		})
		return
	}

	// 验证会话是否存在
	_, err := h.sessionStore.Get(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "SESSION_NOT_FOUND",
				"message": "Session not found",
			},
		})
		return
	}

	// 创建事件
	event := ws.NewEvent(sessionID, req.Type, req.Data)

	// 添加到事件数据
	event.Data["session_id"] = sessionID

	// 广播事件
	h.hub.Broadcast <- *event

	c.JSON(http.StatusOK, gin.H{
		"message":  "Event broadcasted successfully",
		"event_id": event.ID,
		"type":     event.Type,
	})
}

// parseMessage 解析消息
func parseMessage(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
