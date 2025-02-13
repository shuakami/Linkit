package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"linkit/internal/domain"
	"linkit/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// ShortLinkHandler 处理短链接相关的HTTP请求
type ShortLinkHandler struct {
	useCase domain.ShortLinkUseCase
}

// NewShortLinkHandler 创建短链接处理器
func NewShortLinkHandler(useCase domain.ShortLinkUseCase) *ShortLinkHandler {
	return &ShortLinkHandler{
		useCase: useCase,
	}
}

// Register 注册API路由
func (h *ShortLinkHandler) Register(r *gin.RouterGroup) {
	r.GET("/links", h.List) // 获取短链接列表
	r.POST("/links", h.Create)
	r.GET("/links/:code", h.Get)
	r.DELETE("/links/:code", h.Delete)
	r.PUT("/links/:code", h.Update)             // 新增: 更新短链接
	r.GET("/links/:code/logs", h.ListClickLogs) // 新增：获取访问记录列表

	// 规则相关路由
	r.POST("/links/:code/rules", h.CreateRule)
	r.GET("/links/:code/rules", h.GetRules)
	r.PUT("/links/:code/rules/:ruleId", h.UpdateRule)
	r.DELETE("/links/:code/rules/:ruleId", h.DeleteRule)
	r.PUT("/links/:code/rules", h.UpdateRules) // 新增: 批量更新规则
}

// RegisterRoot 注册根路由
func (h *ShortLinkHandler) RegisterRoot(r *gin.Engine) {
	// 注册重定向路由
	r.GET("/:code", h.Redirect)
}

// validateCode 验证短码
func (h *ShortLinkHandler) validateCode(code string) error {
	if code == "" {
		return fmt.Errorf("short code is required")
	}
	if len(code) > 16 {
		return fmt.Errorf("short code too long")
	}
	return nil
}

// handleError 统一错误处理
func (h *ShortLinkHandler) handleError(c *gin.Context, err error) {
	switch err {
	case domain.ErrInvalidURL:
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400001,
			"message": "无效的URL格式",
			"details": "请检查URL是否正确，必须是以http://或https://开头的完整URL",
		})
	case domain.ErrCustomCodeExists:
		c.JSON(http.StatusConflict, gin.H{
			"code":    409001,
			"message": "自定义短码已被使用",
			"details": "请尝试使用其他短码，或让系统自动生成短码",
		})
	case domain.ErrInvalidCustomCode:
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400002,
			"message": "无效的自定义短码",
			"details": "短码只能包含字母、数字、下划线和中划线，长度在4-16个字符之间",
		})
	case domain.ErrShortLinkNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404001,
			"message": "短链接不存在",
			"details": "请检查短码是否正确，或者该链接可能已被删除",
		})
	case domain.ErrShortLinkExpired:
		c.JSON(http.StatusGone, gin.H{
			"code":    410001,
			"message": "短链接已过期",
			"details": "该链接已超过设定的有效期，无法访问",
		})
	case domain.ErrRateLimitExceeded:
		c.JSON(http.StatusTooManyRequests, gin.H{
			"code":    429001,
			"message": "请求频率超限",
			"details": "请稍后再试",
		})
	default:
		// 检查是否包含特定错误信息
		if strings.Contains(err.Error(), "failed to check custom code") {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400003,
				"message": "短码验证失败",
				"details": "请检查短码格式是否正确",
			})
			return
		}

		// 其他未预期的错误
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500001,
			"message": "服务器内部错误",
			"details": "请稍后重试，如果问题持续存在请联系管理员",
		})
	}
}

// detectDevice 检测设备类型
func (h *ShortLinkHandler) detectDevice(userAgent string) domain.DeviceType {
	ua := strings.ToLower(userAgent)
	if strings.Contains(ua, "mobile") {
		return domain.DeviceMobile
	}
	if strings.Contains(ua, "tablet") {
		return domain.DeviceTablet
	}
	return domain.DeviceDesktop
}

// Create 创建短链接
func (h *ShortLinkHandler) Create(c *gin.Context) {
	var input domain.CreateShortLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数", "details": err.Error()})
		return
	}

	// 验证过期时间
	if !input.ExpiresAt.IsZero() && input.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "过期时间不能早于当前时间"})
		return
	}

	// 设置默认跳转类型
	if input.DefaultRedirect == 0 {
		input.DefaultRedirect = domain.RedirectPermanent
	}

	shortLink, err := h.useCase.Create(&input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, shortLink)
}

// Get 获取短链接信息
func (h *ShortLinkHandler) Get(c *gin.Context) {
	code := c.Param("code")
	if err := h.validateCode(code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的短码", "details": err.Error()})
		return
	}

	shortLink, err := h.useCase.Get(code)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, shortLink)
}

