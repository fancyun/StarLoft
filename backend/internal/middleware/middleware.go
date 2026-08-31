package middleware

import (
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

// RequestLogger 请求日志中间件(不记录敏感信息)，写入 access.log
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// 不记录敏感路径的详细信息
		if isSensitivePath(path) {
			utils.AccessLogger.Printf("[%s] %s - %d (%v)", method, path, statusCode, latency)
		} else {
			utils.AccessLogger.Printf("[%s] %s - %d (%v) - IP: %s", method, path, statusCode, latency, c.ClientIP())
		}
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
