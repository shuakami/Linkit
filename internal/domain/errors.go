package domain

import "errors"

var (
	// ErrShortLinkNotFound 表示短链接不存在
	ErrShortLinkNotFound = errors.New("short link not found")

	// ErrShortLinkExpired 表示短链接已过期
	ErrShortLinkExpired = errors.New("short link expired")

	// ErrCustomCodeExists 表示自定义短码已存在
	ErrCustomCodeExists = errors.New("custom code already exists")

	// ErrInvalidCustomCode 表示无效的自定义短码
	ErrInvalidCustomCode = errors.New("invalid custom code")

	// ErrInvalidURL 表示无效的URL
	ErrInvalidURL = errors.New("invalid url")

	// ErrDatabaseError 表示数据库错误
	ErrDatabaseError = errors.New("database error")

	// ErrCacheError 表示缓存错误
	ErrCacheError = errors.New("cache error")

	// ErrRateLimitExceeded 表示超出速率限制
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrMaxVisitsReached 表示访问次数达到上限
	ErrMaxVisitsReached = errors.New("maximum visits limit reached")
)
