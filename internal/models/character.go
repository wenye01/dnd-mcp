package models

import "github.com/google/uuid"

// Character 角色数据
type Character struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	PlayerID  *string   `json:"player_id,omitempty"` // 玩家角色有此字段

	// 属性
	HP        int       `json:"hp"`
	MaxHP     int       `json:"max_hp"`
	AC        int       `json:"ac"`
	Speed     int       `json:"speed"`
	Abilities Abilities `json:"abilities"`

	// 位置
	Position Position `json:"position"`

	// 状态效果
	Conditions []string `json:"conditions"`

	// 法术位
	SpellSlots map[string]int `json:"spell_slots,omitempty"`
}

// Abilities 能力值
type Abilities struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisism"`
	Charisma     int `json:"charisma"`
}

// Position 位置坐标
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// NewCharacter 创建新角色
func NewCharacter(id uuid.UUID, name string) *Character {
	return &Character{
		ID:      id,
		Name:    name,
		PlayerID: nil,
		HP:      10,
		MaxHP:   10,
		AC:      10,
		Speed:   30,
		Abilities: Abilities{
			Strength:     10,
			Dexterity:    10,
			Constitution: 10,
			Intelligence: 10,
			Wisdom:       10,
			Charisma:     10,
		},
		Position:   Position{X: 0, Y: 0},
		Conditions: make([]string, 0),
		SpellSlots: make(map[string]int),
	}
}
