-- 000002_character_extend.down.sql
-- 回滚 Character 扩展字段

ALTER TABLE characters DROP COLUMN IF EXISTS import_meta;
ALTER TABLE characters DROP COLUMN IF EXISTS traits;
ALTER TABLE characters DROP COLUMN IF EXISTS biography;
ALTER TABLE characters DROP COLUMN IF EXISTS features;
ALTER TABLE characters DROP COLUMN IF EXISTS spellbook;
ALTER TABLE characters DROP COLUMN IF EXISTS inventory_items;
ALTER TABLE characters DROP COLUMN IF EXISTS equipment_slots;
ALTER TABLE characters DROP COLUMN IF EXISTS currency;
ALTER TABLE characters DROP COLUMN IF EXISTS saves_detail;
ALTER TABLE characters DROP COLUMN IF EXISTS skills_detail;
ALTER TABLE characters DROP COLUMN IF EXISTS death_saves;
ALTER TABLE characters DROP COLUMN IF EXISTS speed_detail;
ALTER TABLE characters DROP COLUMN IF EXISTS proficiency;
ALTER TABLE characters DROP COLUMN IF EXISTS experience;
ALTER TABLE characters DROP COLUMN IF EXISTS image;
