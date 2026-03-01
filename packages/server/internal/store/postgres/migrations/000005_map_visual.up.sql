-- 005_map_visual.up.sql
-- Add visual map support: PlayerMarker for game_states

-- Add player_marker column to game_states for Image mode support
ALTER TABLE game_states ADD COLUMN IF NOT EXISTS player_marker JSONB;

-- Add comment
COMMENT ON COLUMN game_states.player_marker IS 'Player marker position for Image mode world maps (normalized coordinates 0-1)';
