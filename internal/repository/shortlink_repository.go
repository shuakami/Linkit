package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"linkit/internal/domain"

	"github.com/go-redis/redis/v8"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// ShortLinkRepository 实现短链接仓储接口
type ShortLinkRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewShortLinkRepository 创建短链接仓储实例
func NewShortLinkRepository(db *gorm.DB, redis *redis.Client) domain.ShortLinkRepository {
	return &ShortLinkRepository{
		db:    db,
		redis: redis,
	}
}

// getCacheKey 获取缓存键
func (r *ShortLinkRepository) getCacheKey(code string) string {
	return fmt.Sprintf("link:%s", code)
}

// setCache 设置缓存
func (r *ShortLinkRepository) setCache(ctx context.Context, link *domain.ShortLink) error {
	// 计算剩余过期时间
	expiration := time.Until(link.ExpiresAt)
	if expiration <= 0 {
		return nil // 已过期，不缓存
	}

	// 创建缓存数据结构
	cacheData := struct {
		ID              uint      `json:"id"`
		LongURL         string    `json:"long_url"`
		ExpiresAt       time.Time `json:"expires_at"`
		Clicks          uint64    `json:"clicks"`
		MaxVisits       *uint64   `json:"max_visits"`
		DefaultRedirect uint      `json:"default_redirect"`
		NeverExpire     bool      `json:"never_expire"`
		CreatedAt       time.Time `json:"created_at"`
		UpdatedAt       time.Time `json:"updated_at"`
	}{
		ID:              link.ID,
		LongURL:         link.LongURL,
		ExpiresAt:       link.ExpiresAt,
		Clicks:          link.Clicks,
		MaxVisits:       link.MaxVisits,
		DefaultRedirect: uint(link.DefaultRedirect),
		NeverExpire:     link.NeverExpire,
		CreatedAt:       link.CreatedAt,
		UpdatedAt:       link.UpdatedAt,
	}

	// 序列化数据
	data, err := json.Marshal(cacheData)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	// 设置缓存，过期时间与短链接一致
	return r.redis.Set(ctx, r.getCacheKey(link.ShortCode), string(data), expiration).Err()
}

// Create 创建短链接
func (r *ShortLinkRepository) Create(link *domain.ShortLink) error {
	// 使用事务
	return r.db.Transaction(func(tx *gorm.DB) error {
		fmt.Printf("Creating short link: %s -> %s\n", link.ShortCode, link.LongURL)
		if err := tx.Table("short_links").Create(link).Error; err != nil {
			fmt.Printf("Failed to create short link: %v\n", err)
			return fmt.Errorf("failed to create short link: %w", err)
		}
		fmt.Printf("Short link created successfully: %s\n", link.ShortCode)

		// 设置缓存
		if err := r.setCache(context.Background(), link); err != nil {
			// 缓存错误只记录，不影响事务
			fmt.Printf("Failed to set cache: %v\n", err)
		}

		return nil
	})
}

