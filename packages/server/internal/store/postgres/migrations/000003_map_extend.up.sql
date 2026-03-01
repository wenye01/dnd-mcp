-- 000003_map_extend.up.sql
-- 扩展 Map 和 Token 模型以支持 FVTT 导入所需的完整字段

-- 添加图片和墙体字段到 maps 表
ALTER TABLE maps ADD COLUMN IF NOT EXISTS image JSONB;
ALTER TABLE maps ADD COLUMN IF NOT EXISTS walls JSONB DEFAULT '[]';
ALTER TABLE maps ADD COLUMN IF NOT EXISTS import_meta JSONB;

-- 添加注释
COMMENT ON COLUMN maps.image IS 'Map background image (URL, dimensions, scale)';
COMMENT ON COLUMN maps.walls IS 'Wall definitions for tactical combat';
COMMENT ON COLUMN maps.import_meta IS 'Import metadata for FVTT/UVTT imported maps';
