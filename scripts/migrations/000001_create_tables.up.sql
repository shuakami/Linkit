-- 创建短链接表
CREATE TABLE IF NOT EXISTS short_links (
    id SERIAL PRIMARY KEY,
    short_code VARCHAR(16) UNIQUE NOT NULL,
    long_url TEXT NOT NULL,
    user_id INTEGER,
    clicks BIGINT DEFAULT 0,
    default_redirect INTEGER NOT NULL DEFAULT 1, -- 默认跳转类型：1=301, 2=302, 3=307, 4=308
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建跳转规则表
CREATE TABLE IF NOT EXISTS redirect_rules (
    id SERIAL PRIMARY KEY,
    short_link_id INTEGER REFERENCES short_links(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    priority INTEGER NOT NULL DEFAULT 0,
    type INTEGER NOT NULL, -- 跳转类型：1=301, 2=302, 3=307, 4=308
    target_url TEXT, -- 为空则使用短链接的原始URL
    device INTEGER NOT NULL DEFAULT 0, -- 0=all, 1=mobile, 2=desktop, 3=tablet
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,
    countries TEXT[], -- 国家/地区代码列表
    percentage INTEGER CHECK (percentage BETWEEN 1 AND 100), -- A/B测试流量百分比
    max_visits INTEGER, -- 最大访问次数
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建点击日志表
CREATE TABLE IF NOT EXISTS click_logs (
    id SERIAL PRIMARY KEY,
    short_link_id INTEGER REFERENCES short_links(id) ON DELETE CASCADE,
    rule_id INTEGER REFERENCES redirect_rules(id) ON DELETE SET NULL,
    ip VARCHAR(45) NOT NULL,
    user_agent TEXT,
    referer TEXT,
    country VARCHAR(2), -- ISO 3166-1 alpha-2 国家代码
    device INTEGER NOT NULL DEFAULT 0, -- 0=all, 1=mobile, 2=desktop, 3=tablet
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_short_links_short_code ON short_links(short_code);
CREATE INDEX IF NOT EXISTS idx_short_links_user_id ON short_links(user_id);
CREATE INDEX IF NOT EXISTS idx_redirect_rules_short_link_id ON redirect_rules(short_link_id);
CREATE INDEX IF NOT EXISTS idx_redirect_rules_priority ON redirect_rules(priority);
CREATE INDEX IF NOT EXISTS idx_click_logs_short_link_id ON click_logs(short_link_id);
CREATE INDEX IF NOT EXISTS idx_click_logs_rule_id ON click_logs(rule_id);
CREATE INDEX IF NOT EXISTS idx_click_logs_created_at ON click_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_click_logs_country ON click_logs(country); 