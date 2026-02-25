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

	// Marshal core JSONB fields
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

	// Marshal extended JSONB fields
	speedDetailJSON, err := marshalOptionalJSON(character.SpeedDetail)
	if err != nil {
		return fmt.Errorf("failed to marshal speed_detail: %w", err)
	}

	deathSavesJSON, err := marshalOptionalJSON(character.DeathSaves)
	if err != nil {
		return fmt.Errorf("failed to marshal death_saves: %w", err)
	}

	skillsDetailJSON, err := marshalOptionalJSON(character.SkillsDetail)
	if err != nil {
		return fmt.Errorf("failed to marshal skills_detail: %w", err)
	}

	savesDetailJSON, err := marshalOptionalJSON(character.SavesDetail)
	if err != nil {
		return fmt.Errorf("failed to marshal saves_detail: %w", err)
	}

	currencyJSON, err := marshalOptionalJSON(character.Currency)
	if err != nil {
		return fmt.Errorf("failed to marshal currency: %w", err)
	}

	equipmentSlotsJSON, err := marshalOptionalJSON(character.EquipmentSlots)
	if err != nil {
		return fmt.Errorf("failed to marshal equipment_slots: %w", err)
	}

	inventoryItemsJSON, err := marshalOptionalJSON(character.InventoryItems)
	if err != nil {
		return fmt.Errorf("failed to marshal inventory_items: %w", err)
	}

	spellbookJSON, err := marshalOptionalJSON(character.Spellbook)
	if err != nil {
		return fmt.Errorf("failed to marshal spellbook: %w", err)
	}

	featuresJSON, err := marshalOptionalJSON(character.Features)
	if err != nil {
		return fmt.Errorf("failed to marshal features: %w", err)
	}

	biographyJSON, err := marshalOptionalJSON(character.Biography)
	if err != nil {
		return fmt.Errorf("failed to marshal biography: %w", err)
	}

	traitsJSON, err := marshalOptionalJSON(character.Traits)
	if err != nil {
		return fmt.Errorf("failed to marshal traits: %w", err)
	}

	importMetaJSON, err := marshalOptionalJSON(character.ImportMeta)
	if err != nil {
		return fmt.Errorf("failed to marshal import_meta: %w", err)
	}

	query := `
		INSERT INTO characters (
			id, campaign_id, name, is_npc, npc_type, player_id,
			race, class, level, background, alignment,
			abilities, hp, ac, speed, initiative,
			skills, saves, equipment, inventory, conditions,
			image, experience, proficiency, speed_detail, death_saves,
			skills_detail, saves_detail, currency, equipment_slots, inventory_items,
			spellbook, features, biography, traits, import_meta,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21,
		        $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38)
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
		nullString(character.Image),
		character.Experience,
		character.Proficiency,
		speedDetailJSON,
		deathSavesJSON,
		skillsDetailJSON,
		savesDetailJSON,
		currencyJSON,
		equipmentSlotsJSON,
		inventoryItemsJSON,
		spellbookJSON,
		featuresJSON,
		biographyJSON,
		traitsJSON,
		importMetaJSON,
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
			COALESCE(image, ''), COALESCE(experience, 0), COALESCE(proficiency, 0),
			speed_detail, death_saves, skills_detail, saves_detail,
			currency, equipment_slots, inventory_items, spellbook,
			features, biography, traits, import_meta,
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
			COALESCE(image, ''), COALESCE(experience, 0), COALESCE(proficiency, 0),
			speed_detail, death_saves, skills_detail, saves_detail,
			currency, equipment_slots, inventory_items, spellbook,
			features, biography, traits, import_meta,
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
			COALESCE(image, ''), COALESCE(experience, 0), COALESCE(proficiency, 0),
			speed_detail, death_saves, skills_detail, saves_detail,
			currency, equipment_slots, inventory_items, spellbook,
			features, biography, traits, import_meta,
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
		character, err := scanCharacterFromRows(rows)
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

	// Marshal core JSONB fields
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

	// Marshal extended JSONB fields
	speedDetailJSON, err := marshalOptionalJSON(character.SpeedDetail)
	if err != nil {
		return fmt.Errorf("failed to marshal speed_detail: %w", err)
	}

	deathSavesJSON, err := marshalOptionalJSON(character.DeathSaves)
	if err != nil {
		return fmt.Errorf("failed to marshal death_saves: %w", err)
	}

	skillsDetailJSON, err := marshalOptionalJSON(character.SkillsDetail)
	if err != nil {
		return fmt.Errorf("failed to marshal skills_detail: %w", err)
	}

	savesDetailJSON, err := marshalOptionalJSON(character.SavesDetail)
	if err != nil {
		return fmt.Errorf("failed to marshal saves_detail: %w", err)
	}

	currencyJSON, err := marshalOptionalJSON(character.Currency)
	if err != nil {
		return fmt.Errorf("failed to marshal currency: %w", err)
	}

	equipmentSlotsJSON, err := marshalOptionalJSON(character.EquipmentSlots)
	if err != nil {
		return fmt.Errorf("failed to marshal equipment_slots: %w", err)
	}

	inventoryItemsJSON, err := marshalOptionalJSON(character.InventoryItems)
	if err != nil {
		return fmt.Errorf("failed to marshal inventory_items: %w", err)
	}

	spellbookJSON, err := marshalOptionalJSON(character.Spellbook)
	if err != nil {
		return fmt.Errorf("failed to marshal spellbook: %w", err)
	}

	featuresJSON, err := marshalOptionalJSON(character.Features)
	if err != nil {
		return fmt.Errorf("failed to marshal features: %w", err)
	}

	biographyJSON, err := marshalOptionalJSON(character.Biography)
	if err != nil {
		return fmt.Errorf("failed to marshal biography: %w", err)
	}

	traitsJSON, err := marshalOptionalJSON(character.Traits)
	if err != nil {
		return fmt.Errorf("failed to marshal traits: %w", err)
	}

	importMetaJSON, err := marshalOptionalJSON(character.ImportMeta)
	if err != nil {
		return fmt.Errorf("failed to marshal import_meta: %w", err)
	}

	query := `
		UPDATE characters
		SET name = $1, is_npc = $2, npc_type = $3, player_id = $4,
			race = $5, class = $6, level = $7, background = $8, alignment = $9,
			abilities = $10, hp = $11, ac = $12, speed = $13, initiative = $14,
			skills = $15, saves = $16, equipment = $17, inventory = $18, conditions = $19,
			image = $20, experience = $21, proficiency = $22,
			speed_detail = $23, death_saves = $24, skills_detail = $25, saves_detail = $26,
			currency = $27, equipment_slots = $28, inventory_items = $29, spellbook = $30,
			features = $31, biography = $32, traits = $33, import_meta = $34,
			updated_at = $35
		WHERE id = $36
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
		nullString(character.Image),
		character.Experience,
		character.Proficiency,
		speedDetailJSON,
		deathSavesJSON,
		skillsDetailJSON,
		savesDetailJSON,
		currencyJSON,
		equipmentSlotsJSON,
		inventoryItemsJSON,
		spellbookJSON,
		featuresJSON,
		biographyJSON,
		traitsJSON,
		importMetaJSON,
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

