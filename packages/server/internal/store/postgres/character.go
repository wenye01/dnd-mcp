package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/store"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CharacterStore implements store.CharacterStore using PostgreSQL
type CharacterStore struct {
	pool *pgxpool.Pool
}

// NewCharacterStore creates a new character store
func NewCharacterStore(client *Client) *CharacterStore {
	return &CharacterStore{pool: client.Pool()}
}

// Create creates a new character
func (s *CharacterStore) Create(ctx context.Context, character *models.Character) error {
	// Generate UUID if not set
	if character.ID == "" {
		character.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	if character.CreatedAt.IsZero() {
		character.CreatedAt = now
	}
	character.UpdatedAt = now

	// Marshal JSONB fields
	abilitiesJSON, err := json.Marshal(character.Abilities)
	if err != nil {
		return fmt.Errorf("failed to marshal abilities: %w", err)
	}

	hpJSON, err := json.Marshal(character.HP)
	if err != nil {
		return fmt.Errorf("failed to marshal hp: %w", err)
	}

	skillsJSON, err := json.Marshal(character.Skills)
	if err != nil {
		return fmt.Errorf("failed to marshal skills: %w", err)
	}

	savesJSON, err := json.Marshal(character.Saves)
	if err != nil {
		return fmt.Errorf("failed to marshal saves: %w", err)
	}

	equipmentJSON, err := json.Marshal(character.Equipment)
	if err != nil {
		return fmt.Errorf("failed to marshal equipment: %w", err)
	}

	inventoryJSON, err := json.Marshal(character.Inventory)
	if err != nil {
		return fmt.Errorf("failed to marshal inventory: %w", err)
	}

	conditionsJSON, err := json.Marshal(character.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	query := `
		INSERT INTO characters (
			id, campaign_id, name, is_npc, npc_type, player_id,
			race, class, level, background, alignment,
			abilities, hp, ac, speed, initiative,
			skills, saves, equipment, inventory, conditions,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)
	`

	_, err = s.pool.Exec(ctx, query,
		character.ID,
		character.CampaignID,
		character.Name,
		character.IsNPC,
		nullString(string(character.NPCType)),
		nullString(character.PlayerID),
		character.Race,
		character.Class,
		character.Level,
		nullString(character.Background),
		nullString(character.Alignment),
		abilitiesJSON,
		hpJSON,
		character.AC,
		character.Speed,
		character.Initiative,
		skillsJSON,
		savesJSON,
		equipmentJSON,
		inventoryJSON,
		conditionsJSON,
		character.CreatedAt,
		character.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create character: %w", err)
	}

	return nil
}

// Get retrieves a character by ID
func (s *CharacterStore) Get(ctx context.Context, id string) (*models.Character, error) {
	query := `
		SELECT id, campaign_id, name, is_npc, npc_type, player_id,
			race, class, level, background, alignment,
			abilities, hp, ac, speed, initiative,
			skills, saves, equipment, inventory, conditions,
			created_at, updated_at
		FROM characters
		WHERE id = $1
	`

	return s.scanCharacter(ctx, query, id)
}

// GetByCampaignAndID retrieves a character by campaign ID and character ID
func (s *CharacterStore) GetByCampaignAndID(ctx context.Context, campaignID, id string) (*models.Character, error) {
	query := `
		SELECT id, campaign_id, name, is_npc, npc_type, player_id,
			race, class, level, background, alignment,
			abilities, hp, ac, speed, initiative,
			skills, saves, equipment, inventory, conditions,
			created_at, updated_at
		FROM characters
		WHERE id = $1 AND campaign_id = $2
	`

	return s.scanCharacter(ctx, query, id, campaignID)
}

// List lists characters with optional filters
func (s *CharacterStore) List(ctx context.Context, filter *store.CharacterFilter) ([]*models.Character, error) {
	if filter == nil {
		filter = &store.CharacterFilter{}
	}

	// Build query
	query := `
		SELECT id, campaign_id, name, is_npc, npc_type, player_id,
			race, class, level, background, alignment,
			abilities, hp, ac, speed, initiative,
			skills, saves, equipment, inventory, conditions,
			created_at, updated_at
		FROM characters
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if filter.CampaignID != "" {
		query += fmt.Sprintf(" AND campaign_id = $%d", argIndex)
		args = append(args, filter.CampaignID)
		argIndex++
	}

	if filter.IsNPC != nil {
		query += fmt.Sprintf(" AND is_npc = $%d", argIndex)
		args = append(args, *filter.IsNPC)
		argIndex++
	}

	if filter.PlayerID != "" {
		query += fmt.Sprintf(" AND player_id = $%d", argIndex)
		args = append(args, filter.PlayerID)
		argIndex++
	}

	if filter.NPCType != "" {
		query += fmt.Sprintf(" AND npc_type = $%d", argIndex)
		args = append(args, string(filter.NPCType))
		argIndex++
	}

	// Order by created_at desc
	query += " ORDER BY created_at DESC"

	// Apply pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list characters: %w", err)
	}
	defer rows.Close()

	var characters []*models.Character
	for rows.Next() {
		character, err := scanCharacterFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan character: %w", err)
		}
		characters = append(characters, character)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating characters: %w", err)
	}

	return characters, nil
}

// Update updates a character
func (s *CharacterStore) Update(ctx context.Context, character *models.Character) error {
	character.UpdatedAt = time.Now()

	// Marshal JSONB fields
	abilitiesJSON, err := json.Marshal(character.Abilities)
	if err != nil {
		return fmt.Errorf("failed to marshal abilities: %w", err)
	}

	hpJSON, err := json.Marshal(character.HP)
	if err != nil {
		return fmt.Errorf("failed to marshal hp: %w", err)
	}

	skillsJSON, err := json.Marshal(character.Skills)
	if err != nil {
		return fmt.Errorf("failed to marshal skills: %w", err)
	}

	savesJSON, err := json.Marshal(character.Saves)
	if err != nil {
		return fmt.Errorf("failed to marshal saves: %w", err)
	}

	equipmentJSON, err := json.Marshal(character.Equipment)
	if err != nil {
		return fmt.Errorf("failed to marshal equipment: %w", err)
	}

	inventoryJSON, err := json.Marshal(character.Inventory)
	if err != nil {
		return fmt.Errorf("failed to marshal inventory: %w", err)
	}

	conditionsJSON, err := json.Marshal(character.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	query := `
		UPDATE characters
		SET name = $1, is_npc = $2, npc_type = $3, player_id = $4,
			race = $5, class = $6, level = $7, background = $8, alignment = $9,
			abilities = $10, hp = $11, ac = $12, speed = $13, initiative = $14,
			skills = $15, saves = $16, equipment = $17, inventory = $18, conditions = $19,
			updated_at = $20
		WHERE id = $21
	`

	result, err := s.pool.Exec(ctx, query,
		character.Name,
		character.IsNPC,
		nullString(string(character.NPCType)),
		nullString(character.PlayerID),
		character.Race,
		character.Class,
		character.Level,
		nullString(character.Background),
		nullString(character.Alignment),
		abilitiesJSON,
		hpJSON,
		character.AC,
		character.Speed,
		character.Initiative,
		skillsJSON,
		savesJSON,
		equipmentJSON,
		inventoryJSON,
		conditionsJSON,
		character.UpdatedAt,
		character.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update character: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// Delete deletes a character
func (s *CharacterStore) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM characters WHERE id = $1`

	result, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete character: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// Count counts characters with optional filter
func (s *CharacterStore) Count(ctx context.Context, filter *store.CharacterFilter) (int64, error) {
	if filter == nil {
		filter = &store.CharacterFilter{}
	}

	query := "SELECT COUNT(*) FROM characters WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if filter.CampaignID != "" {
		query += fmt.Sprintf(" AND campaign_id = $%d", argIndex)
		args = append(args, filter.CampaignID)
		argIndex++
	}

	if filter.IsNPC != nil {
		query += fmt.Sprintf(" AND is_npc = $%d", argIndex)
		args = append(args, *filter.IsNPC)
		argIndex++
	}

	if filter.PlayerID != "" {
		query += fmt.Sprintf(" AND player_id = $%d", argIndex)
		args = append(args, filter.PlayerID)
		argIndex++
	}

	if filter.NPCType != "" {
		query += fmt.Sprintf(" AND npc_type = $%d", argIndex)
		args = append(args, string(filter.NPCType))
		argIndex++
	}

	var count int64
	err := s.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count characters: %w", err)
	}

	return count, nil
}

// scanCharacter scans a single character using the provided query
func (s *CharacterStore) scanCharacter(ctx context.Context, query string, args ...interface{}) (*models.Character, error) {
	row := s.pool.QueryRow(ctx, query, args...)
	return scanCharacterFromRow(row)
}

// scanCharacterFromRow scans a character from a row
func scanCharacterFromRow(row pgx.Row) (*models.Character, error) {
	var (
		id           string
		campaignID   string
		name         string
		isNPC        bool
		npcType      sql.NullString
		playerID     sql.NullString
		race         string
		class        string
		level        int
		background   sql.NullString
		alignment    sql.NullString
		abilitiesJSON []byte
		hpJSON       []byte
		ac           int
		speed        int
		initiative   int
		skillsJSON   []byte
		savesJSON    []byte
		equipmentJSON []byte
		inventoryJSON []byte
		conditionsJSON []byte
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := row.Scan(
		&id,
		&campaignID,
		&name,
		&isNPC,
		&npcType,
		&playerID,
		&race,
		&class,
		&level,
		&background,
		&alignment,
		&abilitiesJSON,
		&hpJSON,
		&ac,
		&speed,
		&initiative,
		&skillsJSON,
		&savesJSON,
		&equipmentJSON,
		&inventoryJSON,
		&conditionsJSON,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan character: %w", err)
	}

	// Unmarshal JSONB fields
	var abilities models.Abilities
	if len(abilitiesJSON) > 0 {
		if err := json.Unmarshal(abilitiesJSON, &abilities); err != nil {
			return nil, fmt.Errorf("failed to unmarshal abilities: %w", err)
		}
	}

	var hp models.HP
	if len(hpJSON) > 0 {
		if err := json.Unmarshal(hpJSON, &hp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal hp: %w", err)
		}
	}

	var skills map[string]int
	if len(skillsJSON) > 0 {
		if err := json.Unmarshal(skillsJSON, &skills); err != nil {
			return nil, fmt.Errorf("failed to unmarshal skills: %w", err)
		}
	}

	var saves map[string]int
	if len(savesJSON) > 0 {
		if err := json.Unmarshal(savesJSON, &saves); err != nil {
			return nil, fmt.Errorf("failed to unmarshal saves: %w", err)
		}
	}

	var equipment []models.Equipment
	if len(equipmentJSON) > 0 {
		if err := json.Unmarshal(equipmentJSON, &equipment); err != nil {
			return nil, fmt.Errorf("failed to unmarshal equipment: %w", err)
		}
	}

	var inventory []models.Item
	if len(inventoryJSON) > 0 {
		if err := json.Unmarshal(inventoryJSON, &inventory); err != nil {
			return nil, fmt.Errorf("failed to unmarshal inventory: %w", err)
		}
	}

	var conditions []models.Condition
	if len(conditionsJSON) > 0 {
		if err := json.Unmarshal(conditionsJSON, &conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}
	}

	character := &models.Character{
		ID:          id,
		CampaignID:  campaignID,
		Name:        name,
		IsNPC:       isNPC,
		NPCType:     models.NPCType(npcType.String),
		PlayerID:    playerID.String,
		Race:        race,
		Class:       class,
		Level:       level,
		Background:  background.String,
		Alignment:   alignment.String,
		Abilities:   &abilities,
		HP:          &hp,
		AC:          ac,
		Speed:       speed,
		Initiative:  initiative,
		Skills:      skills,
		Saves:       saves,
		Equipment:   equipment,
		Inventory:   inventory,
		Conditions:  conditions,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	return character, nil
}
