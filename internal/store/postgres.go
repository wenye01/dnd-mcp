// Package store 提供PostgreSQL数据持久化实现
package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dnd-mcp/client/internal/models"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// PostgresStore PostgreSQL存储实现
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore 创建PostgreSQL存储
func NewPostgresStore(dataSourceName string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 验证连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStore{db: db}, nil
}

// Close 关闭数据库连接
func (s *PostgresStore) Close() error {
	return s.db.Close()
}

// CreateSession 创建新会话
func (s *PostgresStore) CreateSession(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO sessions (id, version, created_at, updated_at, game_time, location, campaign_name, state)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	stateJSON, err := json.Marshal(session.State)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	_, err = s.db.ExecContext(
		ctx,
		query,
		session.ID,
		session.Version,
		session.CreatedAt,
		session.UpdatedAt,
		session.GameTime,
		session.Location,
		session.CampaignName,
		stateJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetSession 根据ID获取会话
func (s *PostgresStore) GetSession(ctx context.Context, sessionID uuid.UUID) (*models.Session, error) {
	query := `
		SELECT id, version, created_at, updated_at, game_time, location, campaign_name, state
		FROM sessions
		WHERE id = $1 AND deleted_at IS NULL
	`

	var session models.Session
	var stateJSON []byte

	err := s.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID,
		&session.Version,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.GameTime,
		&session.Location,
		&session.CampaignName,
		&stateJSON,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// 解析state JSON
	if len(stateJSON) > 0 {
		if err := json.Unmarshal(stateJSON, &session.State); err != nil {
			return nil, fmt.Errorf("failed to unmarshal state: %w", err)
		}
	}

	return &session, nil
}

// ListSessions 列出会话
func (s *PostgresStore) ListSessions(ctx context.Context, limit, offset int) ([]*models.Session, error) {
	query := `
		SELECT id, version, created_at, updated_at, game_time, location, campaign_name, state
		FROM sessions
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	sessions := make([]*models.Session, 0)
	for rows.Next() {
		var session models.Session
		var stateJSON []byte

		err := rows.Scan(
			&session.ID,
			&session.Version,
			&session.CreatedAt,
			&session.UpdatedAt,
			&session.GameTime,
			&session.Location,
			&session.CampaignName,
			&stateJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		// 解析state JSON
		if len(stateJSON) > 0 {
			if err := json.Unmarshal(stateJSON, &session.State); err != nil {
				return nil, fmt.Errorf("failed to unmarshal state: %w", err)
			}
		}

		sessions = append(sessions, &session)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// UpdateSession 更新会话
func (s *PostgresStore) UpdateSession(ctx context.Context, session *models.Session) error {
	query := `
		UPDATE sessions
		SET version = $2, updated_at = $3, game_time = $4, location = $5, campaign_name = $6, state = $7
		WHERE id = $1
	`

	stateJSON, err := json.Marshal(session.State)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	session.UpdatedAt = time.Now()

	result, err := s.db.ExecContext(
		ctx,
		query,
		session.ID,
		session.Version,
		session.UpdatedAt,
		session.GameTime,
		session.Location,
		session.CampaignName,
		stateJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// DeleteSession 删除会话(软删除)
func (s *PostgresStore) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	query := `
		UPDATE sessions
		SET deleted_at = NOW()
		WHERE id = $1
	`

	result, err := s.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// CreateMessage 创建新消息
func (s *PostgresStore) CreateMessage(ctx context.Context, message *models.Message) error {
	query := `
		INSERT INTO messages (id, session_id, role, content, tool_calls, player_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	toolCallsJSON, err := json.Marshal(message.ToolCalls)
	if err != nil {
		return fmt.Errorf("failed to marshal tool_calls: %w", err)
	}

	_, err = s.db.ExecContext(
		ctx,
		query,
		message.ID,
		message.SessionID,
		message.Role,
		message.Content,
		toolCallsJSON,
		message.PlayerID,
		message.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

// GetMessages 获取指定会话的消息列表
func (s *PostgresStore) GetMessages(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	query := `
		SELECT id, session_id, role, content, tool_calls, player_id, created_at
		FROM messages
		WHERE session_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, query, sessionID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	messages := make([]*models.Message, 0)
	for rows.Next() {
		var message models.Message
		var toolCallsJSON []byte

		err := rows.Scan(
			&message.ID,
			&message.SessionID,
			&message.Role,
			&message.Content,
			&toolCallsJSON,
			&message.PlayerID,
			&message.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		// 解析tool_calls JSON
		if len(toolCallsJSON) > 0 {
			if err := json.Unmarshal(toolCallsJSON, &message.ToolCalls); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tool_calls: %w", err)
			}
		}

		messages = append(messages, &message)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	return messages, nil
}

// GetRecentMessages 获取最近的消息
func (s *PostgresStore) GetRecentMessages(ctx context.Context, sessionID uuid.UUID, limit int) ([]*models.Message, error) {
	return s.GetMessages(ctx, sessionID, limit, 0)
}
