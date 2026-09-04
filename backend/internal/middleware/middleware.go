package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"time"

	"starloftrpa/internal/utils"
)

// CORSMiddleware 跨域资源共享中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 在生产环境应该配置具体的域名，不要使用 *
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Api-Key, X-Sign, X-Sign-Version, X-Timestamp")
			c.Header("Access-Control-Expose-Headers", "Content-Length, X-RateLimit-Limit, X-RateLimit-Remaining")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Recovery 从panic恢复的中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				utils.ErrorLogger.Printf("Panic recovered: %v", err)
				c.JSON(500, gin.H{
					"code":    500,
					"message": "internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// RequestLogger 请求日志中间件（记录全部 HTTP 访问，按类别 + 用户身份写入 access.log）
// 覆盖：用户操作（console）、平台 API 调用（api）、上游回调（callback）、管理操作（admin）。
// 所有访问、调用、操作均在此统一按类别落 log，日志不做自动删除。
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		category := categorizeAccess(path)

		// 识别已登录用户/管理员身份（JWT 中间件写入 user_id）
		identity := "-"
		if uid, ok := c.Get("user_id"); ok {
			if id, ok2 := uid.(int64); ok2 && id > 0 {
				identity = fmt.Sprintf("id=%d", id)
			}
		}

		// 敏感路径（登录/注册）不记录客户端 IP，保护隐私
		ip := c.ClientIP()
		if isSensitivePath(path) {
			ip = "-"
		}

		utils.AccessLogger.Printf("category=%s user=%s [%s] %s - %d (%v) ip=%s", category, identity, method, path, statusCode, latency, ip)
	}
}

// categorizeAccess 依据路径前缀划分访问类别
func categorizeAccess(path string) string {
	switch {
	case strings.HasPrefix(path, "/admin/"):
		return "admin"
	case strings.HasPrefix(path, "/api/callback/"):
		return "callback"
	case strings.HasPrefix(path, "/api/"):
		return "api"
	case strings.HasPrefix(path, "/console/"):
		return "console"
	default:
		return "other"
	}
}

func isSensitivePath(path string) bool {
	sensitivePaths := []string{"/console/login", "/admin/login", "/console/register"}
	for _, sp := range sensitivePaths {
		if path == sp {
			return true
		}
	}
	return false
}
