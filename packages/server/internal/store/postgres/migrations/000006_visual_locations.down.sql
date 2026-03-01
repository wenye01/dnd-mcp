-- 006_visual_locations.down.sql
-- Rollback visual_locations support

-- Remove mode and visual_locations columns from maps
ALTER TABLE maps DROP COLUMN IF EXISTS mode;
ALTER TABLE maps DROP COLUMN IF EXISTS visual_locations;
