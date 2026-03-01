// Package models provides token bar model definitions
// Token bars display attribute values like HP, AC, etc.
// 规则参考: Foundry VTT Token Bar Structure
package models

import (
	"fmt"
)

// BarID identifies which bar a setting applies to
type BarID string

const (
	// Bar1Primary Primary bar (usually HP)
	Bar1Primary BarID = "bar1"
	// Bar2Secondary Secondary bar (usually temporary HP or other resource)
	Bar2Secondary BarID = "bar2"
)

// AttributeType represents the type of attribute the bar displays
type AttributeType string

const (
	// AttributeHP Hit points
	AttributeHP AttributeType = "hp"
	// AttributeTempHP Temporary hit points
	AttributeTempHP AttributeType = "tempHP"
	// AttributeAC Armor class
	AttributeAC AttributeType = "ac"
	// AttributeResource Generic resource (mana, stamina, etc.)
	AttributeResource AttributeType = "resource"
)

// TokenBar represents a progress bar on a token
// Used to display hit points, armor class, or other resources
// 规则参考: PHB 第1章 - Hit Points and Hit Dice
type TokenBar struct {
	// ID identifies which bar this is
	ID BarID `json:"id"`

	// Attribute is the character attribute this bar displays
	Attribute AttributeType `json:"attribute"`

	// Value is the current value of the bar
	Value int `json:"value"`

	// Max is the maximum value of the bar
	Max int `json:"max"`

	// CustomAttribute allows referencing a custom attribute by name
	CustomAttribute string `json:"custom_attribute,omitempty"`

	// Visible controls whether the bar is shown to players
	Visible bool `json:"visible"`

	// Color is the bar color in hex format
	Color string `json:"color,omitempty"`

	// Label is a custom label for the bar
	Label string `json:"label,omitempty"`
}

// NewTokenBar creates a new token bar
func NewTokenBar(id BarID, attribute AttributeType, value, max int) *TokenBar {
	return &TokenBar{
		ID:       id,
		Attribute: attribute,
		Value:    value,
		Max:      max,
		Visible:  true,
	}
}

// NewHPBar creates a hit point bar with default settings
func NewHPBar(currentHP, maxHP int) *TokenBar {
	bar := &TokenBar{
		ID:       Bar1Primary,
		Attribute: AttributeHP,
		Value:    currentHP,
		Max:      maxHP,
		Visible:  true,
		Color:    "#ff0000", // Red for HP
		Label:    "HP",
	}
	return bar
}

// NewACBar creates an armor class bar
// AC is typically displayed as a static value
func NewACBar(ac int) *TokenBar {
	return &TokenBar{
		ID:       Bar2Secondary,
		Attribute: AttributeAC,
		Value:    ac,
		Max:      ac,
		Visible:  true,
		Color:    "#00ff00", // Green for AC
		Label:    "AC",
	}
}

// Validate validates the token bar
func (b *TokenBar) Validate() error {
	if b.ID != Bar1Primary && b.ID != Bar2Secondary {
		return NewValidationError("token_bar.id", "must be bar1 or bar2")
	}

	if b.Value < 0 {
		return NewValidationError("token_bar.value", "cannot be negative")
	}

	if b.Max < 0 {
		return NewValidationError("token_bar.max", "cannot be negative")
	}

	if b.Value > b.Max {
		return NewValidationError("token_bar.value", "cannot exceed max")
	}

	return nil
}

// IsEmpty returns true if the bar has no value
func (b *TokenBar) IsEmpty() bool {
	return b.Value <= 0
}

// IsFull returns true if the bar is at maximum
func (b *TokenBar) IsFull() bool {
	return b.Value >= b.Max
}

// GetPercentage returns the bar's fill percentage (0-100)
func (b *TokenBar) GetPercentage() int {
	if b.Max == 0 {
		return 0
	}
	percentage := (b.Value * 100) / b.Max
	if percentage < 0 {
		return 0
	}
	if percentage > 100 {
		return 100
	}
	return percentage
}

// SetValue sets the bar value and ensures it stays within bounds
func (b *TokenBar) SetValue(value int) {
	if value < 0 {
		value = 0
	}
	if value > b.Max {
		value = b.Max
	}
	b.Value = value
}

// Adjust adjusts the bar value by the given amount
// Returns true if the bar became empty or full after adjustment
func (b *TokenBar) Adjust(delta int) (becameEmpty, becameFull bool) {
	b.SetValue(b.Value + delta)

	becameEmpty = b.Value <= 0
	becameFull = b.Value >= b.Max

	return becameEmpty, becameFull
}

// SetMax sets the maximum value and adjusts current value if needed
func (b *TokenBar) SetMax(max int) error {
	if max < 0 {
		return NewValidationError("token_bar.max", "cannot be negative")
	}

	// Scale current value proportionally if max changes
	if b.Max > 0 {
		b.Value = (b.Value * max) / b.Max
	}

	b.Max = max

	// Ensure value is within bounds
	if b.Value > b.Max {
		b.Value = b.Max
	}

	return nil
}

// GetColor returns the bar's color, or a default based on fill percentage
func (b *TokenBar) GetColor() string {
	if b.Color != "" {
		return b.Color
	}

	// Default color based on fill percentage
	pct := b.GetPercentage()
	if pct > 50 {
		return "#00ff00" // Green
	} else if pct > 25 {
		return "#ffff00" // Yellow
	}
	return "#ff0000" // Red
}

// GetLabel returns the bar's label
func (b *TokenBar) GetLabel() string {
	if b.Label != "" {
		return b.Label
	}

	switch b.Attribute {
	case AttributeHP:
		return "HP"
	case AttributeTempHP:
		return "Temp HP"
	case AttributeAC:
		return "AC"
	case AttributeResource:
		if b.CustomAttribute != "" {
			return b.CustomAttribute
		}
		return "Resource"
	default:
		return ""
	}
}

// Clone creates a deep copy of the token bar
func (b *TokenBar) Clone() *TokenBar {
	return &TokenBar{
		ID:              b.ID,
		Attribute:       b.Attribute,
		Value:           b.Value,
		Max:             b.Max,
		CustomAttribute: b.CustomAttribute,
		Visible:         b.Visible,
		Color:           b.Color,
		Label:           b.Label,
	}
}

// TokenBars represents a collection of token bars
type TokenBars []*TokenBar

// Get retrieves a bar by ID
func (bs TokenBars) Get(id BarID) *TokenBar {
	for _, bar := range bs {
		if bar.ID == id {
			return bar
		}
	}
	return nil
}

// Set adds or updates a bar
func (bs *TokenBars) Set(bar *TokenBar) error {
	if err := bar.Validate(); err != nil {
		return err
	}

	for i, b := range *bs {
		if b.ID == bar.ID {
			(*bs)[i] = bar
			return nil
		}
	}

	*bs = append(*bs, bar)
	return nil
}

// Validate validates all bars
func (bs TokenBars) Validate() error {
	for i, bar := range bs {
		if err := bar.Validate(); err != nil {
			return fmt.Errorf("bar at index %d: %w", i, err)
		}
	}
	return nil
}

// HasBar returns true if a bar with the given ID exists
func (bs TokenBars) HasBar(id BarID) bool {
	return bs.Get(id) != nil
}

// Clone creates a deep copy of all bars
func (bs TokenBars) Clone() TokenBars {
	result := make(TokenBars, len(bs))
	for i, bar := range bs {
		result[i] = bar.Clone()
	}
	return result
}
