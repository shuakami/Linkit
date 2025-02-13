-- 删除索引
DROP INDEX IF EXISTS idx_redirect_rules_provinces;
DROP INDEX IF EXISTS idx_redirect_rules_cities;

-- 删除字段
ALTER TABLE redirect_rules
DROP COLUMN IF EXISTS provinces,
DROP COLUMN IF EXISTS cities; 