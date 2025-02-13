package handler

import (
	"errors"
	"linkit/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleError(c *gin.Context, err error) {
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
			"details": "该短链接已过期,无法访问",
		})
	case errors.Is(err, domain.ErrMaxVisitsReached):
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403001,
			"message": "访问次数已达上限",
			"details": "该短链接的访问次数已达到限制,无法继续访问",
		})
	case errors.Is(err, domain.ErrInvalidURL):
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400001,
			"message": "无效的URL",
			"details": err.Error(),
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500001,
			"message": "服务器错误",
			"details": err.Error(),
		})
	}
}
