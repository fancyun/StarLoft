package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetAPISecret 获取API密钥详情(需要二次验证)
func (h *UserHandler) GetAPISecret(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// 需要提供短信验证码进行二次验证
	var req struct {
		SMSCode string `json:"sms_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	// 验证短信验证码
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "user not found",
		})
		return
	}

	valid, err := h.smsService.VerifyCode(user.Phone, req.SMSCode)
	if err != nil || !valid {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "短信验证码错误或已过期",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"api_key":    user.APIKey,
			"api_secret": user.APISecret,
		},
	})
}