// marshalOptionalJSON marshals an optional value to JSON, returning nil for nil values
func marshalOptionalJSON(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}

// scanCharacterFromRow scans a character from a single row
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
		image        sql.NullString
		experience   int
		proficiency  int
		speedDetailJSON []byte
		deathSavesJSON []byte
		skillsDetailJSON []byte
		savesDetailJSON []byte
		currencyJSON []byte
		equipmentSlotsJSON []byte
		inventoryItemsJSON []byte
		spellbookJSON []byte
		featuresJSON []byte
		biographyJSON []byte
		traitsJSON   []byte
		importMetaJSON []byte
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
		&image,
		&experience,
		&proficiency,
		&speedDetailJSON,
		&deathSavesJSON,
		&skillsDetailJSON,
		&savesDetailJSON,
		&currencyJSON,
		&equipmentSlotsJSON,
		&inventoryItemsJSON,
		&spellbookJSON,
		&featuresJSON,
		&biographyJSON,
		&traitsJSON,
		&importMetaJSON,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan character: %w", err)
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
		AC:          ac,
		Speed:       speed,
		Initiative:  initiative,
		Image:       image.String,
		Experience:  experience,
		Proficiency: proficiency,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	// Unmarshal core JSONB fields
	if len(abilitiesJSON) > 0 {
		var abilities models.Abilities
		if err := json.Unmarshal(abilitiesJSON, &abilities); err != nil {
			return nil, fmt.Errorf("failed to unmarshal abilities: %w", err)
		}
		character.Abilities = &abilities
	}

	if len(hpJSON) > 0 {
		var hp models.HP
		if err := json.Unmarshal(hpJSON, &hp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal hp: %w", err)
		}
		character.HP = &hp
	}

	if len(skillsJSON) > 0 {
		if err := json.Unmarshal(skillsJSON, &character.Skills); err != nil {
			return nil, fmt.Errorf("failed to unmarshal skills: %w", err)
		}
	}

	if len(savesJSON) > 0 {
		if err := json.Unmarshal(savesJSON, &character.Saves); err != nil {
			return nil, fmt.Errorf("failed to unmarshal saves: %w", err)
		}
	}

	if len(equipmentJSON) > 0 {
		if err := json.Unmarshal(equipmentJSON, &character.Equipment); err != nil {
			return nil, fmt.Errorf("failed to unmarshal equipment: %w", err)
		}
	}

	if len(inventoryJSON) > 0 {
		if err := json.Unmarshal(inventoryJSON, &character.Inventory); err != nil {
			return nil, fmt.Errorf("failed to unmarshal inventory: %w", err)
		}
	}

	if len(conditionsJSON) > 0 {
		if err := json.Unmarshal(conditionsJSON, &character.Conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}
	}

	// Unmarshal extended JSONB fields
	if len(speedDetailJSON) > 0 {
		var speedDetail models.Speed
		if err := json.Unmarshal(speedDetailJSON, &speedDetail); err != nil {
			return nil, fmt.Errorf("failed to unmarshal speed_detail: %w", err)
		}
		character.SpeedDetail = &speedDetail
	}

	if len(deathSavesJSON) > 0 {
		var deathSaves models.DeathSaves
		if err := json.Unmarshal(deathSavesJSON, &deathSaves); err != nil {
			return nil, fmt.Errorf("failed to unmarshal death_saves: %w", err)
		}
		character.DeathSaves = &deathSaves
	}

	if len(skillsDetailJSON) > 0 {
		if err := json.Unmarshal(skillsDetailJSON, &character.SkillsDetail); err != nil {
			return nil, fmt.Errorf("failed to unmarshal skills_detail: %w", err)
		}
	}

	if len(savesDetailJSON) > 0 {
		if err := json.Unmarshal(savesDetailJSON, &character.SavesDetail); err != nil {
			return nil, fmt.Errorf("failed to unmarshal saves_detail: %w", err)
		}
	}

	if len(currencyJSON) > 0 {
		var currency models.Currency
		if err := json.Unmarshal(currencyJSON, &currency); err != nil {
			return nil, fmt.Errorf("failed to unmarshal currency: %w", err)
		}
		character.Currency = &currency
	}

	if len(equipmentSlotsJSON) > 0 {
		var equipmentSlots models.EquipmentSlots
		if err := json.Unmarshal(equipmentSlotsJSON, &equipmentSlots); err != nil {
			return nil, fmt.Errorf("failed to unmarshal equipment_slots: %w", err)
		}
		character.EquipmentSlots = &equipmentSlots
	}

	if len(inventoryItemsJSON) > 0 {
		if err := json.Unmarshal(inventoryItemsJSON, &character.InventoryItems); err != nil {
			return nil, fmt.Errorf("failed to unmarshal inventory_items: %w", err)
		}
	}

	if len(spellbookJSON) > 0 {
		var spellbook models.Spellbook
		if err := json.Unmarshal(spellbookJSON, &spellbook); err != nil {
			return nil, fmt.Errorf("failed to unmarshal spellbook: %w", err)
		}
		character.Spellbook = &spellbook
	}

	if len(featuresJSON) > 0 {
		if err := json.Unmarshal(featuresJSON, &character.Features); err != nil {
			return nil, fmt.Errorf("failed to unmarshal features: %w", err)
		}
	}

	if len(biographyJSON) > 0 {
		var biography models.Biography
		if err := json.Unmarshal(biographyJSON, &biography); err != nil {
			return nil, fmt.Errorf("failed to unmarshal biography: %w", err)
		}
		character.Biography = &biography
	}

	if len(traitsJSON) > 0 {
		var traits models.Traits
		if err := json.Unmarshal(traitsJSON, &traits); err != nil {
			return nil, fmt.Errorf("failed to unmarshal traits: %w", err)
		}
		character.Traits = &traits
	}

	if len(importMetaJSON) > 0 {
		var importMeta models.ImportMeta
		if err := json.Unmarshal(importMetaJSON, &importMeta); err != nil {
			return nil, fmt.Errorf("failed to unmarshal import_meta: %w", err)
		}
		character.ImportMeta = &importMeta
	}

	return character, nil
}

