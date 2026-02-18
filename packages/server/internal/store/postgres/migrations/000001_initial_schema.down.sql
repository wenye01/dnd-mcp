-- 000001_initial_schema.down.sql
-- Rollback initial schema

-- Drop triggers
DROP TRIGGER IF EXISTS update_game_states_updated_at ON game_states;
DROP TRIGGER IF EXISTS update_maps_updated_at ON maps;
DROP TRIGGER IF EXISTS update_characters_updated_at ON characters;
DROP TRIGGER IF EXISTS update_campaigns_updated_at ON campaigns;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order (respecting foreign keys)
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS game_states;
DROP TABLE IF EXISTS maps;
DROP TABLE IF EXISTS combats;
DROP TABLE IF EXISTS characters;
DROP TABLE IF EXISTS campaigns;

-- Drop extension (optional, may want to keep it)
-- DROP EXTENSION IF EXISTS "uuid-ossp";
