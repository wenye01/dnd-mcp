-- DND MCP Client 回滚初始数据库表结构
-- Version: 001
-- Description: 删除所有表

-- 删除表（注意顺序：先删除有外键依赖的表）
DROP INDEX IF EXISTS idx_snapshots_created_at;
DROP TABLE IF EXISTS persistence_snapshots;

DROP INDEX IF EXISTS idx_messages_session_time;
DROP INDEX IF EXISTS idx_messages_role;
DROP INDEX IF EXISTS idx_messages_player;
DROP TABLE IF EXISTS client_messages;

DROP INDEX IF EXISTS idx_sessions_created_at;
DROP INDEX IF EXISTS idx_sessions_updated_at;
DROP INDEX IF EXISTS idx_sessions_deleted_at;
DROP INDEX IF EXISTS idx_sessions_creator_id;
DROP INDEX IF EXISTS idx_sessions_status;
DROP TABLE IF EXISTS client_sessions;

DROP TABLE IF EXISTS schema_migrations;