// GetByCode 根据短码获取短链接
func (r *ShortLinkRepository) GetByCode(code string) (*domain.ShortLink, error) {
	ctx := context.Background()
	cacheKey := r.getCacheKey(code)

	fmt.Printf("[GetByCode] Starting to get short link for code: %s\n", code)

	// 先从Redis缓存中获取
	if data, err := r.redis.Get(ctx, cacheKey).Result(); err == nil && data != "" {
		fmt.Printf("[GetByCode] Found in cache: %s, data: %s\n", code, data)

		// 解析缓存数据
		var cacheData struct {
			ID              uint      `json:"id"`
			LongURL         string    `json:"long_url"`
			ExpiresAt       time.Time `json:"expires_at"`
			Clicks          uint64    `json:"clicks"`
			MaxVisits       *uint64   `json:"max_visits"`
			DefaultRedirect uint      `json:"default_redirect"`
			NeverExpire     bool      `json:"never_expire"`
			CreatedAt       time.Time `json:"created_at"`
			UpdatedAt       time.Time `json:"updated_at"`
		}
		if err := json.Unmarshal([]byte(data), &cacheData); err != nil {
			fmt.Printf("[GetByCode] Failed to unmarshal cache data: %v\n", err)
		} else {
			fmt.Printf("[GetByCode] Cache data parsed successfully: %+v\n", cacheData)
			return &domain.ShortLink{
				ID:              cacheData.ID,
				ShortCode:       code,
				LongURL:         cacheData.LongURL,
				ExpiresAt:       cacheData.ExpiresAt,
				Clicks:          cacheData.Clicks,
				MaxVisits:       cacheData.MaxVisits,
				DefaultRedirect: domain.RedirectType(cacheData.DefaultRedirect),
				NeverExpire:     cacheData.NeverExpire,
				CreatedAt:       cacheData.CreatedAt,
				UpdatedAt:       cacheData.UpdatedAt,
			}, nil
		}
	} else {
		fmt.Printf("[GetByCode] Cache miss for code: %s, error: %v\n", code, err)
	}

	// 从数据库中获取
	var link domain.ShortLink
	fmt.Printf("[GetByCode] Starting database query for code: %s\n", code)

	err := r.db.Table("short_links").
		Select("id, short_code, long_url, user_id, clicks, max_visits, expires_at, never_expire, default_redirect, created_at, updated_at").
		Where("short_code = ?", code).
		First(&link).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Printf("[GetByCode] No record found in database for code: %s\n", code)
			// 设置空值缓存，防止缓存穿透
			r.redis.Set(ctx, cacheKey, "", 5*time.Minute)
			return nil, domain.ErrShortLinkNotFound
		}
		fmt.Printf("[GetByCode] Database error for code %s: %v\n", code, err)
		return nil, domain.ErrShortLinkNotFound
	}

	// 如果default_redirect为0,设置为默认值1
	if link.DefaultRedirect == 0 {
		link.DefaultRedirect = domain.RedirectPermanent
	}

	fmt.Printf("[GetByCode] Found in database: %+v\n", link)

	// 设置缓存
	if err := r.setCache(ctx, &link); err != nil {
		fmt.Printf("[GetByCode] Failed to set cache for code %s: %v\n", code, err)
	} else {
		fmt.Printf("[GetByCode] Cache set successfully for code: %s\n", code)
	}

	return &link, nil
}

// Update 更新短链接
func (r *ShortLinkRepository) Update(link *domain.ShortLink) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("short_links").Save(link).Error; err != nil {
			return fmt.Errorf("failed to update short link: %w", err)
		}

		// 更新缓存
		if err := r.setCache(context.Background(), link); err != nil {
			fmt.Printf("failed to update cache: %v\n", err)
		}

		return nil
	})
}

// Delete 删除短链接
func (r *ShortLinkRepository) Delete(code string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		fmt.Printf("Deleting short link: %s\n", code)
		// 删除数据库记录
		if err := tx.Table("short_links").Where("short_code = ?", code).Delete(&domain.ShortLink{}).Error; err != nil {
			fmt.Printf("Failed to delete short link: %v\n", err)
			return fmt.Errorf("failed to delete short link: %w", err)
		}
		fmt.Printf("Short link deleted successfully: %s\n", code)

		// 删除缓存
		ctx := context.Background()
		if err := r.redis.Del(ctx, r.getCacheKey(code)).Err(); err != nil {
			fmt.Printf("Failed to delete cache: %v\n", err)
		}

		return nil
	})
}

// getRulesCacheKey 获取规则缓存键
func (r *ShortLinkRepository) getRulesCacheKey(shortLinkID uint) string {
	return fmt.Sprintf("rules:%d", shortLinkID)
}

// setRulesCache 设置规则缓存
func (r *ShortLinkRepository) setRulesCache(ctx context.Context, shortLinkID uint, rules []domain.RedirectRule) error {
	data, err := json.Marshal(rules)
	if err != nil {
		return err
	}
	// 缓存规则,过期时间5分钟
	return r.redis.Set(ctx, r.getRulesCacheKey(shortLinkID), string(data), 5*time.Minute).Err()
}

