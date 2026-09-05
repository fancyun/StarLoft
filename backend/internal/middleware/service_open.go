package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"starloftrpa/internal/repository"
)

// ServiceOpenMiddleware 服务开通校验：拦截已登录用户对指定服务（serviceCode）的调用，
// 若用户尚未开通该服务则返回 403，提示先开通。
func ServiceOpenMiddleware(userServiceRepo *repository.UserServiceRepository, serviceCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		opened, err := userServiceRepo.IsOpen(userID, serviceCode)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":    500,
				"message": "查询服务开通状态失败",
			})
			return
		}
		if !opened {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":    403,
				"message": "服务未开通，请先完成实名认证并开通服务",
			})
			return
		}
		c.Next()
	}
}
