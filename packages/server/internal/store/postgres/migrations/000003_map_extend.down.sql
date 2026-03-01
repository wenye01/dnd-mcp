-- 000003_map_extend.down.sql
-- 回滚地图扩展字段

ALTER TABLE maps DROP COLUMN IF EXISTS image;
ALTER TABLE maps DROP COLUMN IF EXISTS walls;
ALTER TABLE maps DROP COLUMN IF EXISTS import_meta;
