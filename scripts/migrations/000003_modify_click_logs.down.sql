-- 恢复country字段长度
ALTER TABLE click_logs
ALTER COLUMN country TYPE VARCHAR(2); 