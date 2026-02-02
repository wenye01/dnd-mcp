-- 回滚初始Schema

DROP INDEX IF EXISTS idx_messages_session_created;
DROP INDEX IF EXISTS idx_messages_created_at;
DROP INDEX IF EXISTS idx_messages_session_id;
DROP TABLE IF EXISTS messages;

DROP INDEX IF EXISTS idx_sessions_campaign_name;
DROP INDEX IF EXISTS idx_sessions_deleted_at;
DROP INDEX IF EXISTS idx_sessions_created_at;
DROP TABLE IF EXISTS sessions;

DROP EXTENSION IF EXISTS "uuid-ossp";
