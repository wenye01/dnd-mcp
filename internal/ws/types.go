// Package ws 提供 WebSocket 类型和接口定义
package ws

import (
	"time"

	"github.com/google/uuid"
)

// ClientMessage 客户端发送的消息
type ClientMessage struct {
	Type string                 `json:"type"` // subscribe, unsubscribe, ping
	Data map[string]interface{} `json:"data"`
}

// ClientMessageData 客户端消息的 Data 字段详细结构
type ClientMessageData struct {
	// SubscribeData 订阅事件数据
	SubscribeData struct {
		Events []string `json:"events"` // state_changed, combat_updated, dice_rolled, etc.
	} `json:"-"`

	// PingData 心跳数据
	PingData struct {
		Timestamp string `json:"timestamp"`
	} `json:"-"`
}

// ServerMessage 服务器发送的消息
type ServerMessage struct {
	Type string                 `json:"type"` // new_message, state_changed, combat_updated, dice_rolled, pong, error
	Data map[string]interface{} `json:"data"`
}

// Event 事件结构
type Event struct {
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id"`
	Type      string                 `json:"type"` // state_changed, combat_updated, dice_rolled, etc.
	Data      map[string]interface{} `json:"data"`
	Timestamp string                 `json:"timestamp"`
}

// NewEvent 创建新事件
func NewEvent(sessionID, eventType string, data map[string]interface{}) *Event {
	return &Event{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// ToServerMessage 转换为服务器消息
func (e *Event) ToServerMessage() ServerMessage {
	return ServerMessage{
		Type: e.Type,
		Data: map[string]interface{}{
			"session_id": e.SessionID,
			"timestamp":  e.Timestamp,
		},
	}
}

// SubscribeEvent 订阅事件
type SubscribeEvent struct {
	Events []string `json:"events"` // state_changed, combat_updated, dice_rolled, etc.
}

// UnsubscribeEvent 取消订阅事件
type UnsubscribeEvent struct {
	Events []string `json:"events"`
}

// PongEvent 心跳响应
type PongEvent struct {
	Timestamp string `json:"timestamp"`
}

// NewMessageEventData 新消息事件数据
type NewMessageEventData struct {
	MessageID string `json:"message_id"`
	SessionID string `json:"session_id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	PlayerID  string `json:"player_id,omitempty"`
	Timestamp string `json:"timestamp"`
}

// StateChangedEventData 状态变更事件数据
type StateChangedEventData struct {
	SessionID string                 `json:"session_id"`
	Changes   map[string]interface{} `json:"changes"`   // location, game_time, combat, etc.
	Timestamp string                 `json:"timestamp"`
}

// CombatUpdatedEventData 战斗事件数据
type CombatUpdatedEventData struct {
	SessionID string                 `json:"session_id"`
	CombatID  string                 `json:"combat_id"`
	Action    string                 `json:"action"`    // attack, defend, use_skill, etc.
	Attacker  string                 `json:"attacker"`
	Target    string                 `json:"target"`
	Result    map[string]interface{} `json:"result"`    // hit, damage, effect, etc.
	Timestamp string                 `json:"timestamp"`
}

// DiceRolledEventData 骰子事件数据
type DiceRolledEventData struct {
	SessionID string `json:"session_id"`
	PlayerID  string `json:"player_id"`
	RollType  string `json:"roll_type"`  // d20, ability_check, saving_throw, etc.
	Result    int    `json:"result"`     // 骰子结果
	Details   string `json:"details"`    // 详细描述
	Timestamp string `json:"timestamp"`
}

// ErrorEventData 错误事件数据
type ErrorEventData struct {
	Code    string `json:"code"`    // INVALID_MESSAGE, SUBSCRIPTION_FAILED, etc.
	Message string `json:"message"`
	Timestamp string `json:"timestamp"`
}
