package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"starloftrpa/internal/service"
	"starloftrpa/internal/utils"
)

var authValidator = utils.NewInputValidator()

type AuthHandler struct {
	authService    *service.AuthService
	balanceService *service.BalanceService
}

func NewAuthHandler(authService *service.AuthService, balanceService *service.BalanceService) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		balanceService: balanceService,
	}
}

// StartAuth 发起认证（API 调用）
func (h *AuthHandler) StartAuth(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// 检查用户是否已实名
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "user not found",
		})
		return
	}
	if user.IsKYCVerified != 1 {
		c.JSON(http.StatusOK, gin.H{
			"code":    403,
			"message": "用户未完成实名认证，请先实名",
		})
		return
	}

	var req struct {
		BizNo        string `json:"biz_no" binding:"required"`
		Name         string `json:"name" binding:"required"`
		IDCard       string `json:"id_card" binding:"required"`
		ReturnURL    string `json:"return_url" binding:"required"`
		NotifyURL    string `json:"notify_url" binding:"required"`
		BizExtraData string `json:"biz_extra_data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	// 验证姓名
	if err := authValidator.ValidateName(req.Name); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid name format",
		})
		return
	}

	// 验证身份证号
	if err := authValidator.ValidateIDCardMinAge(req.IDCard, utils.MinKycAge); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	// 验证URL
	if err := authValidator.ValidateURL(req.ReturnURL); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid return_url format",
		})
		return
	}

	if err := authValidator.ValidateURL(req.NotifyURL); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid notify_url format",
		})
		return
	}

	// 清理可能的XSS
	req.BizExtraData = authValidator.SanitizeString(req.BizExtraData)

	// 发起认证
	result, err := h.authService.StartAuth(
		userID,
		req.Name,
		req.IDCard,
		req.BizNo,
		req.ReturnURL,
		req.NotifyURL,
		req.BizExtraData,
		true,  // 下游调用：平台中转，向上游传平台地址，平台收到后再通知/跳转下游
		false, // 下游 API 业务调用：仍按 kyc_price 从余额扣费
	)
	if err != nil {
		if err == service.ErrInsufficientBalance {
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"message": "insufficient balance",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"platform_biz_no": result.Order.PlatformBizNo,
			"auth_url":        result.AuthURL,
			"expired_time":    result.Order.CreatedAt.Add(15 * time.Minute).Unix(),
			"expired_in":      900,
		},
	})
}

// GetAuthResult 查询认证结果
func (h *AuthHandler) GetAuthResult(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// 检查用户是否已实名
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "user not found",
		})
		return
	}
	if user.IsKYCVerified != 1 {
		c.JSON(http.StatusOK, gin.H{
			"code":    403,
			"message": "用户未完成实名认证，请先实名",
		})
		return
	}

	var req struct {
		BizNo         string `json:"biz_no"`
		PlatformBizNo string `json:"platform_biz_no"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	order, err := h.authService.GetAuthResult(userID, req.BizNo, req.PlatformBizNo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "order not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"platform_biz_no": order.PlatformBizNo,
			"biz_no":          order.BizNo,
			"result_code":     order.ResultCode,
			"result_message":  order.ResultMessage,
			"status":          order.Status,
			"cost":            order.Cost,
		},
	})
}

// GetPublicKycResult 公开查询认证结果（/kyc 中转页调用，按平台流水号查询）
func (h *AuthHandler) GetPublicKycResult(c *gin.Context) {
	bizNo := c.Query("biz_no")
	if bizNo == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "missing biz_no",
		})
		return
	}

	order, err := h.authService.GetPublicOrderResult(bizNo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "order not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"status":         order.Status,
			"result_code":    order.ResultCode,
			"result_message": order.ResultMessage,
			"return_url":     order.ReturnURL,
			"up_token":       order.UpToken,
		},
	})
}