// Delete 删除短链接
func (h *ShortLinkHandler) Delete(c *gin.Context) {
	code := c.Param("code")
	if err := h.validateCode(code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的短码", "details": err.Error()})
		return
	}

	if err := h.useCase.Delete(code); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// Redirect 重定向到原始URL
func (h *ShortLinkHandler) Redirect(c *gin.Context) {
	code := c.Param("code")
	if err := h.validateCode(code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的短码", "details": err.Error()})
		return
	}

	// 获取IP地区
	clientIP := c.ClientIP()
	region := utils.GetIPRegion(clientIP)

	clickLog := &domain.ClickLog{
		IP:        clientIP,
		UserAgent: c.Request.UserAgent(),
		Referer:   c.Request.Referer(),
		Device:    h.detectDevice(c.Request.UserAgent()),
		Country:   region.Country,
		CreatedAt: time.Now(),
	}

	url, redirectType, err := h.useCase.Redirect(code, clickLog)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrShortLinkNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404001,
				"message": "短链接不存在",
				"details": "请检查短链接是否正确",
			})
		case errors.Is(err, domain.ErrShortLinkExpired):
			c.JSON(http.StatusGone, gin.H{
				"code":    410001,
				"message": "短链接已过期",
				"details": "该链接已超过设定的有效期，无法访问",
			})
		case errors.Is(err, domain.ErrMaxVisitsReached):
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403001,
				"message": "访问次数已达上限",
				"details": "该短链接的访问次数已达到限制,无法继续访问",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500001,
				"message": "服务器内部错误",
				"details": "请稍后重试，如果问题持续存在请联系管理员",
			})
		}
		return
	}

	// 根据规则设置不同的状态码
	var statusCode int
	switch redirectType {
	case domain.RedirectPermanent:
		statusCode = http.StatusMovedPermanently // 301
	case domain.RedirectTemporary:
		statusCode = http.StatusFound // 302
	case domain.RedirectTemporaryKeepMethod:
		statusCode = http.StatusTemporaryRedirect // 307
	case domain.RedirectPermanentKeepMethod:
		statusCode = http.StatusPermanentRedirect // 308
	default:
		statusCode = http.StatusMovedPermanently // 默认301
	}

	c.Redirect(statusCode, url)
}

// CreateRule 创建跳转规则
func (h *ShortLinkHandler) CreateRule(c *gin.Context) {
	code := c.Param("code")
	if err := h.validateCode(code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的短码", "details": err.Error()})
		return
	}

	// 先获取短链接信息
	shortLink, err := h.useCase.Get(code)
	if err != nil {
		h.handleError(c, err)
		return
	}

	var input domain.CreateRuleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数", "details": err.Error()})
		return
	}

	// 设置ShortLinkID
	input.ShortLinkID = shortLink.ID

	// 验证时间范围
	if input.StartTime != nil && input.EndTime != nil && input.EndTime.Before(*input.StartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结束时间不能早于开始时间"})
		return
	}

	rule, err := h.useCase.CreateRule(&input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// GetRules 获取短链接的所有规则
func (h *ShortLinkHandler) GetRules(c *gin.Context) {
	code := c.Param("code")
	if err := h.validateCode(code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的短码", "details": err.Error()})
		return
	}

	shortLink, err := h.useCase.Get(code)
	if err != nil {
		h.handleError(c, err)
		return
	}

	rules, err := h.useCase.GetRules(shortLink.ID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, rules)
}

// UpdateRule 更新跳转规则
func (h *ShortLinkHandler) UpdateRule(c *gin.Context) {
	code := c.Param("code")
	if err := h.validateCode(code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的短码", "details": err.Error()})
		return
	}

	ruleIDStr := c.Param("ruleId")
	ruleID, err := strconv.ParseUint(ruleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则ID", "details": err.Error()})
		return
	}

	var input domain.CreateRuleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数", "details": err.Error()})
		return
	}

	// 验证时间范围
	if input.StartTime != nil && input.EndTime != nil && input.EndTime.Before(*input.StartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结束时间不能早于开始时间"})
		return
	}

	rule, err := h.useCase.UpdateRule(uint(ruleID), &input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, rule)
}

