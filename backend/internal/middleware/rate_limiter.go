package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"starloftrpa/internal/redis"
)

const (
	// API 调用频率限制前缀
	APIRateLimitPrefix = "rate:api:"
	// 默认限制：每分钟60次
	DefaultRateLimit = 60
	// 限制时间窗口
	RateLimitWindow = 60 * time.Second
)

// RateLimiter API 频率限制中间件
func RateLimiter(limit int) gin.HandlerFunc {
	if limit <= 0 {
		limit = DefaultRateLimit
	}

	return func(c *gin.Context) {
		// 获取用户标识（优先使用用户ID，其次使用IP）
		userID, exists := c.Get("user_id")
		var identifier string
		if exists {
			identifier = fmt.Sprintf("user:%v", userID)
		} else {
			identifier = fmt.Sprintf("ip:%s", c.ClientIP())
		}

		// 构造 Redis 键
		key := APIRateLimitPrefix + identifier

		// 增加计数
		count, err := redis.Incr(key)
		if err != nil {
			// Redis 错误时不阻止请求，只记录日志
			c.Next()
			return
		}

		// 如果是第一次计数，设置过期时间
		if count == 1 {
			_ = redis.Expire(key, RateLimitWindow)
		}

		// 检查是否超过限制
		if count > int64(limit) {
			// 获取剩余时间
			ttl, _ := redis.TTL(key)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":        429,
				"message":     "请求过于频繁，请稍后再试",
				"retry_after": int(ttl.Seconds()),
			})
			c.Abort()
			return
		}

		// 设置响应头
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-int(count)))

		c.Next()
	}
}

// RateLimiterForUser 针对已登录用户的频率限制
func RateLimiterForUser(limit int) gin.HandlerFunc {
	return RateLimiter(limit)
}

// RateLimiterForIP 针对 IP 的频率限制（用于登录、注册等接口）
func RateLimiterForIP(limit int) gin.HandlerFunc {
	if limit <= 0 {
		limit = DefaultRateLimit
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := APIRateLimitPrefix + "ip:" + ip

		count, err := redis.Incr(key)
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			_ = redis.Expire(key, RateLimitWindow)
		}

		if count > int64(limit) {
			ttl, _ := redis.TTL(key)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":        429,
				"message":     "请求过于频繁，请稍后再试",
				"retry_after": int(ttl.Seconds()),
			})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-int(count)))

		c.Next()
	}
}
