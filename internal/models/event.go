package models

import (
	"time"

	"github.com/google/uuid"
)

// Event 事件
type Event struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	SessionID uuid.UUID              `json:"session_id" db:"session_id"`
	Type      string                 `json:"event_type" db:"event_type"`
	Data      map[string]interface{} `json:"data" db:"data"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// NewEvent 创建新事件
func NewEvent(sessionID uuid.UUID, eventType string, data map[string]interface{}) *Event {
	return &Event{
		ID:        uuid.New(),
		SessionID: sessionID,
		Type:      eventType,
		Data:      data,
		CreatedAt: time.Now(),
	}
}

// 事件类型常量
const (
	EventTypeDiceRolled          = "dice.rolled"
	EventTypeCombatStarted       = "combat.started"
	EventTypeCombatRoundStarted  = "combat.round_started"
	EventTypeCombatTurnChanged   = "combat.turn_changed"
	EventTypeCombatAttackResolved = "combat.attack_resolved"
	EventTypeCombatCharacterDowned = "combat.character_downed"
	EventTypeCombatEnded         = "combat.ended"
	EventTypeCharacterMoved      = "character.moved"
	EventTypeCharacterHPChanged  = "character.hp_changed"
	EventTypeCharacterConditionAdded = "character.condition_added"
	EventTypeCharacterConditionRemoved = "character.condition_removed"
	EventTypeCharacterSpellCast  = "character.spell_cast"
	EventTypeMapImported         = "map.imported"
	EventTypeMapRevealedArea     = "map.revealed_area"
	EventTypeMapDoorOpened       = "map.door_opened"
	EventTypeMapLightChanged     = "map.light_changed"
)
