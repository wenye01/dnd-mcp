-- 006_visual_locations.up.sql
-- Add visual_locations and mode support for Image mode maps

-- Add mode column to maps for grid/image mode selection
ALTER TABLE maps ADD COLUMN IF NOT EXISTS mode VARCHAR(50) DEFAULT 'grid';

-- Add visual_locations column to maps for Image mode world map locations
ALTER TABLE maps ADD COLUMN IF NOT EXISTS visual_locations JSONB DEFAULT '[]';

-- Add comments
COMMENT ON COLUMN maps.mode IS 'Map mode: grid (default) or image';
COMMENT ON COLUMN maps.visual_locations IS 'AI-identified locations on Image mode maps with normalized coordinates (0-1)';
