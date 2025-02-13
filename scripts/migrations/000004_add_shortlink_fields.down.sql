-- 删除索引
DROP INDEX IF EXISTS idx_short_links_max_visits;
DROP INDEX IF EXISTS idx_short_links_never_expire;

-- 删除字段
ALTER TABLE short_links
DROP COLUMN IF EXISTS max_visits,
DROP COLUMN IF EXISTS never_expire; 