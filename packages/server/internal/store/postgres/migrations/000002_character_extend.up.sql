-- 000002_character_extend.up.sql
-- 扩展 Character 模型以支持 FVTT 导入所需的完整字段

-- 添加扩展字段到 characters 表
ALTER TABLE characters ADD COLUMN IF NOT EXISTS image TEXT;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS experience INT DEFAULT 0;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS proficiency INT DEFAULT 0;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS speed_detail JSONB;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS death_saves JSONB;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS skills_detail JSONB;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS saves_detail JSONB;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS currency JSONB;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS equipment_slots JSONB;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS inventory_items JSONB;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS spellbook JSONB;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS features JSONB;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS biography JSONB;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS traits JSONB;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS import_meta JSONB;

-- 添加注释
COMMENT ON COLUMN characters.image IS 'Character image (Base64 or URL)';
COMMENT ON COLUMN characters.experience IS 'Experience points';
COMMENT ON COLUMN characters.proficiency IS 'Proficiency bonus override (0 = auto-calculate)';
COMMENT ON COLUMN characters.speed_detail IS 'Detailed movement speeds (walk, burrow, climb, fly, swim)';
COMMENT ON COLUMN characters.death_saves IS 'Death save tracking (successes, failures)';
COMMENT ON COLUMN characters.skills_detail IS 'Detailed skill information with proficiency/expertise';
COMMENT ON COLUMN characters.saves_detail IS 'Detailed saving throw information with proficiency';
COMMENT ON COLUMN characters.currency IS 'Currency amounts (pp, gp, ep, sp, cp)';
COMMENT ON COLUMN characters.equipment_slots IS 'Structured equipment slots';
COMMENT ON COLUMN characters.inventory_items IS 'Detailed inventory items with weight and charges';
COMMENT ON COLUMN characters.spellbook IS 'Spell slots, known spells, and prepared spells';
COMMENT ON COLUMN characters.features IS 'Racial traits, class features, and feats';
COMMENT ON COLUMN characters.biography IS 'Character biography and personality';
COMMENT ON COLUMN characters.traits IS 'Damage resistances, immunities, languages, senses';
COMMENT ON COLUMN characters.import_meta IS 'Import metadata for FVTT/UVTT imported characters';