// GetRules 获取短链接的所有规则
func (r *ShortLinkRepository) GetRules(shortLinkID uint) ([]domain.RedirectRule, error) {
	ctx := context.Background()
	cacheKey := r.getRulesCacheKey(shortLinkID)

	// 尝试从缓存获取
	if data, err := r.redis.Get(ctx, cacheKey).Result(); err == nil {
		var rules []domain.RedirectRule
		if err := json.Unmarshal([]byte(data), &rules); err == nil {
			return rules, nil
		}
	}

	// 缓存未命中,从数据库查询
	var rules []domain.RedirectRule

	// 使用临时结构体来处理数组字段
	type tempRule struct {
		ID          uint           `gorm:"column:id"`
		ShortLinkID uint           `gorm:"column:short_link_id"`
		Name        string         `gorm:"column:name"`
		Description string         `gorm:"column:description"`
		Priority    int            `gorm:"column:priority"`
		Type        uint           `gorm:"column:type"`
		TargetURL   string         `gorm:"column:target_url"`
		Device      uint           `gorm:"column:device"`
		StartTime   *time.Time     `gorm:"column:start_time"`
		EndTime     *time.Time     `gorm:"column:end_time"`
		Countries   pq.StringArray `gorm:"column:countries;type:text[]"`
		Provinces   pq.StringArray `gorm:"column:provinces;type:text[]"`
		Cities      pq.StringArray `gorm:"column:cities;type:text[]"`
		Percentage  *int           `gorm:"column:percentage"`
		MaxVisits   *int           `gorm:"column:max_visits"`
		CreatedAt   time.Time      `gorm:"column:created_at"`
		UpdatedAt   time.Time      `gorm:"column:updated_at"`
	}

	var tempRules []tempRule

	sql := `
		SELECT id, short_link_id, name, description, priority, type, target_url,
			device, start_time, end_time, countries, provinces, cities,
			percentage, max_visits, created_at, updated_at
		FROM redirect_rules 
		WHERE short_link_id = ?
		ORDER BY priority DESC`

	if err := r.db.Raw(sql, shortLinkID).Scan(&tempRules).Error; err != nil {
		return nil, fmt.Errorf("failed to get rules: %w", err)
	}

	// 转换为domain.RedirectRule
	rules = make([]domain.RedirectRule, len(tempRules))
	for i, tr := range tempRules {
		rules[i] = domain.RedirectRule{
			ID:          tr.ID,
			ShortLinkID: tr.ShortLinkID,
			Name:        tr.Name,
			Description: tr.Description,
			Priority:    tr.Priority,
			Type:        domain.RedirectType(tr.Type),
			TargetURL:   tr.TargetURL,
			Device:      domain.DeviceType(tr.Device),
			StartTime:   tr.StartTime,
			EndTime:     tr.EndTime,
			Countries:   []string(tr.Countries),
			Provinces:   []string(tr.Provinces),
			Cities:      []string(tr.Cities),
			Percentage:  tr.Percentage,
			MaxVisits:   tr.MaxVisits,
			CreatedAt:   tr.CreatedAt,
			UpdatedAt:   tr.UpdatedAt,
		}
	}

	// 设置缓存
	go r.setRulesCache(ctx, shortLinkID, rules)

	return rules, nil
}

