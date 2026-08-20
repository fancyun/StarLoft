package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"starloftrpa/internal/config"
)

type PublicHandler struct {
	config *config.Config
}

func NewPublicHandler(cfg *config.Config) *PublicHandler {
	return &PublicHandler{
		config: cfg,
	}
}

// GetPublicConfig 获取公开配置（无需登录）
func (h *PublicHandler) GetPublicConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"captcha_app_id": h.config.Tencent.Captcha.CaptchaAppID,
			// 可以添加其他前端需要的公开配置
		},
	})
}