// QueryBalance 查询余额
func (h *AuthHandler) QueryBalance(c *gin.Context) {
	userID := c.GetInt64("user_id")

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "user not found",
		})
		return
	}

	if user.IsKYCVerified != 1 {
		c.JSON(http.StatusOK, gin.H{
			"code":    403,
			"message": "用户未完成实名认证，请先实名",
		})
		return
	}

	// 获取用户的实名认证单价（已从用户表的 kyc_price 字段获取）
	kycPrice := user.KYCPrice

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"balance":   user.Balance,
			"kyc_price": kycPrice,
		},
	})
}

// GetUserAuthStatus 查询用户的 KYC 认证状态（Web 前端调用）
func (h *AuthHandler) GetUserAuthStatus(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// 获取用户信息
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "user not found",
		})
		return
	}

	// 获取最新 kyc_record
	kycRecord, _ := h.authService.GetLatestKycRecord(userID)

	// 提取 platform_user 实名信息
	kycName := ""
	if user.KYCName.Valid {
		kycName = user.KYCName.String
	}
	kycIDCard := ""
	if user.KYCIDCard.Valid {
		kycIDCard = user.KYCIDCard.String
	}

	// 构建响应：包含 platform_user 实名状态 + kyc_record 最新记录
	resp := gin.H{
		"is_kyc_verified": user.IsKYCVerified,
		"kyc_name":        kycName,
		"kyc_id_card":     kycIDCard,
	}

	if kycRecord != nil {
		resp["record_status"] = kycRecord.Status
		resp["record_name"] = kycRecord.Name
		resp["record_id_card"] = kycRecord.IDCard
		resp["record_id"] = kycRecord.ID
	} else {
		resp["record_status"] = -1 // 无记录
	}

	// 如果状态为进行中，返回关联的认证订单 token（用于继续认证）
	if kycRecord != nil && kycRecord.Status == 1 {
		pendingOrder, err := h.authService.GetLatestPendingOrder(userID)
		if err == nil && pendingOrder != nil && pendingOrder.UpToken != "" {
			resp["pending_token"] = pendingOrder.UpToken
		}
		if err == nil && pendingOrder != nil && pendingOrder.PlatformBizNo != "" {
			resp["pending_biz_no"] = pendingOrder.PlatformBizNo
		}
		// 返回下游的 return_url（API 调用方传入的，用于结果页跳转回下游）
		if err == nil && pendingOrder != nil && pendingOrder.ReturnURL != "" {
			resp["return_url"] = pendingOrder.ReturnURL
		}
	}

	// 如果已有结果（状态 2/3），也返回 return_url 用于结果页跳转
	if kycRecord != nil && (kycRecord.Status == 2 || kycRecord.Status == 3) {
		if resp["return_url"] == nil {
			latestOrder, err := h.authService.GetLatestOrder(userID)
			if err == nil && latestOrder != nil && latestOrder.ReturnURL != "" {
				resp["return_url"] = latestOrder.ReturnURL
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    resp,
	})
}

// SyncKycResult 同步上游认证结果（保留兼容；/kyc 页面现统一走公开结果查询接口）
func (h *AuthHandler) SyncKycResult(c *gin.Context) {
	userID := c.GetInt64("user_id")

	order, err := h.authService.SyncOrderByToken(userID)
	if err != nil {
		// 没有进行中的订单，返回当前状态
		kycRecord, _ := h.authService.GetLatestKycRecord(userID)
		if kycRecord != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data": gin.H{
					"synced":        false,
					"record_status": kycRecord.Status,
				},
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
			"data": gin.H{
				"synced":        false,
				"record_status": -1,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"synced":       true,
			"order_status": order.Status,
			"up_token":     order.UpToken,
			"return_url":   order.ReturnURL,
		},
	})
}

// CancelKycRecord 取消当前进行中的认证记录
func (h *AuthHandler) CancelKycRecord(c *gin.Context) {
	userID := c.GetInt64("user_id")

	err := h.authService.CancelKycRecord(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "认证已取消",
	})
}

