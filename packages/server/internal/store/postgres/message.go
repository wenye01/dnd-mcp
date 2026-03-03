package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MessageStore implements message storage using PostgreSQL
type MessageStore struct {
	pool *pgxpool.Pool
}

// NewMessageStore creates a new message store
func NewMessageStore(client *Client) *MessageStore {
	return &MessageStore{pool: client.Pool()}
}

// Create creates a new message
func (s *MessageStore) Create(ctx context.Context, message *models.Message) error {
	// Set timestamp if not already set
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}

	// Marshal tool_calls to JSONB
	var toolCallsJSON []byte
	var err error
	if len(message.ToolCalls) > 0 {
		toolCallsJSON, err = json.Marshal(message.ToolCalls)
		if err != nil {
			return fmt.Errorf("failed to marshal tool_calls: %w", err)
		}
	}

	query := `
		INSERT INTO messages (id, campaign_id, role, content, player_id, tool_calls, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = s.pool.Exec(ctx, query,
		message.ID,
		message.CampaignID,
		string(message.Role),
		message.Content,
		nullString(message.PlayerID),
		nullJSON(toolCallsJSON),
		message.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

// Get retrieves a message by ID
func (s *MessageStore) Get(ctx context.Context, id string) (*models.Message, error) {
	query := `
		SELECT id, campaign_id, role, content, player_id, tool_calls, created_at
		FROM messages
		WHERE id = $1
	`

	return s.scanMessage(ctx, query, id)
}

// GetByCampaignID retrieves a message by campaign ID and message ID
func (s *MessageStore) GetByCampaignID(ctx context.Context, campaignID, id string) (*models.Message, error) {
	query := `
		SELECT id, campaign_id, role, content, player_id, tool_calls, created_at
		FROM messages
		WHERE campaign_id = $1 AND id = $2
	`

	return s.scanMessage(ctx, query, campaignID, id)
}

// ListByCampaign retrieves messages for a campaign, ordered by created_at
func (s *MessageStore) ListByCampaign(ctx context.Context, campaignID string, limit int) ([]*models.Message, error) {
	query := `
		SELECT id, campaign_id, role, content, player_id, tool_calls, created_at
		FROM messages
		WHERE campaign_id = $1
		ORDER BY created_at ASC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.pool.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		msg, err := scanMessageFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	return messages, nil
}

// ListByCampaignWithOffset retrieves messages with pagination
func (s *MessageStore) ListByCampaignWithOffset(ctx context.Context, campaignID string, limit, offset int) ([]*models.Message, error) {
	query := `
		SELECT id, campaign_id, role, content, player_id, tool_calls, created_at
		FROM messages
		WHERE campaign_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.pool.Query(ctx, query, campaignID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		msg, err := scanMessageFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	return messages, nil
}

// CountByCampaign returns the number of messages in a campaign
func (s *MessageStore) CountByCampaign(ctx context.Context, campaignID string) (int, error) {
	query := `SELECT COUNT(*) FROM messages WHERE campaign_id = $1`

	var count int
	err := s.pool.QueryRow(ctx, query, campaignID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}

	return count, nil
}

// Delete deletes a message
func (s *MessageStore) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM messages WHERE id = $1`

	result, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteByCampaign deletes all messages for a campaign
func (s *MessageStore) DeleteByCampaign(ctx context.Context, campaignID string) error {
	query := `DELETE FROM messages WHERE campaign_id = $1`

	_, err := s.pool.Exec(ctx, query, campaignID)
	if err != nil {
		return fmt.Errorf("failed to delete campaign messages: %w", err)
	}

	return nil
}

// DeleteByCampaignBeforeDate deletes messages before a specific date
func (s *MessageStore) DeleteByCampaignBeforeDate(ctx context.Context, campaignID string, beforeDate time.Time) (int64, error) {
	query := `DELETE FROM messages WHERE campaign_id = $1 AND created_at < $2`

	result, err := s.pool.Exec(ctx, query, campaignID, beforeDate)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old messages: %w", err)
	}

	return result.RowsAffected(), nil
}

// scanMessage scans a single message using the provided query
func (s *MessageStore) scanMessage(ctx context.Context, query string, args ...interface{}) (*models.Message, error) {
	row := s.pool.QueryRow(ctx, query, args...)
	return scanMessageFromRow(row)
}

// scanMessageFromRow scans a message from a row
func scanMessageFromRow(row pgx.Row) (*models.Message, error) {
	var (
		id            string
		campaignID    string
		role          string
		content       string
		playerID      sql.NullString
		toolCallsJSON []byte
		createdAt     time.Time
	)

	err := row.Scan(
		&id,
		&campaignID,
		&role,
		&content,
		&playerID,
		&toolCallsJSON,
		&createdAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan message: %w", err)
	}

	// Unmarshal tool_calls
	var toolCalls []models.ToolCall
	if len(toolCallsJSON) > 0 {
		if err := json.Unmarshal(toolCallsJSON, &toolCalls); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tool_calls: %w", err)
		}
	}

	message := &models.Message{
		ID:         id,
		CampaignID: campaignID,
		Role:       models.MessageRole(role),
		Content:    content,
		PlayerID:   playerID.String,
		ToolCalls:  toolCalls,
		CreatedAt:  createdAt,
	}

	return message, nil
}
