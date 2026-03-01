-- 005_map_visual.down.sql
-- Rollback visual map support

-- Remove player_marker column from game_states
ALTER TABLE game_states DROP COLUMN IF EXISTS player_marker;
