package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

// RateLimiter 限流中间件
func RateLimiter() gin.HandlerFunc {
	// 从配置中获取限流参数
	requests := viper.GetInt("ratelimit.requests")
	duration := viper.GetDuration("ratelimit.duration")

	// 创建限流器
	limiter := rate.NewLimiter(rate.Every(duration/time.Duration(requests)), requests)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(429, gin.H{
				"error": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