// IncrementClicks 增加点击次数(异步)
func (r *ShortLinkRepository) IncrementClicks(code string) error {
	// 使用Redis原子递增
	key := fmt.Sprintf("clicks:%s", code)
	syncKey := fmt.Sprintf("clicks_sync:%s", code)
	cacheKey := r.getCacheKey(code)

	// 递增Redis计数器
	if err := r.redis.Incr(context.Background(), key).Err(); err != nil {
		return err
	}

	// 如果有缓存,也更新缓存中的clicks
	if data, err := r.redis.Get(context.Background(), cacheKey).Result(); err == nil && data != "" {
		var cacheData struct {
			ID              uint      `json:"id"`
			LongURL         string    `json:"long_url"`
			ExpiresAt       time.Time `json:"expires_at"`
			Clicks          uint64    `json:"clicks"`
			MaxVisits       *uint64   `json:"max_visits"`
			DefaultRedirect uint      `json:"default_redirect"`
			CreatedAt       time.Time `json:"created_at"`
			UpdatedAt       time.Time `json:"updated_at"`
		}
		if err := json.Unmarshal([]byte(data), &cacheData); err == nil {
			cacheData.Clicks++
			if newData, err := json.Marshal(cacheData); err == nil {
				// 使用原有的过期时间
				ttl := r.redis.TTL(context.Background(), cacheKey).Val()
				r.redis.Set(context.Background(), cacheKey, string(newData), ttl)
			}
		}
	}

	// 异步更新数据库
	go func() {
		// 获取同步锁,防止并发更新
		if !r.redis.SetNX(context.Background(), syncKey, "1", 10*time.Second).Val() {
			return
		}
		defer r.redis.Del(context.Background(), syncKey)

		// 获取当前计数
		count, err := r.redis.Get(context.Background(), key).Int64()
		if err != nil {
			if err != redis.Nil {
				fmt.Printf("Failed to get click count: %v\n", err)
			}
			return
		}

		// 如果计数大于0,同步到数据库
		if count > 0 {
			// 使用事务保证原子性
			err := r.db.Transaction(func(tx *gorm.DB) error {
				// 更新数据库
				if err := tx.Exec("UPDATE short_links SET clicks = clicks + ? WHERE short_code = ?", count, code).Error; err != nil {
					return err
				}
				// 重置计数器
				if err := r.redis.DecrBy(context.Background(), key, count).Err(); err != nil {
					return err
				}
				return nil
			})

			if err != nil {
				fmt.Printf("Failed to sync clicks: %v\n", err)
				// 出错时等待1秒后重试
				time.Sleep(time.Second)
				return
			}
		}
	}()

	// 启动定时同步
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// 获取同步锁
			if !r.redis.SetNX(context.Background(), syncKey, "1", 10*time.Second).Val() {
				continue
			}
			defer r.redis.Del(context.Background(), syncKey)

			// 获取当前计数
			count, err := r.redis.Get(context.Background(), key).Int64()
			if err != nil {
				if err != redis.Nil {
					fmt.Printf("Failed to get click count in timer: %v\n", err)
				}
				continue
			}

			// 如果有计数,同步到数据库
			if count > 0 {
				err := r.db.Transaction(func(tx *gorm.DB) error {
					if err := tx.Exec("UPDATE short_links SET clicks = clicks + ? WHERE short_code = ?", count, code).Error; err != nil {
						return err
					}
					if err := r.redis.DecrBy(context.Background(), key, count).Err(); err != nil {
						return err
					}
					return nil
				})

				if err != nil {
					fmt.Printf("Failed to sync clicks in timer: %v\n", err)
				}
			}
		}
	}()

	return nil
}

// LogClick 记录点击日志(异步)
func (r *ShortLinkRepository) LogClick(log *domain.ClickLog) error {
	// 异步写入日志
	go func() {
		if err := r.db.Create(log).Error; err != nil {
			fmt.Printf("Failed to create click log: %v\n", err)
		}
	}()
	return nil
}

// CreateRule 创建跳转规则
func (r *ShortLinkRepository) CreateRule(rule *domain.RedirectRule) error {
	fmt.Printf("Creating redirect rule for short link ID: %d\n", rule.ShortLinkID)

	// 构建SQL语句
	sql := `
		INSERT INTO redirect_rules (
			short_link_id, name, description, priority, type, target_url,
			device, start_time, end_time, countries, provinces, cities,
			percentage, max_visits, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?::text[], ?::text[], ?::text[],
			?, ?, ?, ?
		) RETURNING id`

	// 准备参数
	now := time.Now()
	err := r.db.Raw(sql,
		rule.ShortLinkID, rule.Name, rule.Description, rule.Priority, rule.Type, rule.TargetURL,
		rule.Device, rule.StartTime, rule.EndTime, pq.Array(rule.Countries), pq.Array(rule.Provinces), pq.Array(rule.Cities),
		rule.Percentage, rule.MaxVisits, now, now,
	).Scan(&rule.ID).Error

	if err != nil {
		fmt.Printf("Failed to create rule: %v\n", err)
		return fmt.Errorf("failed to create rule: %w", err)
	}

	fmt.Printf("Rule created successfully with ID: %d\n", rule.ID)
	return nil
}