// DeleteRule 删除跳转规则
func (h *ShortLinkHandler) DeleteRule(c *gin.Context) {
	code := c.Param("code")
	if err := h.validateCode(code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的短码", "details": err.Error()})
		return
	}

	ruleIDStr := c.Param("ruleId")
	ruleID, err := strconv.ParseUint(ruleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则ID", "details": err.Error()})
		return
	}

	if err := h.useCase.DeleteRule(uint(ruleID)); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// List 获取短链接列表
func (h *ShortLinkHandler) List(c *gin.Context) {
	// 解析分页参数
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400004,
			"message": "无效的页码",
			"details": "页码必须是大于0的整数",
		})
		return
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400005,
			"message": "无效的每页数量",
			"details": "每页数量必须是1-100之间的整数",
		})
		return
	}

	// 构建查询对象
	query := &domain.PaginationQuery{
		Page:     page,
		PageSize: pageSize,
	}

	// 解析过滤条件
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			uid := uint(userID)
			query.Filter = &domain.ShortLinkFilter{
				UserID: &uid,
			}
		}
	}

	if isExpiredStr := c.Query("is_expired"); isExpiredStr != "" {
		isExpired := isExpiredStr == "true"
		if query.Filter == nil {
			query.Filter = &domain.ShortLinkFilter{}
		}
		query.Filter.IsExpired = &isExpired
	}

	if minClicksStr := c.Query("min_clicks"); minClicksStr != "" {
		if minClicks, err := strconv.ParseUint(minClicksStr, 10, 64); err == nil {
			if query.Filter == nil {
				query.Filter = &domain.ShortLinkFilter{}
			}
			query.Filter.MinClicks = &minClicks
		}
	}

	if maxClicksStr := c.Query("max_clicks"); maxClicksStr != "" {
		if maxClicks, err := strconv.ParseUint(maxClicksStr, 10, 64); err == nil {
			if query.Filter == nil {
				query.Filter = &domain.ShortLinkFilter{}
			}
			query.Filter.MaxClicks = &maxClicks
		}
	}

	// 解析时间范围
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			if query.Filter == nil {
				query.Filter = &domain.ShortLinkFilter{}
			}
			query.Filter.StartTime = &startTime
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			if query.Filter == nil {
				query.Filter = &domain.ShortLinkFilter{}
			}
			query.Filter.EndTime = &endTime
		}
	}

	// 解析排序参数
	if sortField := c.Query("sort_field"); sortField != "" {
		direction := domain.SortDesc
		if c.Query("sort_direction") == "asc" {
			direction = domain.SortAsc
		}
		query.Sort = &domain.ShortLinkSort{
			Field:     sortField,
			Direction: direction,
		}
	}

	// 调用usecase层获取数据
	result, err := h.useCase.List(query)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// Update 更新短链接
func (h *ShortLinkHandler) Update(c *gin.Context) {
	code := c.Param("code")
	if err := h.validateCode(code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的短码", "details": err.Error()})
		return
	}

	var input domain.UpdateShortLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数", "details": err.Error()})
		return
	}

	// 验证过期时间
	if input.ExpiresAt != nil && !input.ExpiresAt.IsZero() && input.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "过期时间不能早于当前时间"})
		return
	}

	shortLink, err := h.useCase.Update(code, &input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, shortLink)
}

// UpdateRules 批量更新规则
func (h *ShortLinkHandler) UpdateRules(c *gin.Context) {
	code := c.Param("code")
	if err := h.validateCode(code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的短码", "details": err.Error()})
		return
	}

	// 先获取短链接信息
	shortLink, err := h.useCase.Get(code)
	if err != nil {
		h.handleError(c, err)
		return
	}

	var inputs []domain.CreateRuleInput
	if err := c.ShouldBindJSON(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数", "details": err.Error()})
		return
	}

	// 验证规则
	for _, input := range inputs {
		if input.StartTime != nil && input.EndTime != nil && input.EndTime.Before(*input.StartTime) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "结束时间不能早于开始时间"})
			return
		}
	}

	rules, err := h.useCase.UpdateRules(shortLink.ID, inputs)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, rules)
}

// ListClickLogs 获取访问记录列表
func (h *ShortLinkHandler) ListClickLogs(c *gin.Context) {
	code := c.Param("code")
	if err := h.validateCode(code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400004,
			"message": "无效的短码",
			"details": err.Error(),
		})
		return
	}

	// 解析分页参数
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400004,
			"message": "无效的页码",
			"details": "页码必须是大于0的整数",
		})
		return
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400005,
			"message": "无效的每页数量",
			"details": "每页数量必须是1-100之间的整数",
		})
		return
	}

	// 构建查询对象
	query := &domain.ClickLogQuery{
		Page:     page,
		PageSize: pageSize,
	}

	// 解析过滤条件
	filter := &domain.ClickLogFilter{}
	hasFilter := false

	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filter.StartTime = &startTime
			hasFilter = true
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filter.EndTime = &endTime
			hasFilter = true
		}
	}

	if ip := c.Query("ip"); ip != "" {
		filter.IP = &ip
		hasFilter = true
	}

	if country := c.Query("country"); country != "" {
		filter.Country = &country
		hasFilter = true
	}

	if deviceStr := c.Query("device"); deviceStr != "" {
		if device, err := strconv.Atoi(deviceStr); err == nil {
			deviceType := domain.DeviceType(device)
			filter.Device = &deviceType
			hasFilter = true
		}
	}

	if ruleIDStr := c.Query("rule_id"); ruleIDStr != "" {
		if ruleID, err := strconv.ParseUint(ruleIDStr, 10, 32); err == nil {
			rid := uint(ruleID)
			filter.RuleID = &rid
			hasFilter = true
		}
	}

	if hasFilter {
		query.Filter = filter
	}

	// 解析排序条件
	if field := c.Query("sort_field"); field != "" {
		direction := domain.SortDesc
		if c.Query("sort_direction") == "asc" {
			direction = domain.SortAsc
		}
		query.Sort = &domain.ClickLogSort{
			Field:     field,
			Direction: direction,
		}
	}

	// 获取访问记录
	logs, err := h.useCase.ListClickLogs(code, query)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, logs)
}
