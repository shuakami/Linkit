-- 添加省份和城市字段
ALTER TABLE redirect_rules
ADD COLUMN provinces TEXT[],
ADD COLUMN cities TEXT[];

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_redirect_rules_provinces ON redirect_rules USING GIN(provinces);
CREATE INDEX IF NOT EXISTS idx_redirect_rules_cities ON redirect_rules USING GIN(cities); 