// UpdateRule 更新跳转规则
func (r *ShortLinkRepository) UpdateRule(rule *domain.RedirectRule) error {
	fmt.Printf("Updating redirect rule ID: %d\n", rule.ID)

	// 构建SQL语句
	sql := `
		UPDATE redirect_rules SET
			name = ?, description = ?, priority = ?, type = ?, target_url = ?,
			device = ?, start_time = ?, end_time = ?, countries = ?::text[],
			provinces = ?::text[], cities = ?::text[], percentage = ?,
			max_visits = ?, updated_at = ?
		WHERE id = ?`

	// 准备参数
	now := time.Now()
	err := r.db.Exec(sql,
		rule.Name, rule.Description, rule.Priority, rule.Type, rule.TargetURL,
		rule.Device, rule.StartTime, rule.EndTime, pq.Array(rule.Countries),
		pq.Array(rule.Provinces), pq.Array(rule.Cities), rule.Percentage,
		rule.MaxVisits, now, rule.ID,
	).Error

	if err != nil {
		fmt.Printf("Failed to update rule: %v\n", err)
		return fmt.Errorf("failed to update rule: %w", err)
	}

	fmt.Printf("Rule updated successfully: %d\n", rule.ID)
	return nil
}

// DeleteRule 删除跳转规则
func (r *ShortLinkRepository) DeleteRule(ruleID uint) error {
	fmt.Printf("Deleting redirect rule ID: %d\n", ruleID)
	if err := r.db.Table("redirect_rules").Delete(&domain.RedirectRule{}, ruleID).Error; err != nil {
		fmt.Printf("Failed to delete rule: %v\n", err)
		return fmt.Errorf("failed to delete rule: %w", err)
	}
	fmt.Printf("Rule deleted successfully: %d\n", ruleID)
	return nil
}

// List 获取短链接列表
func (r *ShortLinkRepository) List(query *domain.PaginationQuery) (*domain.PaginatedShortLinks, error) {
	fmt.Printf("[List] Starting to get short links with query: %+v\n", query)

	var total int64
	var links []domain.ShortLink

	// 构建查询
	db := r.db.Table("short_links")

	// 应用过滤条件
	if query.Filter != nil {
		if query.Filter.UserID != nil {
			db = db.Where("user_id = ?", *query.Filter.UserID)
		}
		if query.Filter.IsExpired != nil {
			if *query.Filter.IsExpired {
				db = db.Where("expires_at < ?", time.Now())
			} else {
				db = db.Where("expires_at > ?", time.Now())
			}
		}
		if query.Filter.StartTime != nil {
			db = db.Where("created_at >= ?", query.Filter.StartTime)
		}
		if query.Filter.EndTime != nil {
			db = db.Where("created_at <= ?", query.Filter.EndTime)
		}
		if query.Filter.MinClicks != nil {
			db = db.Where("clicks >= ?", query.Filter.MinClicks)
		}
		if query.Filter.MaxClicks != nil {
			db = db.Where("clicks <= ?", query.Filter.MaxClicks)
		}
	}

	// 获取总记录数
	if err := db.Count(&total).Error; err != nil {
		fmt.Printf("[List] Failed to get total count: %v\n", err)
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// 计算总页数
	totalPages := (int(total) + query.PageSize - 1) / query.PageSize

	// 应用排序
	if query.Sort != nil && query.Sort.Field != "" {
		direction := "DESC"
		if query.Sort.Direction == domain.SortAsc {
			direction = "ASC"
		}
		db = db.Order(fmt.Sprintf("%s %s", query.Sort.Field, direction))
	} else {
		// 默认按创建时间降序
		db = db.Order("created_at DESC")
	}

	// 应用分页
	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Find(&links).Error; err != nil {
		fmt.Printf("[List] Failed to get links: %v\n", err)
		return nil, fmt.Errorf("failed to get links: %w", err)
	}

	fmt.Printf("[List] Successfully got %d links\n", len(links))

	return &domain.PaginatedShortLinks{
		Total:       total,
		TotalPages:  totalPages,
		CurrentPage: query.Page,
		PageSize:    query.PageSize,
		Data:        links,
	}, nil
}

// UpdateRules 批量更新规则
func (r *ShortLinkRepository) UpdateRules(shortLinkID uint, rules []domain.RedirectRule) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除现有规则
		if err := tx.Where("short_link_id = ?", shortLinkID).Delete(&domain.RedirectRule{}).Error; err != nil {
			return fmt.Errorf("failed to delete existing rules: %w", err)
		}

		// 插入新规则
		for _, rule := range rules {
			sql := `
				INSERT INTO redirect_rules (
					short_link_id, name, description, priority, type, target_url,
					device, start_time, end_time, countries, provinces, cities,
					percentage, max_visits, created_at, updated_at
				) VALUES (
					?, ?, ?, ?, ?, ?,
					?, ?, ?, ?::text[], ?::text[], ?::text[],
					?, ?, ?, ?
				)`

			err := tx.Exec(sql,
				shortLinkID, rule.Name, rule.Description, rule.Priority, rule.Type, rule.TargetURL,
				rule.Device, rule.StartTime, rule.EndTime, pq.Array(rule.Countries), pq.Array(rule.Provinces), pq.Array(rule.Cities),
				rule.Percentage, rule.MaxVisits, rule.CreatedAt, rule.UpdatedAt,
			).Error

			if err != nil {
				return fmt.Errorf("failed to create rule: %w", err)
			}
		}

		// 删除规则缓存
		ctx := context.Background()
		if err := r.redis.Del(ctx, r.getRulesCacheKey(shortLinkID)).Err(); err != nil {
			fmt.Printf("Failed to delete rules cache: %v\n", err)
		}

		return nil
	})
}

