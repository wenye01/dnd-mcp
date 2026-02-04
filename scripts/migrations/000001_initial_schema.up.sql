-- DND MCP Client 初始化数据库表结构
-- Version: 001
-- Description: 创建 client_sessions, client_messages, persistence_snapshots, schema_migrations 表

-- client_sessions (会话元数据表)
CREATE TABLE IF NOT EXISTS client_sessions (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    name VARCHAR(255) NOT NULL,
    creator_id VARCHAR(255) NOT NULL,
    mcp_server_url VARCHAR(512) NOT NULL,
    websocket_key VARCHAR(255) NOT NULL,
    max_players INTEGER,
    settings JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'active'
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_sessions_created_at ON client_sessions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sessions_updated_at ON client_sessions(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_sessions_deleted_at ON client_sessions(deleted_at);
CREATE INDEX IF NOT EXISTS idx_sessions_creator_id ON client_sessions(creator_id);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON client_sessions(status);

-- client_messages (对话消息表)
CREATE TABLE IF NOT EXISTS client_messages (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL,
    role VARCHAR(20) NOT NULL,
    content TEXT,
    tool_calls JSONB,
    player_id VARCHAR(255),
    CONSTRAINT fk_session
        FOREIGN KEY(session_id)
        REFERENCES client_sessions(id)
        ON DELETE CASCADE,
    CONSTRAINT valid_role
        CHECK (role IN ('system', 'user', 'assistant', 'tool'))
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_messages_session_time
    ON client_messages(session_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_role ON client_messages(role);
CREATE INDEX IF NOT EXISTS idx_messages_player ON client_messages(player_id);

-- persistence_snapshots (持久化快照表)
CREATE TABLE IF NOT EXISTS persistence_snapshots (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    session_count INTEGER NOT NULL,
    message_count INTEGER NOT NULL,
    duration_ms INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_snapshots_created_at
    ON persistence_snapshots(created_at DESC);

-- schema_migrations (迁移历史表)
CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT PRIMARY KEY,
    applied_at TIMESTAMP NOT NULL
);
