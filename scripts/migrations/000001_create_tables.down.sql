-- 删除索引
DROP INDEX IF EXISTS idx_click_logs_country;
DROP INDEX IF EXISTS idx_click_logs_created_at;
DROP INDEX IF EXISTS idx_click_logs_rule_id;
DROP INDEX IF EXISTS idx_click_logs_short_link_id;
DROP INDEX IF EXISTS idx_redirect_rules_priority;
DROP INDEX IF EXISTS idx_redirect_rules_short_link_id;
DROP INDEX IF EXISTS idx_short_links_user_id;
DROP INDEX IF EXISTS idx_short_links_short_code;

-- 删除表
DROP TABLE IF EXISTS click_logs;
DROP TABLE IF EXISTS redirect_rules;
DROP TABLE IF EXISTS short_links; 