package http

import (
	"github.com/gin-gonic/gin"
)

// Handler 处理所有HTTP请求的处理器接口
type Handler interface {
	// 注册API路由
	Register(r *gin.RouterGroup)
	// 注册根路由（用于重定向等特殊路由）
	RegisterRoot(r *gin.Engine)
}

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.Engine, handlers ...Handler) {
	// API版本分组
	v1 := r.Group("/api/v1")

	// 注册API处理器
	for _, h := range handlers {
		h.Register(v1)
		h.RegisterRoot(r)
	}

	// 健康检查路由
	r.GET("/health", healthCheck)
}

// healthCheck 健康检查处理函数
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}
