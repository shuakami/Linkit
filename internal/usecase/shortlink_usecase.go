package usecase

import (
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"linkit/internal/domain"
	"linkit/pkg/utils"

	"github.com/spf13/viper"
)

// ShortLinkUseCase 实现短链接用例接口
type ShortLinkUseCase struct {
	repo domain.ShortLinkRepository
}

// NewShortLinkUseCase 创建短链接用例实例
func NewShortLinkUseCase(repo domain.ShortLinkRepository) domain.ShortLinkUseCase {
	return &ShortLinkUseCase{
		repo: repo,
	}
}

// validateURL 验证URL的安全性
func (u *ShortLinkUseCase) validateURL(longURL string) error {
	// 解析URL
	parsedURL, err := url.Parse(longURL)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidURL, err)
	}

	// 检查URL scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("%w: invalid scheme", domain.ErrInvalidURL)
	}

	// 检查URL长度
	if len(longURL) > 2048 {
		return fmt.Errorf("%w: url too long", domain.ErrInvalidURL)
	}

	return nil
}

// matchRule 检查规则是否匹配
func (u *ShortLinkUseCase) matchRule(rule *domain.RedirectRule, clickLog *domain.ClickLog) bool {
	fmt.Printf("[规则] 正在匹配: %s\n", rule.Name)
	fmt.Printf("      访问者: %s (%s)\n", clickLog.IP, clickLog.Country)

	// 检查设备类型
	if rule.Device != domain.DeviceAll && rule.Device != clickLog.Device {
		fmt.Printf("      ✗ 设备不匹配\n")
		return false
	}

	// 检查时间范围
	now := time.Now()
	if rule.StartTime != nil && now.Before(*rule.StartTime) {
		fmt.Printf("      ✗ 未到开始时间\n")
		return false
	}
	if rule.EndTime != nil && now.After(*rule.EndTime) {
		fmt.Printf("      ✗ 已过结束时间\n")
		return false
	}

	// 检查地区匹配
	if len(rule.Countries) > 0 || len(rule.Provinces) > 0 || len(rule.Cities) > 0 {
		region := utils.GetIPRegion(clickLog.IP)

		// 检查国家匹配
		if len(rule.Countries) > 0 {
			matched := false
			for _, country := range rule.Countries {
				if country == region.Country {
					matched = true
					break
				}
			}
			if !matched {
				fmt.Printf("      ✗ 国家不匹配: %s\n", region.Country)
				return false
			}
		}

		// 检查省份匹配
		if len(rule.Provinces) > 0 {
			matched := false
			for _, province := range rule.Provinces {
				if province == region.Province {
					matched = true
					break
				}
			}
			if !matched {
				fmt.Printf("      ✗ 省份不匹配: %s\n", region.Province)
				return false
			}
		}

		// 检查城市匹配
		if len(rule.Cities) > 0 {
			matched := false
			for _, city := range rule.Cities {
				if city == region.City {
					matched = true
					break
				}
			}
			if !matched {
				fmt.Printf("      ✗ 城市不匹配: %s\n", region.City)
				return false
			}
		}
	}

	// 检查访问次数
	if rule.MaxVisits != nil && *rule.MaxVisits <= 0 {
		fmt.Printf("      ✗ 已达访问上限\n")
		return false
	}

	// 检查百分比
	if rule.Percentage != nil {
		randNum := rand.Intn(100) + 1
		if randNum > *rule.Percentage {
			fmt.Printf("      ✗ 未通过流量限制\n")
			return false
		}
	}

	fmt.Printf("      ✓ 规则匹配成功\n")
	return true
}

