package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"starloftrpa/internal/model"
)

// ServiceAccessGuard 校验 API Key 对指定服务的访问资格（与 APIKeyMiddleware 配合，需在其之后挂载）：
//   - 权限范围：密钥 permission 为 "all" 或等于调用服务 serviceCode 才放行
//   - 实名等级：所属用户实名等级须达到 needKYC（如 kyc 服务要求企业实名 needKYC=2）
func ServiceAccessGuard(serviceCode string, needKYC int) gin.HandlerFunc {
	return func(c *gin.Context) {
		credVal, ok := c.Get("api")
		if !ok {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":    403,
				"message": "API Key 权限校验失败",
			})
			return
		}
		cred := credVal.(*model.ApiKey)
		if cred.Permission != "all" && cred.Permission != serviceCode {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":    403,
				"message": "API Key 无权访问该服务",
			})
			return
		}

		userVal, ok := c.Get("user")
		if !ok {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":    403,
				"message": "用户信息获取失败",
			})
			return
		}
		user := userVal.(*model.User)
		if user.IsKYCVerified < needKYC {
			if needKYC == 2 {
				c.AbortWithStatusJSON(http.StatusOK, gin.H{
					"code":    403,
					"message": "该服务需完成企业实名认证后使用",
				})
			} else {
				c.AbortWithStatusJSON(http.StatusOK, gin.H{
					"code":    403,
					"message": "该服务需完成实名认证后使用",
				})
			}
			return
		}

		c.Next()
	}
}