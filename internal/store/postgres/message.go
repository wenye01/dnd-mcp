// Package postgres 提供 PostgreSQL 消息存储实现
// 实现 persistence 包中定义的接口（接口在使用方定义，实现方实现）
package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dnd-mcp/client/internal/models"
	"github.com/dnd-mcp/client/internal/persistence"
	"github.com/jackc/pgx/v5"
)

// PostgresMessageStore PostgreSQL 消息存储
type PostgresMessageStore struct {
	client *Client
}

// 确保 PostgresMessageStore 实现了 MessageWriter 和 MessageReader 接口
var _ persistence.MessageWriter = (*PostgresMessageStore)(nil)
var _ persistence.MessageReader = (*PostgresMessageStore)(nil)

// NewPostgresMessageStore 创建 PostgreSQL 消息存储
func NewPostgresMessageStore(client *Client) *PostgresMessageStore {
	return &PostgresMessageStore{client: client}
}

// Create 创建消息（实现 persistence.MessageWriter）
func (m *PostgresMessageStore) Create(ctx context.Context, message *models.Message) error {
	query := `
		INSERT INTO client_messages (
			id, session_id, created_at, role, content, tool_calls, player_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		ON CONFLICT (id) DO NOTHING
	`

	// 序列化 tool_calls
	var toolCallsJSON []byte
	if len(message.ToolCalls) > 0 {
		var err error
		toolCallsJSON, err = json.Marshal(message.ToolCalls)
		if err != nil {
			return fmt.Errorf("序列化 tool_calls 失败: %w", err)
		}
	}

	_, err := m.client.Pool().Exec(ctx, query,
		message.ID,
		message.SessionID,
		message.CreatedAt,
		message.Role,
		message.Content,
		toolCallsJSON,
		message.PlayerID,
	)

	if err != nil {
		return fmt.Errorf("创建消息失败: %w", err)
	}

	return nil
}

// BatchCreate 批量创建消息
// 使用 COPY 命令或批量 INSERT
func (m *PostgresMessageStore) BatchCreate(ctx context.Context, messages []*models.Message) error {
	if len(messages) == 0 {
		return nil
	}

	// 使用 pgx CopyFrom 进行高性能批量插入
	// CopyFrom 使用 COPY 命令，性能最佳
	tableName := pgx.Identifier{"client_messages"}
	columns := []string{"id", "session_id", "created_at", "role", "content", "tool_calls", "player_id"}

	// 准备数据
	rows := make([][]interface{}, len(messages))
	for i, msg := range messages {
		// 序列化 tool_calls
		var toolCallsJSON []byte
		if len(msg.ToolCalls) > 0 {
			var err error
			toolCallsJSON, err = json.Marshal(msg.ToolCalls)
			if err != nil {
				return fmt.Errorf("序列化 tool_calls 失败: %w", err)
			}
		}

		rows[i] = []interface{}{
			msg.ID,
			msg.SessionID,
			msg.CreatedAt,
			msg.Role,
			msg.Content,
			toolCallsJSON,
			msg.PlayerID,
		}
	}

	// 使用 CopyFrom
	// 需要实现 pgx.CopyFromSource 接口
	source := &messageCopySource{
		rows: rows,
		idx:  -1,
	}

	_, err := m.client.Pool().CopyFrom(
		ctx,
		tableName,
		columns,
		source,
	)

	if err != nil {
		return fmt.Errorf("批量创建消息失败: %w", err)
	}

	return nil
}

// Get 获取消息（实现 persistence.MessageReader）
func (m *PostgresMessageStore) Get(ctx context.Context, sessionID, messageID string) (*models.Message, error) {
	query := `
		SELECT id, session_id, created_at, role, content, tool_calls, player_id
		FROM client_messages
		WHERE session_id = $1 AND id = $2
	`

	row := m.client.Pool().QueryRow(ctx, query, sessionID, messageID)

	var message models.Message
	var toolCallsJSON []byte

	err := row.Scan(
		&message.ID,
		&message.SessionID,
		&message.CreatedAt,
		&message.Role,
		&message.Content,
		&toolCallsJSON,
		&message.PlayerID,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("消息不存在: %s", messageID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询消息失败: %w", err)
	}

	// 反序列化 tool_calls
	if len(toolCallsJSON) > 0 {
		if err := json.Unmarshal(toolCallsJSON, &message.ToolCalls); err != nil {
			return nil, fmt.Errorf("反序列化 tool_calls 失败: %w", err)
		}
	}

	return &message, nil
}

// List 获取消息列表（按时间升序，从旧到新）
func (m *PostgresMessageStore) List(ctx context.Context, sessionID string, limit int) ([]*models.Message, error) {
	query := `
		SELECT id, session_id, created_at, role, content, tool_calls, player_id
		FROM client_messages
		WHERE session_id = $1
		ORDER BY created_at ASC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := m.client.Pool().Query(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("查询消息列表失败: %w", err)
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var message models.Message
		var toolCallsJSON []byte

		if err := rows.Scan(
			&message.ID,
			&message.SessionID,
			&message.CreatedAt,
			&message.Role,
			&message.Content,
			&toolCallsJSON,
			&message.PlayerID,
		); err != nil {
			return nil, fmt.Errorf("扫描消息行失败: %w", err)
		}

		// 反序列化 tool_calls
		if len(toolCallsJSON) > 0 {
			if err := json.Unmarshal(toolCallsJSON, &message.ToolCalls); err != nil {
				return nil, fmt.Errorf("反序列化 tool_calls 失败: %w", err)
			}
		}

		messages = append(messages, &message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历消息行失败: %w", err)
	}

	return messages, nil
}

// ListByRole 按角色获取消息
func (m *PostgresMessageStore) ListByRole(ctx context.Context, sessionID, role string, limit int) ([]*models.Message, error) {
	query := `
		SELECT id, session_id, created_at, role, content, tool_calls, player_id
		FROM client_messages
		WHERE session_id = $1 AND role = $2
		ORDER BY created_at ASC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := m.client.Pool().Query(ctx, query, sessionID, role)
	if err != nil {
		return nil, fmt.Errorf("查询消息列表失败: %w", err)
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var message models.Message
		var toolCallsJSON []byte

		if err := rows.Scan(
			&message.ID,
			&message.SessionID,
			&message.CreatedAt,
			&message.Role,
			&message.Content,
			&toolCallsJSON,
			&message.PlayerID,
		); err != nil {
			return nil, fmt.Errorf("扫描消息行失败: %w", err)
		}

		// 反序列化 tool_calls
		if len(toolCallsJSON) > 0 {
			if err := json.Unmarshal(toolCallsJSON, &message.ToolCalls); err != nil {
				return nil, fmt.Errorf("反序列化 tool_calls 失败: %w", err)
			}
		}

		messages = append(messages, &message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历消息行失败: %w", err)
	}

	return messages, nil
}

// messageCopySource 实现 pgx.CopyFromSource 接口
type messageCopySource struct {
	rows [][]interface{}
	idx  int
}

// Next 返回下一行数据
func (s *messageCopySource) Next() bool {
	s.idx++
	return s.idx < len(s.rows)
}

// Values 返回当前行的值
func (s *messageCopySource) Values() ([]interface{}, error) {
	if s.idx < 0 || s.idx >= len(s.rows) {
		return nil, fmt.Errorf("索引越界: %d", s.idx)
	}
	return s.rows[s.idx], nil
}

// Err 返回错误
func (s *messageCopySource) Err() error {
	return nil
}

// TotalRows 返回总行数（可选实现）
func (s *messageCopySource) TotalRows() int {
	return len(s.rows)
}
