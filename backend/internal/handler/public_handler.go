package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	qrcode "github.com/skip2/go-qrcode"

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

// GetQRCode 渲染二维码图片（用于微信Native支付扫码展示）
func (h *PublicHandler) GetQRCode(c *gin.Context) {
	data := c.Query("data")
	if data == "" || len(data) > 2048 {
		c.String(http.StatusBadRequest, "invalid data")
		return
	}

	size := 280
	if s := c.Query("size"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n >= 100 && n <= 1000 {
			size = n
		}
	}

	png, err := qrcode.Encode(data, qrcode.Medium, size)
	if err != nil {
		c.String(http.StatusInternalServerError, "qr encode failed")
		return
	}
	c.Data(http.StatusOK, "image/png", png)
}