// StartAuthForWeb Web 端发起 KYC 认证（账户实名）
func (h *AuthHandler) StartAuthForWeb(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// 从 JSON 读取参数
	var req struct {
		Name      string `json:"name" binding:"required"`
		IDCard    string `json:"id_card" binding:"required"`
		ReturnURL string `json:"return_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "name and id_card are required",
		})
		return
	}

	// 设置默认回调地址（Web 账户实名：认证完成后跳回账户实名页）
	if req.ReturnURL == "" {
		req.ReturnURL = "/user/kyc"
	}
	// notifyURL 留空，由 service 层使用 .env 环境变量中的 FINAUTH_NOTIFY_URL
	notifyURL := ""

	// 年龄限制：平台实名需年满16周岁
	if err := authValidator.ValidateIDCardMinAge(req.IDCard, utils.MinKycAge); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	// 生成业务流水号（用于标识这是用户实名认证）
	bizNo := fmt.Sprintf("WEB_%d_%d", userID, time.Now().Unix())

	// 标记这是用户实名认证
	bizExtraData := fmt.Sprintf(`{"type":"user_auth","user_id":%d}`, userID)

	// 发起认证
	result, err := h.authService.StartAuth(
		userID,
		req.Name,
		req.IDCard,
		bizNo,
		req.ReturnURL,
		notifyURL,
		bizExtraData,
		true, // Web 前端实名：经平台 /kyc 中转页处理回调，再跳回 /user/kyc
		true, // 账户实名：免费
	)
	if err != nil {
		if err == service.ErrInsufficientBalance {
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"message": "insufficient balance",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"platform_biz_no": result.Order.PlatformBizNo,
			"auth_url":        result.AuthURL,
			"expired_time":    result.Order.CreatedAt.Add(15 * time.Minute).Unix(),
			"expired_in":      900,
		},
	})
}

// GetUserAuthRecords 查询用户的认证记录列表（带分页）
func (h *AuthHandler) GetUserAuthRecords(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		Page     int `json:"page" form:"page"`
		PageSize int `json:"page_size" form:"page_size"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	orders, total, err := h.authService.GetUserAuthRecords(userID, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to query records",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":      orders,
			"total":     total,
			"page":      req.Page,
			"page_size": req.PageSize,
		},
	})
}

// GetUserAuthCallStats 统计用户近30天的认证调用次数
func (h *AuthHandler) GetUserAuthCallStats(c *gin.Context) {
	userID := c.GetInt64("user_id")

	dates, counts, err := h.authService.GetUserAuthCallStats(userID, 30)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to query call stats",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"dates":  dates,
			"counts": counts,
		},
	})
}

// CreateRecharge 发起充值（调用银联支付）
func (h *AuthHandler) CreateRecharge(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	// 创建充值订单
	order, err := h.balanceService.CreateRecharge(userID, req.Amount)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to create recharge order",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"pay_order_no": order.PayOrderNo,
			"amount":       order.Amount,
			"expire_time":  order.ExpireTime.Unix(),
			"qr_code":      fmt.Sprintf("https://yourdomain.com/pay/%s", order.PayOrderNo),
		},
	})
}

// GetRechargeResult 查询充值结果（轮询查询支付订单状态）
func (h *AuthHandler) GetRechargeResult(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		PayOrderNo string `json:"pay_order_no" form:"pay_order_no" binding:"required"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"message": "invalid request parameters",
			})
			return
		}
	}

	// 查询支付订单
	order, err := h.balanceService.GetPaymentOrder(req.PayOrderNo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "order not found",
		})
		return
	}

	// 检查订单是否属于该用户
	if order.UserID != userID {
		c.JSON(http.StatusOK, gin.H{
			"code":    403,
			"message": "permission denied: order does not belong to you",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"pay_order_no":     order.PayOrderNo,
			"amount":           order.Amount,
			"status":           order.Status,
			"channel":          order.Channel,
			"channel_trade_no": order.ChannelTradeNo,
			"paid_at":          order.PaidAt,
		},
	})
}
