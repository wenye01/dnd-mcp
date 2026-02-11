// Package ws 提供 Mock 事件生成器
package ws

import (
	"math/rand"
	"time"
)

// MockEventGenerator Mock 事件生成器
type MockEventGenerator struct {
	Hub *Hub
}

// NewMockEventGenerator 创建 Mock 事件生成器
func NewMockEventGenerator(hub *Hub) *MockEventGenerator {
	return &MockEventGenerator{
		Hub: hub,
	}
}

// Start 启动事件生成
func (m *MockEventGenerator) Start() {
	// 每 10 秒生成一次事件
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 随机生成不同类型的事件
		eventType := m.randomEventType()

		var event Event
		switch eventType {
		case "state_changed":
			event = m.generateStateChangeEvent()
		case "combat_updated":
			event = m.generateCombatUpdatedEvent()
		case "dice_rolled":
			event = m.generateDiceRolledEvent()
		}

		// 广播事件
		m.Hub.Broadcast <- event
	}
}

// randomEventType 随机事件类型
func (m *MockEventGenerator) randomEventType() string {
	types := []string{"state_changed", "combat_updated", "dice_rolled"}
	return types[rand.Intn(len(types))]
}

// generateStateChangeEvent 生成状态变更事件
func (m *MockEventGenerator) generateStateChangeEvent() Event {
	locations := []string{"幽暗森林", "地下城入口", "古老遗迹", "龙穴", "村庄广场"}
	gameTimes := []string{"第1天 08:00", "第1天 12:30", "第2天 06:15", "第3天 20:45", "第5天 14:20"}
	combatStates := []string{"active", "inactive", "victory", "defeat"}

	event := Event{
		ID:        "event-" + randomID(),
		SessionID: "mock-session-id",
		Type:      "state_changed",
		Data: map[string]interface{}{
			"changes": map[string]interface{}{
				"location":  locations[rand.Intn(len(locations))],
				"game_time": gameTimes[rand.Intn(len(gameTimes))],
				"combat":    combatStates[rand.Intn(len(combatStates))],
			},
			"timestamp": time.Now().Format(time.RFC3339),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return event
}

// generateCombatUpdatedEvent 生成战斗更新事件
func (m *MockEventGenerator) generateCombatUpdatedEvent() Event {
	actions := []string{"attack", "defend", "use_skill", "cast_spell"}
	attackers := []string{"player-123", "player-456", "player-789"}
	targets := []string{"goblin-1", "orc-2", "dragon-3", "wolf-4"}

	event := Event{
		ID:        "event-" + randomID(),
		SessionID: "mock-session-id",
		Type:      "combat_updated",
		Data: map[string]interface{}{
			"session_id": "mock-session-id",
			"combat_id":  "combat-" + randomID(),
			"action":     actions[rand.Intn(len(actions))],
			"attacker":   attackers[rand.Intn(len(attackers))],
			"target":     targets[rand.Intn(len(targets))],
			"result": map[string]interface{}{
				"hit":    rand.Intn(2) == 1,
				"damage": rand.Intn(20) + 1,
			},
			"timestamp": time.Now().Format(time.RFC3339),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return event
}

// generateDiceRolledEvent 生成骰子投掷事件
func (m *MockEventGenerator) generateDiceRolledEvent() Event {
	rollTypes := []string{"d20", "d12", "d10", "d8", "d6", "d4"}
	players := []string{"player-123", "player-456", "player-789"}

	rollType := rollTypes[rand.Intn(len(rollTypes))]
	var maxRoll int
	switch rollType {
	case "d20":
		maxRoll = 20
	case "d12":
		maxRoll = 12
	case "d10":
		maxRoll = 10
	case "d8":
		maxRoll = 8
	case "d6":
		maxRoll = 6
	case "d4":
		maxRoll = 4
	default:
		maxRoll = 20
	}

	event := Event{
		ID:        "event-" + randomID(),
		SessionID: "mock-session-id",
		Type:      "dice_rolled",
		Data: map[string]interface{}{
			"session_id": "mock-session-id",
			"player_id":  players[rand.Intn(len(players))],
			"roll_type":  rollType,
			"result":     rand.Intn(maxRoll) + 1,
			"details":    "玩家投掷 " + rollType + " 骰子",
			"timestamp":  time.Now().Format(time.RFC3339),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return event
}

// randomID 生成随机ID
func randomID() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
