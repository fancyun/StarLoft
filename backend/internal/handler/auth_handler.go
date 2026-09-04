package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"starloftrpa/internal/repository"
	"starloftrpa/internal/service"
	"starloftrpa/internal/upstream"
	"starloftrpa/internal/utils"
)

var authValidator = utils.NewInputValidator()

type AuthHandler struct {
	authService      *service.AuthService
	balanceService   *service.BalanceService
	resourcePackRepo *repository.ResourcePackRepository
	alipayClient     *upstream.AlipayClient
	wechatClient     *upstream.WeChatPayClient
}

func NewAuthHandler(
	authService *service.AuthService,
	balanceService *service.BalanceService,
	resourcePackRepo *repository.ResourcePackRepository,
	alipayClient *upstream.AlipayClient,
	wechatClient *upstream.WeChatPayClient,
) *AuthHandler {
	return &AuthHandler{
		authService:      authService,
		balanceService:   balanceService,
		resourcePackRepo: resourcePackRepo,
		alipayClient:     alipayClient,
		wechatClient:     wechatClient,
	}
}

// StartAuth 发起认证（API 调用）
func (h *AuthHandler) StartAuth(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// 平台用户需先完成实名才能调用
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
		2, // API 业务调用：始终计费
		false,
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

	// 平台用户需先完成实名才能调用
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

// QueryBalance 查询余额
func (h *AuthHandler) QueryBalance(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// 获取平台认证单价（统一按平台价格扣费，已取消个人单价设置）
	kycPrice := h.authService.GetPlatformKycPrice()

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

	// 如果状态为进行中，返回关联的认证 token 与认证地址（用于继续认证，直接跳转上游）
	if kycRecord != nil && kycRecord.Status == 1 {
		pendingToken := kycRecord.UpToken
		pendingBizNo := kycRecord.PlatformBizNo
		// 兼容历史：API 调用（source=2）产生的实名记录不存上游 token，回退查订单
		if pendingToken == "" {
			if pendingOrder, err := h.authService.GetLatestPendingOrder(userID); err == nil && pendingOrder != nil {
				pendingToken = pendingOrder.UpToken
				pendingBizNo = pendingOrder.PlatformBizNo
			}
		}
		if pendingToken != "" {
			resp["pending_token"] = pendingToken
			resp["pending_auth_url"] = h.authService.BuildAuthURL(pendingToken)
		}
		if pendingBizNo != "" {
			resp["pending_biz_no"] = pendingBizNo
		}
		// 返回下游的 return_url（认证完成后跳转回下游地址）
		if kycRecord.ReturnURL != "" {
			resp["return_url"] = kycRecord.ReturnURL
		}
	}

	// 如果已有结果（状态 2/3），也返回 return_url 用于结果页跳转
	if kycRecord != nil && (kycRecord.Status == 2 || kycRecord.Status == 3) {
		if resp["return_url"] == nil && kycRecord.ReturnURL != "" {
			resp["return_url"] = kycRecord.ReturnURL
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

	record, err := h.authService.SyncKycRecord(userID)
	if err != nil {
		// 没有进行中的认证记录，返回当前状态
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
			"synced":        true,
			"record_status": record.Status,
			"up_token":      record.UpToken,
			"return_url":    record.ReturnURL,
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
	// notifyURL 留空，由 service 层使用写死的上游异步通知地址（finAuthNotifyURL）
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

	// 发起认证（账户实名 source=1：不写认证订单、不计费，实名信息单独储存在系统库实名记录表）
	result, err := h.authService.StartAuth(
		userID,
		req.Name,
		req.IDCard,
		bizNo,
		req.ReturnURL,
		notifyURL,
		bizExtraData,
		1,    // 账户实名
		true, // retained参数，账户实名始终免费，不再使用
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
			"platform_biz_no": result.Record.PlatformBizNo,
			"auth_url":        result.AuthURL,
			"expired_time":    result.Record.CreatedAt.Add(15 * time.Minute).Unix(),
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

	records, total, err := h.authService.GetUserAuthRecords(userID, req.Page, req.PageSize)
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
			"list":      records,
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

// CreateRecharge 发起充值（支付宝电脑网站支付 / 微信Native支付）
func (h *AuthHandler) CreateRecharge(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		Amount  float64 `json:"amount" binding:"required,gt=0"`
		Channel string  `json:"channel" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	if req.Channel != "alipay" && req.Channel != "wechat" {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "不支持的支付渠道",
		})
		return
	}

	// 创建充值订单
	order, err := h.balanceService.CreateRecharge(userID, req.Amount, req.Channel)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to create recharge order",
		})
		return
	}

	data := gin.H{
		"pay_order_no": order.PayOrderNo,
		"amount":       order.Amount,
		"expire_time":  order.ExpireTime.Unix(),
		"channel":      order.Channel,
	}

	// 按渠道生成支付信息
	switch req.Channel {
	case "alipay":
		if h.alipayClient == nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"message": "支付宝支付未配置",
			})
			return
		}
		payURL, err := h.alipayClient.BuildPagePayURL(order.PayOrderNo, order.Amount)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"message": "生成支付宝支付链接失败",
			})
			return
		}
		data["pay_url"] = payURL
	case "wechat":
		if h.wechatClient == nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"message": "微信支付未配置",
			})
			return
		}
		codeURL, err := h.wechatClient.CreateNativeOrder(order.PayOrderNo, order.Amount, "账户余额充值")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"message": "生成微信支付二维码失败",
			})
			return
		}
		data["code_url"] = codeURL
		data["qr_url"] = "/console/qr?data=" + url.QueryEscape(codeURL) + "&size=280"
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
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

// ---------- 资源包（Web 用户） ----------

// ListResourcePacks 在售资源包列表
func (h *AuthHandler) ListResourcePacks(c *gin.Context) {
	status := 1
	packs, err := h.resourcePackRepo.ListPacks(&status)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to get resource packs",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list": packs,
		},
	})
}

// PurchaseResourcePack 使用余额购买资源包（不支持直接为资源包付费，需先充值再购买）
func (h *AuthHandler) PurchaseResourcePack(c *gin.Context) {
	userID := c.GetInt64("user_id")
	packID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || packID <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid pack id",
		})
		return
	}

	up, err := h.balanceService.PurchaseResourcePack(userID, packID)
	if err != nil {
		switch err {
		case service.ErrInsufficientBalance:
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"message": "余额不足，请先充值",
			})
		case repository.ErrPackSoldOut:
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"message": "资源包已售罄",
			})
		case repository.ErrPackOffSale:
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"message": "资源包已下架",
			})
		case repository.ErrPackNotFound:
			c.JSON(http.StatusOK, gin.H{
				"code":    404,
				"message": "资源包不存在",
			})
		default:
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"message": "购买资源包失败",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "购买成功",
		"data": gin.H{
			"user_pack": up,
		},
	})
}

// MyResourcePacks 我的资源包列表
func (h *AuthHandler) MyResourcePacks(c *gin.Context) {
	userID := c.GetInt64("user_id")
	packs, err := h.resourcePackRepo.ListUserPacks(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to get my resource packs",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list": packs,
		},
	})
}
