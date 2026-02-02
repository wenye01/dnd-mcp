-- DND MCP Client 初始数据库Schema

-- 扩展UUID支持
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 会话表
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    -- 游戏元数据
    game_time VARCHAR(50) NOT NULL DEFAULT 'Morning',
    location VARCHAR(255) NOT NULL DEFAULT 'Unknown',
    campaign_name VARCHAR(255) NOT NULL,

    -- 完整状态(JSONB格式,灵活存储)
    state JSONB NOT NULL DEFAULT '{}'::jsonb
);

-- 会话索引
CREATE INDEX idx_sessions_created_at ON sessions(created_at DESC);
CREATE INDEX idx_sessions_deleted_at ON sessions(deleted_at);
CREATE INDEX idx_sessions_campaign_name ON sessions(campaign_name);

-- 消息表
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- 消息内容
    role VARCHAR(20) NOT NULL CHECK (role IN ('system', 'user', 'assistant', 'tool')),
    content TEXT NOT NULL,

    -- 工具调用(JSONB格式)
    tool_calls JSONB DEFAULT '[]'::jsonb,

    -- 玩家ID(可选,用于标识玩家消息)
    player_id VARCHAR(255)
);

-- 消息索引
CREATE INDEX idx_messages_session_id ON messages(session_id);
CREATE INDEX idx_messages_created_at ON messages(created_at ASC);
CREATE INDEX idx_messages_session_created ON messages(session_id, created_at ASC);

-- 添加注释
COMMENT ON TABLE sessions IS '游戏会话表,存储D&D游戏会话状态';
COMMENT ON COLUMN sessions.state IS '游戏完整状态,JSONB格式存储';

COMMENT ON TABLE messages IS '对话消息表,存储AI和玩家的对话历史';
COMMENT ON COLUMN messages.tool_calls IS 'LLM工具调用记录';