// Create 创建短链接
func (u *ShortLinkUseCase) Create(input *domain.CreateShortLinkInput) (*domain.ShortLink, error) {
	// 验证URL安全性
	if err := u.validateURL(input.LongURL); err != nil {
		return nil, err
	}

	// 验证自定义短码
	if input.CustomCode != "" {
		if !utils.ValidateCustomCode(input.CustomCode) {
			return nil, fmt.Errorf("%w: invalid format", domain.ErrInvalidCustomCode)
		}
		// 检查自定义短码是否已存在
		if _, err := u.repo.GetByCode(input.CustomCode); err == nil {
			return nil, domain.ErrCustomCodeExists
		} else if err != domain.ErrShortLinkNotFound {
			return nil, fmt.Errorf("failed to check custom code: %w", err)
		}
	}

	// 生成短码
	var shortCode string
	var err error
	if input.CustomCode != "" {
		shortCode = input.CustomCode
	} else {
		shortCode, err = utils.GenerateShortCode(viper.GetInt("shortlink.length"))
		if err != nil {
			return nil, fmt.Errorf("failed to generate short code: %w", err)
		}
	}

	// 设置过期时间
	expiresAt := input.ExpiresAt
	if input.NeverExpire {
		// 如果设置为永不过期,将过期时间设置为100年后
		expiresAt = time.Now().AddDate(100, 0, 0)
	} else if expiresAt.IsZero() {
		duration, err := time.ParseDuration(viper.GetString("shortlink.expiration"))
		if err != nil {
			return nil, fmt.Errorf("failed to parse expiration duration: %w", err)
		}
		expiresAt = time.Now().Add(duration)
	}

	// 创建短链接
	shortLink := &domain.ShortLink{
		ShortCode:       shortCode,
		LongURL:         input.LongURL,
		UserID:          input.UserID,
		DefaultRedirect: input.DefaultRedirect,
		ExpiresAt:       expiresAt,
		NeverExpire:     input.NeverExpire,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := u.repo.Create(shortLink); err != nil {
		return nil, fmt.Errorf("failed to create short link: %w", err)
	}

	return shortLink, nil
}

// Get 获取短链接信息
func (u *ShortLinkUseCase) Get(code string) (*domain.ShortLink, error) {
	shortLink, err := u.repo.GetByCode(code)
	if err != nil {
		if err == domain.ErrShortLinkNotFound {
			return nil, domain.ErrShortLinkNotFound
		}
		return nil, fmt.Errorf("failed to get short link: %w", err)
	}

	// 检查是否过期
	if time.Now().After(shortLink.ExpiresAt) {
		return nil, domain.ErrShortLinkExpired
	}

	// 加载规则
	rules, err := u.repo.GetRules(shortLink.ID)
	if err != nil {
		fmt.Printf("failed to get rules for short link %d: %v\n", shortLink.ID, err)
	} else {
		shortLink.Rules = rules
	}

	return shortLink, nil
}

// Redirect 重定向并记录点击
func (u *ShortLinkUseCase) Redirect(code string, clickLog *domain.ClickLog) (string, domain.RedirectType, error) {
	fmt.Printf("[访问] 短链接: %s\n", code)
	fmt.Printf("      来源: %s (%s)\n", clickLog.IP, clickLog.Country)

	shortLink, err := u.Get(code)
	if err != nil {
		fmt.Printf("      ✗ 获取失败: %v\n", err)
		return "", 0, err
	}

	// 检查访问次数限制
	if shortLink.MaxVisits != nil && shortLink.Clicks >= *shortLink.MaxVisits {
		fmt.Printf("      ✗ 已达到最大访问次数限制\n")
		return "", 0, domain.ErrMaxVisitsReached
	}

	// 获取所有规则
	rules, err := u.repo.GetRules(shortLink.ID)
	if err != nil {
		fmt.Printf("      ✗ 获取规则失败\n")
		return "", 0, fmt.Errorf("failed to get rules: %w", err)
	}

	// 按优先级排序并匹配规则
	var matchedRule *domain.RedirectRule
	for _, rule := range rules {
		if u.matchRule(&rule, clickLog) {
			matchedRule = &rule
			break
		}
	}

	// 设置重定向URL和类型
	targetURL := shortLink.LongURL
	redirectType := shortLink.DefaultRedirect

	if matchedRule != nil {
		if matchedRule.TargetURL != "" {
			targetURL = matchedRule.TargetURL
		}
		redirectType = matchedRule.Type
		clickLog.RuleID = &matchedRule.ID
		fmt.Printf("      ✓ 使用规则: %s\n", matchedRule.Name)
	} else {
		fmt.Printf("      ✓ 使用默认规则\n")
	}

	fmt.Printf("      → 目标: %s\n", targetURL)

	// 增加点击次数
	if err := u.repo.IncrementClicks(code); err != nil {
		fmt.Printf("      ✗ 更新点击失败\n")
		return "", 0, fmt.Errorf("failed to increment clicks: %w", err)
	}

	// 记录点击日志
	clickLog.ShortLinkID = shortLink.ID
	if err := u.repo.LogClick(clickLog); err != nil {
		fmt.Printf("      ✗ 记录日志失败\n")
		return "", 0, fmt.Errorf("failed to log click: %w", err)
	}

	return targetURL, redirectType, nil
}

// Delete 删除短链接
func (u *ShortLinkUseCase) Delete(code string) error {
	// 检查短链接是否存在
	if _, err := u.repo.GetByCode(code); err != nil {
		return fmt.Errorf("failed to check short link existence: %w", err)
	}

	if err := u.repo.Delete(code); err != nil {
		return fmt.Errorf("failed to delete short link: %w", err)
	}

	return nil
}

// CreateRule 创建跳转规则
func (u *ShortLinkUseCase) CreateRule(input *domain.CreateRuleInput) (*domain.RedirectRule, error) {
	rule := &domain.RedirectRule{
		ShortLinkID: input.ShortLinkID,
		Name:        input.Name,
		Description: input.Description,
		Priority:    input.Priority,
		Type:        input.Type,
		TargetURL:   input.TargetURL,
		Device:      input.Device,
		StartTime:   input.StartTime,
		EndTime:     input.EndTime,
		Countries:   input.Countries,
		Percentage:  input.Percentage,
		MaxVisits:   input.MaxVisits,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := u.repo.CreateRule(rule); err != nil {
		return nil, fmt.Errorf("failed to create rule: %w", err)
	}

	return rule, nil
}

// UpdateRule 更新跳转规则
func (u *ShortLinkUseCase) UpdateRule(ruleID uint, input *domain.CreateRuleInput) (*domain.RedirectRule, error) {
	rule := &domain.RedirectRule{
		ID:          ruleID,
		ShortLinkID: input.ShortLinkID,
		Name:        input.Name,
		Description: input.Description,
		Priority:    input.Priority,
		Type:        input.Type,
		TargetURL:   input.TargetURL,
		Device:      input.Device,
		StartTime:   input.StartTime,
		EndTime:     input.EndTime,
		Countries:   input.Countries,
		Percentage:  input.Percentage,
		MaxVisits:   input.MaxVisits,
		UpdatedAt:   time.Now(),
	}

	if err := u.repo.UpdateRule(rule); err != nil {
		return nil, fmt.Errorf("failed to update rule: %w", err)
	}

	return rule, nil
}

// DeleteRule 删除跳转规则
func (u *ShortLinkUseCase) DeleteRule(ruleID uint) error {
	if err := u.repo.DeleteRule(ruleID); err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}
	return nil
}

// GetRules 获取短链接的所有规则
func (u *ShortLinkUseCase) GetRules(shortLinkID uint) ([]domain.RedirectRule, error) {
	rules, err := u.repo.GetRules(shortLinkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules: %w", err)
	}
	return rules, nil
}

// List 获取短链接列表
func (u *ShortLinkUseCase) List(query *domain.PaginationQuery) (*domain.PaginatedShortLinks, error) {
	// 验证排序字段
	if query.Sort != nil && query.Sort.Field != "" {
		// 检查排序字段是否合法
		validFields := map[string]bool{
			"created_at": true,
			"expires_at": true,
			"clicks":     true,
			"short_code": true,
		}
		if !validFields[query.Sort.Field] {
			return nil, fmt.Errorf("invalid sort field: %s", query.Sort.Field)
		}
	}

	// 调用repository层获取数据
	result, err := u.repo.List(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list short links: %w", err)
	}

	// 对于每个短链接，检查是否需要加载规则
	for i := range result.Data {
		rules, err := u.repo.GetRules(result.Data[i].ID)
		if err != nil {
			fmt.Printf("failed to get rules for short link %d: %v\n", result.Data[i].ID, err)
			continue
		}
		result.Data[i].Rules = rules
	}

	return result, nil
}

// Update 更新短链接
func (u *ShortLinkUseCase) Update(code string, input *domain.UpdateShortLinkInput) (*domain.ShortLink, error) {
	// 获取现有短链接
	link, err := u.repo.GetByCode(code)
	if err != nil {
		return nil, fmt.Errorf("failed to get short link: %w", err)
	}

	// 更新字段
	if input.LongURL != nil {
		// 验证新URL
		if err := u.validateURL(*input.LongURL); err != nil {
			return nil, err
		}
		link.LongURL = *input.LongURL
	}

	if input.MaxVisits != nil {
		link.MaxVisits = input.MaxVisits
	}

	if input.ExpiresAt != nil {
		if !input.ExpiresAt.IsZero() && input.ExpiresAt.Before(time.Now()) {
			return nil, fmt.Errorf("expiration time cannot be in the past")
		}
		link.ExpiresAt = *input.ExpiresAt
	}

	if input.NeverExpire != nil {
		link.NeverExpire = *input.NeverExpire
		// 如果设置为永不过期,将过期时间设置为100年后
		if *input.NeverExpire {
			future := time.Now().AddDate(100, 0, 0)
			link.ExpiresAt = future
		}
	}

	if input.DefaultRedirect != nil {
		link.DefaultRedirect = *input.DefaultRedirect
	}

	// 更新时间
	link.UpdatedAt = time.Now()

	// 保存更新
	if err := u.repo.Update(link); err != nil {
		return nil, fmt.Errorf("failed to update short link: %w", err)
	}

	return link, nil
}

// UpdateRules 批量更新规则
func (u *ShortLinkUseCase) UpdateRules(shortLinkID uint, inputs []domain.CreateRuleInput) ([]domain.RedirectRule, error) {
	rules := make([]domain.RedirectRule, len(inputs))
	for i, input := range inputs {
		rule := &domain.RedirectRule{
			ShortLinkID: shortLinkID,
			Name:        input.Name,
			Description: input.Description,
			Priority:    input.Priority,
			Type:        input.Type,
			TargetURL:   input.TargetURL,
			Device:      input.Device,
			StartTime:   input.StartTime,
			EndTime:     input.EndTime,
			Countries:   input.Countries,
			Percentage:  input.Percentage,
			MaxVisits:   input.MaxVisits,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		rules[i] = *rule
	}

	if err := u.repo.UpdateRules(shortLinkID, rules); err != nil {
		return nil, fmt.Errorf("failed to update rules: %w", err)
	}

	return rules, nil
}

// ListClickLogs 获取访问记录列表
func (u *ShortLinkUseCase) ListClickLogs(code string, query *domain.ClickLogQuery) (*domain.PaginatedClickLogs, error) {
	// 先获取短链接信息
	shortLink, err := u.repo.GetByCode(code)
	if err != nil {
		if err == domain.ErrShortLinkNotFound {
			return nil, domain.ErrShortLinkNotFound
		}
		return nil, fmt.Errorf("failed to get short link: %w", err)
	}

	// 获取访问记录
	logs, err := u.repo.ListClickLogs(shortLink.ID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get click logs: %w", err)
	}

	return logs, nil
}