// ListClickLogs 获取访问记录列表
func (r *ShortLinkRepository) ListClickLogs(shortLinkID uint, query *domain.ClickLogQuery) (*domain.PaginatedClickLogs, error) {
	fmt.Printf("[ListClickLogs] Starting to get click logs for short link %d with query: %+v\n", shortLinkID, query)

	var total int64
	var logs []domain.ClickLog

	// 构建查询
	db := r.db.Table("click_logs").Where("short_link_id = ?", shortLinkID)

	// 应用过滤条件
	if query.Filter != nil {
		if query.Filter.StartTime != nil {
			db = db.Where("created_at >= ?", query.Filter.StartTime)
		}
		if query.Filter.EndTime != nil {
			db = db.Where("created_at <= ?", query.Filter.EndTime)
		}
		if query.Filter.IP != nil {
			db = db.Where("ip = ?", *query.Filter.IP)
		}
		if query.Filter.Country != nil {
			db = db.Where("country = ?", *query.Filter.Country)
		}
		if query.Filter.Device != nil {
			db = db.Where("device = ?", *query.Filter.Device)
		}
		if query.Filter.RuleID != nil {
			db = db.Where("rule_id = ?", *query.Filter.RuleID)
		}
	}

	// 获取总记录数
	if err := db.Count(&total).Error; err != nil {
		fmt.Printf("[ListClickLogs] Failed to get total count: %v\n", err)
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// 计算总页数
	totalPages := (int(total) + query.PageSize - 1) / query.PageSize

	// 应用排序
	if query.Sort != nil && query.Sort.Field != "" {
		direction := "DESC"
		if query.Sort.Direction == domain.SortAsc {
			direction = "ASC"
		}
		db = db.Order(fmt.Sprintf("%s %s", query.Sort.Field, direction))
	} else {
		// 默认按创建时间降序
		db = db.Order("created_at DESC")
	}

	// 应用分页
	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Find(&logs).Error; err != nil {
		fmt.Printf("[ListClickLogs] Failed to get logs: %v\n", err)
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	fmt.Printf("[ListClickLogs] Successfully got %d logs\n", len(logs))

	return &domain.PaginatedClickLogs{
		Total:       total,
		TotalPages:  totalPages,
		CurrentPage: query.Page,
		PageSize:    query.PageSize,
		Data:        logs,
	}, nil
}
