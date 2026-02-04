-- 添加 sessions 表缺失的字段
-- Version: 002
-- Description: 添加 websocket_key, max_players, settings 字段到 sessions 表

-- 添加 websocket_key 字段
ALTER TABLE client_sessions ADD COLUMN IF NOT EXISTS websocket_key TEXT;

-- 添加 max_players 字段
ALTER TABLE client_sessions ADD COLUMN IF NOT EXISTS max_players INTEGER NOT NULL DEFAULT 4;

-- 添加 settings 字段
ALTER TABLE client_sessions ADD COLUMN IF NOT EXISTS settings JSONB DEFAULT '{}'::jsonb;
