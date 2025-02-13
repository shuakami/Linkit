-- 添加短链接新字段
ALTER TABLE short_links
ADD COLUMN max_visits BIGINT,
ADD COLUMN never_expire BOOLEAN NOT NULL DEFAULT FALSE;

-- 更新现有记录
UPDATE short_links
SET never_expire = FALSE
WHERE never_expire IS NULL;

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_short_links_max_visits ON short_links(max_visits);
CREATE INDEX IF NOT EXISTS idx_short_links_never_expire ON short_links(never_expire); 