// scanCharacterFromRows scans a character from rows
func scanCharacterFromRows(rows pgx.Rows) (*models.Character, error) {
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
		image        sql.NullString
		experience   int
		proficiency  int
		speedDetailJSON []byte
		deathSavesJSON []byte
		skillsDetailJSON []byte
		savesDetailJSON []byte
		currencyJSON []byte
		equipmentSlotsJSON []byte
		inventoryItemsJSON []byte
		spellbookJSON []byte
		featuresJSON []byte
		biographyJSON []byte
		traitsJSON   []byte
		importMetaJSON []byte
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := rows.Scan(
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
		&image,
		&experience,
		&proficiency,
		&speedDetailJSON,
		&deathSavesJSON,
		&skillsDetailJSON,
		&savesDetailJSON,
		&currencyJSON,
		&equipmentSlotsJSON,
		&inventoryItemsJSON,
		&spellbookJSON,
		&featuresJSON,
		&biographyJSON,
		&traitsJSON,
		&importMetaJSON,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan character: %w", err)
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
		AC:          ac,
		Speed:       speed,
		Initiative:  initiative,
		Image:       image.String,
		Experience:  experience,
		Proficiency: proficiency,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	// Unmarshal core JSONB fields
	if len(abilitiesJSON) > 0 {
		var abilities models.Abilities
		if err := json.Unmarshal(abilitiesJSON, &abilities); err != nil {
			return nil, fmt.Errorf("failed to unmarshal abilities: %w", err)
		}
		character.Abilities = &abilities
	}

	if len(hpJSON) > 0 {
		var hp models.HP
		if err := json.Unmarshal(hpJSON, &hp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal hp: %w", err)
		}
		character.HP = &hp
	}

	if len(skillsJSON) > 0 {
		if err := json.Unmarshal(skillsJSON, &character.Skills); err != nil {
			return nil, fmt.Errorf("failed to unmarshal skills: %w", err)
		}
	}

	if len(savesJSON) > 0 {
		if err := json.Unmarshal(savesJSON, &character.Saves); err != nil {
			return nil, fmt.Errorf("failed to unmarshal saves: %w", err)
		}
	}

	if len(equipmentJSON) > 0 {
		if err := json.Unmarshal(equipmentJSON, &character.Equipment); err != nil {
			return nil, fmt.Errorf("failed to unmarshal equipment: %w", err)
		}
	}

	if len(inventoryJSON) > 0 {
		if err := json.Unmarshal(inventoryJSON, &character.Inventory); err != nil {
			return nil, fmt.Errorf("failed to unmarshal inventory: %w", err)
		}
	}

	if len(conditionsJSON) > 0 {
		if err := json.Unmarshal(conditionsJSON, &character.Conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}
	}

	// Unmarshal extended JSONB fields
	if len(speedDetailJSON) > 0 {
		var speedDetail models.Speed
		if err := json.Unmarshal(speedDetailJSON, &speedDetail); err != nil {
			return nil, fmt.Errorf("failed to unmarshal speed_detail: %w", err)
		}
		character.SpeedDetail = &speedDetail
	}

	if len(deathSavesJSON) > 0 {
		var deathSaves models.DeathSaves
		if err := json.Unmarshal(deathSavesJSON, &deathSaves); err != nil {
			return nil, fmt.Errorf("failed to unmarshal death_saves: %w", err)
		}
		character.DeathSaves = &deathSaves
	}

	if len(skillsDetailJSON) > 0 {
		if err := json.Unmarshal(skillsDetailJSON, &character.SkillsDetail); err != nil {
			return nil, fmt.Errorf("failed to unmarshal skills_detail: %w", err)
		}
	}

	if len(savesDetailJSON) > 0 {
		if err := json.Unmarshal(savesDetailJSON, &character.SavesDetail); err != nil {
			return nil, fmt.Errorf("failed to unmarshal saves_detail: %w", err)
		}
	}

	if len(currencyJSON) > 0 {
		var currency models.Currency
		if err := json.Unmarshal(currencyJSON, &currency); err != nil {
			return nil, fmt.Errorf("failed to unmarshal currency: %w", err)
		}
		character.Currency = &currency
	}

	if len(equipmentSlotsJSON) > 0 {
		var equipmentSlots models.EquipmentSlots
		if err := json.Unmarshal(equipmentSlotsJSON, &equipmentSlots); err != nil {
			return nil, fmt.Errorf("failed to unmarshal equipment_slots: %w", err)
		}
		character.EquipmentSlots = &equipmentSlots
	}

	if len(inventoryItemsJSON) > 0 {
		if err := json.Unmarshal(inventoryItemsJSON, &character.InventoryItems); err != nil {
			return nil, fmt.Errorf("failed to unmarshal inventory_items: %w", err)
		}
	}

	if len(spellbookJSON) > 0 {
		var spellbook models.Spellbook
		if err := json.Unmarshal(spellbookJSON, &spellbook); err != nil {
			return nil, fmt.Errorf("failed to unmarshal spellbook: %w", err)
		}
		character.Spellbook = &spellbook
	}

	if len(featuresJSON) > 0 {
		if err := json.Unmarshal(featuresJSON, &character.Features); err != nil {
			return nil, fmt.Errorf("failed to unmarshal features: %w", err)
		}
	}

	if len(biographyJSON) > 0 {
		var biography models.Biography
		if err := json.Unmarshal(biographyJSON, &biography); err != nil {
			return nil, fmt.Errorf("failed to unmarshal biography: %w", err)
		}
		character.Biography = &biography
	}

	if len(traitsJSON) > 0 {
		var traits models.Traits
		if err := json.Unmarshal(traitsJSON, &traits); err != nil {
			return nil, fmt.Errorf("failed to unmarshal traits: %w", err)
		}
		character.Traits = &traits
	}

	if len(importMetaJSON) > 0 {
		var importMeta models.ImportMeta
		if err := json.Unmarshal(importMetaJSON, &importMeta); err != nil {
			return nil, fmt.Errorf("failed to unmarshal import_meta: %w", err)
		}
		character.ImportMeta = &importMeta
	}

	return character, nil
}
