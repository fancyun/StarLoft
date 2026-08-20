package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"starloftrpa/internal/repository"
	"starloftrpa/internal/utils"
)

// AuthMiddleware JWT authentication middleware
func AuthMiddleware(jwtManager *utils.JWTManager, userType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization from Header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "missing authorization header",
			})
			c.Abort()
			return
		}

		// Bearer Token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate Token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "invalid or expired token",
			})
			c.Abort()
			return
		}

		// Check user type
		if userType != "" && claims.UserType != userType {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "permission denied",
			})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("user_id", claims.UserID)
		c.Set("phone", claims.Phone)
		c.Set("user_type", claims.UserType)

		c.Next()
	}
}

// APIKeyMiddleware API Key authentication middleware
func APIKeyMiddleware(userRepo *repository.UserRepository, signMgr *utils.SignatureManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API Key and signature info from Header
		apiKey := c.GetHeader("X-Api-Key")
		sign := c.GetHeader("X-Sign")
		signVersion := c.GetHeader("X-Sign-Version")
		timestamp := c.GetHeader("X-Timestamp")

		if apiKey == "" || sign == "" || signVersion == "" || timestamp == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "missing required headers",
			})
			c.Abort()
			return
		}

		// Query user
		user, err := userRepo.GetByAPIKey(apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "invalid api key",
			})
			c.Abort()
			return
		}

		// Check user status
		if user.Status == 0 {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "user disabled",
			})
			c.Abort()
			return
		}

		// Verify timestamp (prevent replay, allow ±5 minutes)
		var ts int64
		_, err = fmt.Sscanf(timestamp, "%d", &ts)
		if err != nil || !signMgr.VerifyTimestamp(ts, 300) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "invalid or expired timestamp",
			})
			c.Abort()
			return
		}

		// Read request body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "failed to read request body",
			})
			c.Abort()
			return
		}

		// Restore request body (for subsequent Handler use)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Verify signature
		if signVersion != "hmac_sha256" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "unsupported sign version",
			})
			c.Abort()
			return
		}

		if !signMgr.VerifyHMACSHA256(user.APISecret, string(bodyBytes), sign) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "invalid signature",
			})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("user_id", user.ID)
		c.Set("user", user)

		c.Next()
	}
}
