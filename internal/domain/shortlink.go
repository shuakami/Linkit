package domain

import (
	"time"
)

// RedirectType 表示跳转类型
type RedirectType int

const (
	// RedirectPermanent 永久重定向 (301)
	RedirectPermanent RedirectType = iota + 1
	// RedirectTemporary 临时重定向 (302)
	RedirectTemporary
	// RedirectTemporaryKeepMethod 临时重定向保持方法 (307)
	RedirectTemporaryKeepMethod
	// RedirectPermanentKeepMethod 永久重定向保持方法 (308)
	RedirectPermanentKeepMethod
)

// DeviceType 表示设备类型
type DeviceType int

const (
	// DeviceAll 所有设备
	DeviceAll DeviceType = iota
	// DeviceMobile 移动设备
	DeviceMobile
	// DeviceDesktop 桌面设备
	DeviceDesktop
	// DeviceTablet 平板设备
	DeviceTablet
)

// RedirectRule 表示跳转规则
type RedirectRule struct {
	ID          uint         `json:"id" gorm:"column:id;primaryKey"`
	ShortLinkID uint         `json:"short_link_id" gorm:"column:short_link_id"`
	Name        string       `json:"name" gorm:"column:name"`                                    // 规则名称
	Description string       `json:"description" gorm:"column:description"`                      // 规则描述
	Priority    int          `json:"priority" gorm:"column:priority;default:0"`                  // 优先级，数字越大优先级越高
	Type        RedirectType `json:"type" gorm:"column:type"`                                    // 跳转类型
	TargetURL   string       `json:"target_url" gorm:"column:target_url"`                        // 目标URL，为空则使用短链接的原始URL
	Device      DeviceType   `json:"device" gorm:"column:device;default:0"`                      // 设备类型
	StartTime   *time.Time   `json:"start_time" gorm:"column:start_time"`                        // 生效开始时间
	EndTime     *time.Time   `json:"end_time" gorm:"column:end_time"`                            // 生效结束时间
	Countries   []string     `json:"countries" gorm:"column:countries;type:text[];default:'{}'"` // 国家列表
	Provinces   []string     `json:"provinces" gorm:"column:provinces;type:text[];default:'{}'"` // 省份列表
	Cities      []string     `json:"cities" gorm:"column:cities;type:text[];default:'{}'"`       // 城市列表
	Percentage  *int         `json:"percentage" gorm:"column:percentage"`                        // A/B测试流量百分比（1-100）
	MaxVisits   *int         `json:"max_visits" gorm:"column:max_visits"`                        // 最大访问次数
	CreatedAt   time.Time    `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time    `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名
func (RedirectRule) TableName() string {
	return "redirect_rules"
}

// ShortLink 表示一个短链接实体
type ShortLink struct {
	ID              uint           `json:"id" gorm:"column:id;primaryKey"`
	ShortCode       string         `json:"short_code" gorm:"column:short_code;uniqueIndex"`
	LongURL         string         `json:"long_url" gorm:"column:long_url"`
	UserID          uint           `json:"user_id,omitempty" gorm:"column:user_id"`
	Clicks          uint64         `json:"clicks" gorm:"column:clicks;default:0"`
	MaxVisits       *uint64        `json:"max_visits" gorm:"column:max_visits"` // 最大访问次数限制
	ExpiresAt       time.Time      `json:"expires_at" gorm:"column:expires_at"`
	NeverExpire     bool           `json:"never_expire" gorm:"column:never_expire;default:false"`     // 是否永不过期
	DefaultRedirect RedirectType   `json:"default_redirect" gorm:"column:default_redirect;default:1"` // 默认跳转类型
	Rules           []RedirectRule `json:"rules,omitempty" gorm:"-"`                                  // 跳转规则列表
	CreatedAt       time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名
func (ShortLink) TableName() string {
	return "short_links"
}

// CreateShortLinkInput 表示创建短链接的输入参数
type CreateShortLinkInput struct {
	LongURL         string       `json:"long_url" binding:"required,url"`
	CustomCode      string       `json:"custom_code,omitempty"`
	ExpiresAt       time.Time    `json:"expires_at,omitempty"`
	UserID          uint         `json:"user_id,omitempty"`
	DefaultRedirect RedirectType `json:"default_redirect,omitempty"` // 默认跳转类型
	NeverExpire     bool         `json:"never_expire,omitempty"`     // 是否永不过期
}

// CreateRuleInput 表示创建跳转规则的输入参数
type CreateRuleInput struct {
	ShortLinkID uint         `json:"short_link_id"`
	Name        string       `json:"name" binding:"required"`
	Description string       `json:"description"`
	Priority    int          `json:"priority"`
	Type        RedirectType `json:"type" binding:"required"`
	TargetURL   string       `json:"target_url"`
	Device      DeviceType   `json:"device"`
	StartTime   *time.Time   `json:"start_time"`
	EndTime     *time.Time   `json:"end_time"`
	Countries   []string     `json:"countries"`
	Percentage  *int         `json:"percentage"`
	MaxVisits   *int         `json:"max_visits"`
}

// ClickLog 表示一个点击日志实体
type ClickLog struct {
	ID          uint       `json:"id" gorm:"column:id;primaryKey"`
	ShortLinkID uint       `json:"short_link_id" gorm:"column:short_link_id"`
	RuleID      *uint      `json:"rule_id" gorm:"column:rule_id"` // 使用的规则ID
	IP          string     `json:"ip" gorm:"column:ip"`
	UserAgent   string     `json:"user_agent" gorm:"column:user_agent"`
	Referer     string     `json:"referer" gorm:"column:referer"`
	Country     string     `json:"country" gorm:"column:country"`         // 访问者国家/地区
	Device      DeviceType `json:"device" gorm:"column:device;default:0"` // 访问者设备类型
	CreatedAt   time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

// TableName 指定表名
func (ClickLog) TableName() string {
	return "click_logs"
}

// SortDirection 表示排序方向
type SortDirection string

const (
	// SortAsc 升序排序
	SortAsc SortDirection = "asc"
	// SortDesc 降序排序
	SortDesc SortDirection = "desc"
)

// ShortLinkFilter 表示短链接查询过滤条件
type ShortLinkFilter struct {
	UserID    *uint      `json:"user_id,omitempty"`    // 用户ID过滤
	IsExpired *bool      `json:"is_expired,omitempty"` // 是否已过期
	StartTime *time.Time `json:"start_time,omitempty"` // 创建时间范围开始
	EndTime   *time.Time `json:"end_time,omitempty"`   // 创建时间范围结束
	MinClicks *uint64    `json:"min_clicks,omitempty"` // 最小点击数
	MaxClicks *uint64    `json:"max_clicks,omitempty"` // 最大点击数
}

// ShortLinkSort 表示短链接排序条件
type ShortLinkSort struct {
	Field     string        `json:"field"`     // 排序字段
	Direction SortDirection `json:"direction"` // 排序方向
}

// PaginationQuery 表示分页查询参数
type PaginationQuery struct {
	Page     int              `json:"page" binding:"required,min=1"`              // 页码，从1开始
	PageSize int              `json:"page_size" binding:"required,min=1,max=100"` // 每页数量
	Filter   *ShortLinkFilter `json:"filter,omitempty"`                           // 过滤条件
	Sort     *ShortLinkSort   `json:"sort,omitempty"`                             // 排序条件
}

// PaginatedShortLinks 表示分页的短链接列表
type PaginatedShortLinks struct {
	Total       int64       `json:"total"`        // 总记录数
	TotalPages  int         `json:"total_pages"`  // 总页数
	CurrentPage int         `json:"current_page"` // 当前页码
	PageSize    int         `json:"page_size"`    // 每页数量
	Data        []ShortLink `json:"data"`         // 当前页数据
}

// UpdateShortLinkInput 表示更新短链接的输入参数
type UpdateShortLinkInput struct {
	LongURL         *string       `json:"long_url,omitempty"`
	MaxVisits       *uint64       `json:"max_visits,omitempty"`
	ExpiresAt       *time.Time    `json:"expires_at,omitempty"`
	NeverExpire     *bool         `json:"never_expire,omitempty"`
	DefaultRedirect *RedirectType `json:"default_redirect,omitempty"`
}

// ClickLogFilter 表示访问记录查询过滤条件
type ClickLogFilter struct {
	StartTime *time.Time  `json:"start_time,omitempty"` // 开始时间
	EndTime   *time.Time  `json:"end_time,omitempty"`   // 结束时间
	IP        *string     `json:"ip,omitempty"`         // IP地址
	Country   *string     `json:"country,omitempty"`    // 国家/地区
	Device    *DeviceType `json:"device,omitempty"`     // 设备类型
	RuleID    *uint       `json:"rule_id,omitempty"`    // 规则ID
}

// ClickLogSort 表示访问记录排序条件
type ClickLogSort struct {
	Field     string        `json:"field"`     // 排序字段
	Direction SortDirection `json:"direction"` // 排序方向
}

// ClickLogQuery 表示访问记录查询参数
type ClickLogQuery struct {
	Page     int             `json:"page" binding:"required,min=1"`              // 页码，从1开始
	PageSize int             `json:"page_size" binding:"required,min=1,max=100"` // 每页数量
	Filter   *ClickLogFilter `json:"filter,omitempty"`                           // 过滤条件
	Sort     *ClickLogSort   `json:"sort,omitempty"`                             // 排序条件
}

// PaginatedClickLogs 表示分页的访问记录列表
type PaginatedClickLogs struct {
	Total       int64      `json:"total"`        // 总记录数
	TotalPages  int        `json:"total_pages"`  // 总页数
	CurrentPage int        `json:"current_page"` // 当前页码
	PageSize    int        `json:"page_size"`    // 每页数量
	Data        []ClickLog `json:"data"`         // 当前页数据
}

// ShortLinkRepository 定义短链接仓储接口
type ShortLinkRepository interface {
	Create(link *ShortLink) error
	GetByCode(code string) (*ShortLink, error)
	Update(link *ShortLink) error
	Delete(code string) error
	IncrementClicks(code string) error
	LogClick(log *ClickLog) error
	List(query *PaginationQuery) (*PaginatedShortLinks, error)
	ListClickLogs(shortLinkID uint, query *ClickLogQuery) (*PaginatedClickLogs, error) // 新增：获取访问记录列表

	// 规则相关
	CreateRule(rule *RedirectRule) error
	UpdateRule(rule *RedirectRule) error
	DeleteRule(ruleID uint) error
	GetRules(shortLinkID uint) ([]RedirectRule, error)
	UpdateRules(shortLinkID uint, rules []RedirectRule) error
}

// ShortLinkUseCase 定义短链接用例接口
type ShortLinkUseCase interface {
	Create(input *CreateShortLinkInput) (*ShortLink, error)
	Get(code string) (*ShortLink, error)
	Redirect(code string, clickLog *ClickLog) (string, RedirectType, error)
	Delete(code string) error
	List(query *PaginationQuery) (*PaginatedShortLinks, error)
	Update(code string, input *UpdateShortLinkInput) (*ShortLink, error)
	ListClickLogs(code string, query *ClickLogQuery) (*PaginatedClickLogs, error)

	// 规则相关
	CreateRule(input *CreateRuleInput) (*RedirectRule, error)
	UpdateRule(ruleID uint, input *CreateRuleInput) (*RedirectRule, error)
	DeleteRule(ruleID uint) error
	GetRules(shortLinkID uint) ([]RedirectRule, error)
	UpdateRules(shortLinkID uint, rules []CreateRuleInput) ([]RedirectRule, error)
}
