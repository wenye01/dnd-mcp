// Package postgres 提供 PostgreSQL 会话存储实现
// 实现 persistence 包中定义的接口（接口在使用方定义，实现方实现）
package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/dnd-mcp/client/internal/models"
	"github.com/dnd-mcp/client/internal/persistence"
)

// PostgresSessionStore PostgreSQL 会话存储
type PostgresSessionStore struct {
	client *Client
}

// 确保 PostgresSessionStore 实现了 SessionWriter 和 SessionReader 接口
var _ persistence.SessionWriter = (*PostgresSessionStore)(nil)
var _ persistence.SessionReader = (*PostgresSessionStore)(nil)

// NewPostgresSessionStore 创建 PostgreSQL 会话存储
func NewPostgresSessionStore(client *Client) *PostgresSessionStore {
	return &PostgresSessionStore{client: client}
}

// Create 创建会话（实现 persistence.SessionWriter 接口）
// 使用 UPSERT: INSERT ... ON CONFLICT DO UPDATE
func (s *PostgresSessionStore) Create(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO client_sessions (
			id, created_at, updated_at, deleted_at,
			name, creator_id, mcp_server_url, websocket_key,
			max_players, settings, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
		ON CONFLICT (id) DO UPDATE SET
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at,
			name = EXCLUDED.name,
			max_players = EXCLUDED.max_players,
			settings = EXCLUDED.settings,
			status = EXCLUDED.status
	`

	// 序列化 settings
	settingsJSON, err := json.Marshal(session.Settings)
	if err != nil {
		return fmt.Errorf("序列化 settings 失败: %w", err)
	}

	// 处理 deleted_at
	deletedAt := interface{}(nil)
	if !session.DeletedAt.IsZero() {
		deletedAt = session.DeletedAt
	}

	_, err = s.client.Pool().Exec(ctx, query,
		session.ID,
		session.CreatedAt,
		session.UpdatedAt,
		deletedAt,
		session.Name,
		session.CreatorID,
		session.MCPServerURL,
		session.WebSocketKey,
		session.MaxPlayers,
		settingsJSON,
		session.Status,
	)

	if err != nil {
		return fmt.Errorf("创建会话失败: %w", err)
	}

	return nil
}

// BatchCreate 批量创建会话
func (s *PostgresSessionStore) BatchCreate(ctx context.Context, sessions []*models.Session) error {
	if len(sessions) == 0 {
		return nil
	}

	// 使用批量插入（使用 COPY 或批量 INSERT）
	// 这里使用批量 INSERT，因为 pgx 的 COPY 需要更多配置
	batch := &pgx.Batch{}

	for _, session := range sessions {
		query := `
			INSERT INTO client_sessions (
				id, created_at, updated_at, deleted_at,
				name, creator_id, mcp_server_url, websocket_key,
				max_players, settings, status
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
			)
			ON CONFLICT (id) DO UPDATE SET
				updated_at = EXCLUDED.updated_at,
				deleted_at = EXCLUDED.deleted_at,
				name = EXCLUDED.name,
				max_players = EXCLUDED.max_players,
				settings = EXCLUDED.settings,
				status = EXCLUDED.status
		`

		settingsJSON, err := json.Marshal(session.Settings)
		if err != nil {
			return fmt.Errorf("序列化 settings 失败: %w", err)
		}

		deletedAt := interface{}(nil)
		if !session.DeletedAt.IsZero() {
			deletedAt = session.DeletedAt
		}

		batch.Queue(query,
			session.ID,
			session.CreatedAt,
			session.UpdatedAt,
			deletedAt,
			session.Name,
			session.CreatorID,
			session.MCPServerURL,
			session.WebSocketKey,
			session.MaxPlayers,
			settingsJSON,
			session.Status,
		)
	}

	// 执行批量操作
	results := s.client.Pool().SendBatch(ctx, batch)
	defer results.Close()

	// 检查每个操作的结果
	for i := 0; i < len(sessions); i++ {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("批量创建会话失败 (第 %d 个): %w", i+1, err)
		}
	}

	return nil
}

// Update 更新会话
func (s *PostgresSessionStore) Update(ctx context.Context, session *models.Session) error {
	query := `
		UPDATE client_sessions SET
			updated_at = $2,
			deleted_at = $3,
			name = $4,
			max_players = $5,
			settings = $6,
			status = $7
		WHERE id = $1
	`

	settingsJSON, err := json.Marshal(session.Settings)
	if err != nil {
		return fmt.Errorf("序列化 settings 失败: %w", err)
	}

	deletedAt := interface{}(nil)
	if !session.DeletedAt.IsZero() {
		deletedAt = session.DeletedAt
	}

	result, err := s.client.Pool().Exec(ctx, query,
		session.ID,
		session.UpdatedAt,
		deletedAt,
		session.Name,
		session.MaxPlayers,
		settingsJSON,
		session.Status,
	)

	if err != nil {
		return fmt.Errorf("更新会话失败: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("会话不存在: %s", session.ID)
	}

	return nil
}

// Get 获取会话（实现 persistence.SessionReader 接口）
func (s *PostgresSessionStore) Get(ctx context.Context, id string) (*models.Session, error) {
	query := `
		SELECT id, created_at, updated_at, deleted_at,
		       name, creator_id, mcp_server_url, websocket_key,
		       max_players, settings, status
		FROM client_sessions
		WHERE id = $1
	`

	row := s.client.Pool().QueryRow(ctx, query, id)

	var session models.Session
	var deletedAt interface{}
	var settingsJSON []byte

	err := row.Scan(
		&session.ID,
		&session.CreatedAt,
		&session.UpdatedAt,
		&deletedAt,
		&session.Name,
		&session.CreatorID,
		&session.MCPServerURL,
		&session.WebSocketKey,
		&session.MaxPlayers,
		&settingsJSON,
		&session.Status,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("会话不存在: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("查询会话失败: %w", err)
	}

	// 处理 deleted_at
	if deletedAt != nil {
		session.DeletedAt = deletedAt.(time.Time)
	}

	// 反序列化 settings
	if len(settingsJSON) > 0 {
		if err := json.Unmarshal(settingsJSON, &session.Settings); err != nil {
			return nil, fmt.Errorf("反序列化 settings 失败: %w", err)
		}
	} else {
		session.Settings = make(map[string]interface{})
	}

	return &session, nil
}

// List 列出所有会话
func (s *PostgresSessionStore) List(ctx context.Context) ([]*models.Session, error) {
	query := `
		SELECT id, created_at, updated_at, deleted_at,
		       name, creator_id, mcp_server_url, websocket_key,
		       max_players, settings, status
		FROM client_sessions
		ORDER BY created_at DESC
	`

	rows, err := s.client.Pool().Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询会话列表失败: %w", err)
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		var session models.Session
		var deletedAt interface{}
		var settingsJSON []byte

		if err := rows.Scan(
			&session.ID,
			&session.CreatedAt,
			&session.UpdatedAt,
			&deletedAt,
			&session.Name,
			&session.CreatorID,
			&session.MCPServerURL,
			&session.WebSocketKey,
			&session.MaxPlayers,
			&settingsJSON,
			&session.Status,
		); err != nil {
			return nil, fmt.Errorf("扫描会话行失败: %w", err)
		}

		// 处理 deleted_at
		if deletedAt != nil {
			session.DeletedAt = deletedAt.(time.Time)
		}

		// 反序列化 settings
		if len(settingsJSON) > 0 {
			if err := json.Unmarshal(settingsJSON, &session.Settings); err != nil {
				return nil, fmt.Errorf("反序列化 settings 失败: %w", err)
			}
		} else {
			session.Settings = make(map[string]interface{})
		}

		sessions = append(sessions, &session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历会话行失败: %w", err)
	}

	return sessions, nil
}

// ListActive 列出活跃会话（软删除过滤）
func (s *PostgresSessionStore) ListActive(ctx context.Context) ([]*models.Session, error) {
	query := `
		SELECT id, created_at, updated_at, deleted_at,
		       name, creator_id, mcp_server_url, websocket_key,
		       max_players, settings, status
		FROM client_sessions
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := s.client.Pool().Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询活跃会话列表失败: %w", err)
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		var session models.Session
		var deletedAt interface{}
		var settingsJSON []byte

		if err := rows.Scan(
			&session.ID,
			&session.CreatedAt,
			&session.UpdatedAt,
			&deletedAt,
			&session.Name,
			&session.CreatorID,
			&session.MCPServerURL,
			&session.WebSocketKey,
			&session.MaxPlayers,
			&settingsJSON,
			&session.Status,
		); err != nil {
			return nil, fmt.Errorf("扫描会话行失败: %w", err)
		}

		// 反序列化 settings
		if len(settingsJSON) > 0 {
			if err := json.Unmarshal(settingsJSON, &session.Settings); err != nil {
				return nil, fmt.Errorf("反序列化 settings 失败: %w", err)
			}
		} else {
			session.Settings = make(map[string]interface{})
		}

		sessions = append(sessions, &session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历会话行失败: %w", err)
	}

	return sessions, nil
}
