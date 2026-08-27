package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"starloftrpa/internal/config"
	"starloftrpa/internal/repository"
)

type PublicHandler struct {
	config     *config.Config
	configRepo *repository.SystemConfigRepository
}

func NewPublicHandler(cfg *config.Config, configRepo *repository.SystemConfigRepository) *PublicHandler {
	return &PublicHandler{
		config:     cfg,
		configRepo: configRepo,
	}
}

// GetPublicConfig 获取公开配置（无需登录）
func (h *PublicHandler) GetPublicConfig(c *gin.Context) {
	// 读取平台 KYC 认证单价（系统配置 kyc_price，统一按平台价格扣费）
	kycPrice := 1.00
	if priceStr, err := h.configRepo.GetConfig("kyc_price"); err == nil && priceStr != "" {
		if price, err := strconv.ParseFloat(priceStr, 64); err == nil && price > 0 {
			kycPrice = price
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"captcha_app_id": h.config.Tencent.Captcha.CaptchaAppID,
			"kyc_price":      kycPrice,
		},
	})
}
