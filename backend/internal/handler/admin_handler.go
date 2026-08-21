package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"starloftrpa/internal/model"
	"starloftrpa/internal/repository"
	"starloftrpa/internal/service"
	"starloftrpa/internal/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var adminValidator = utils.NewInputValidator()

type AdminHandler struct {
	adminRepo      *repository.AdminRepository
	userRepo       *repository.UserRepository
	authRepo       *repository.AuthOrderRepository
	paymentRepo    *repository.PaymentOrderRepository
	configRepo     *repository.SystemConfigRepository
	balanceLogRepo *repository.BalanceLogRepository
	balanceSvc     *service.BalanceService
	authSvc        *service.AuthService
	jwtSecret      string
}

func NewAdminHandler(
	adminRepo *repository.AdminRepository,
	userRepo *repository.UserRepository,
	authRepo *repository.AuthOrderRepository,
	paymentRepo *repository.PaymentOrderRepository,
	configRepo *repository.SystemConfigRepository,
	balanceLogRepo *repository.BalanceLogRepository,
	balanceSvc *service.BalanceService,
	authSvc *service.AuthService,
	jwtSecret string,
) *AdminHandler {
	return &AdminHandler{
		adminRepo:      adminRepo,
		userRepo:       userRepo,
		authRepo:       authRepo,
		paymentRepo:    paymentRepo,
		configRepo:     configRepo,
		balanceLogRepo: balanceLogRepo,
		balanceSvc:     balanceSvc,
		authSvc:        authSvc,
		jwtSecret:      jwtSecret,
	}
}

// AdminLogin 管理员登录
func (h *AdminHandler) AdminLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	// 查询管理员
	admin, err := h.adminRepo.GetAdminByUsername(req.Username)
	if err != nil {
		log.Printf("Admin login failed: username=%s, error=%v", req.Username, err)
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "invalid username or password",
		})
		return
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password))
	if err != nil {
		log.Printf("Admin login failed: username=%s, password mismatch", req.Username)
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "invalid username or password",
		})
		return
	}

	// 检查账号状态
	if admin.Status != 1 {
		c.JSON(http.StatusOK, gin.H{
			"code":    403,
			"message": "account disabled",
		})
		return
	}

	// 更新最后登录时间
	err = h.adminRepo.UpdateLastLoginTime(admin.ID)
	if err != nil {
		log.Printf("Failed to update admin last login time: %v", err)
	}

	// 生成 JWT Token
	jwtManager := utils.NewJWTManager(h.jwtSecret)
	token, err := jwtManager.GenerateToken(admin.ID, admin.Username, "admin", 24*time.Hour)
	if err != nil {
		log.Printf("Failed to generate admin token: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"admin_id":  admin.ID,
			"username":  admin.Username,
			"nickname":  admin.Nickname,
			"token":     token,
			"expire_in": 86400,
		},
	})
}

// GetSystemConfig 获取系统配置
func (h *AdminHandler) GetSystemConfig(c *gin.Context) {
	configs, err := h.configRepo.GetAllConfigs()
	if err != nil {
		log.Printf("Failed to get system config: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to get system config",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    configs,
	})
}

// UpdateSystemConfig 更新系统配置
func (h *AdminHandler) UpdateSystemConfig(c *gin.Context) {
	var req map[string]interface{}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	// 将所有值转换为字符串
	stringReq := make(map[string]string)
	for k, v := range req {
		stringReq[k] = fmt.Sprintf("%v", v)
	}

	err := h.configRepo.BatchUpdateConfigs(stringReq)
	if err != nil {
		log.Printf("Failed to update system config: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to update system config",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// GetStatisticsOverview 获取统计概览
func (h *AdminHandler) GetStatisticsOverview(c *gin.Context) {
	// 统计用户数
	_, totalUsers, err := h.userRepo.GetAllUsers(1, 1, "")
	if err != nil {
		totalUsers = 0
	}

	today := time.Now().Format("2006-01-02")
	monthStart := time.Now().Format("2006-01") + "-01"

	// 今日订单数
	todayOrderStats, _ := h.authRepo.GetDailyOrderStats(today, today)
	var todayOrders int64 = 0
	if v, ok := todayOrderStats[today]; ok {
		todayOrders = v
	}

	// 今日收入（仅统计已完成订单）
	todayIncomeStats, _ := h.authRepo.GetDailyIncomeStats(today, today)
	var todayRevenue float64 = 0
	if v, ok := todayIncomeStats[today]; ok {
		todayRevenue = v
	}

	// 本月收入（仅统计已完成订单）
	monthIncomeStats, _ := h.authRepo.GetDailyIncomeStats(monthStart, today)
	var monthRevenue float64 = 0
	for _, v := range monthIncomeStats {
		monthRevenue += v
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"total_users":   totalUsers,
			"today_orders":  todayOrders,
			"today_revenue": todayRevenue,
			"month_revenue": monthRevenue,
		},
	})
}

// GetOrderStatistics 获取订单统计（按日期）
func (h *AdminHandler) GetOrderStatistics(c *gin.Context) {
	days := c.DefaultQuery("days", "7")
	dayNum, err := strconv.Atoi(days)
	if err != nil || dayNum <= 0 || dayNum > 90 {
		dayNum = 7
	}

	startDate := time.Now().AddDate(0, 0, -(dayNum - 1)).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	stats, err := h.authRepo.GetDailyOrderStats(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to get order statistics",
		})
		return
	}

	dates := make([]string, 0, dayNum)
	counts := make([]int64, 0, dayNum)
	for i := dayNum - 1; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		dates = append(dates, d)
		counts = append(counts, stats[d])
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

// GetIncomeStatistics 获取收入统计
func (h *AdminHandler) GetIncomeStatistics(c *gin.Context) {
	days := c.DefaultQuery("days", "7")
	dayNum, err := strconv.Atoi(days)
	if err != nil || dayNum <= 0 || dayNum > 90 {
		dayNum = 7
	}

	startDate := time.Now().AddDate(0, 0, -(dayNum - 1)).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	stats, err := h.authRepo.GetDailyIncomeStats(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to get income statistics",
		})
		return
	}

	dates := make([]string, 0, dayNum)
	amounts := make([]float64, 0, dayNum)
	for i := dayNum - 1; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		dates = append(dates, d)
		amounts = append(amounts, stats[d])
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"dates":   dates,
			"amounts": amounts,
		},
	})
}

// GetRecentAuthOrders 获取最近认证订单列表
func (h *AdminHandler) GetRecentAuthOrders(c *gin.Context) {
	limit := 10
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}

	orders, err := h.authRepo.GetRecentOrders(limit)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to get recent orders",
		})
		return
	}

	records := make([]gin.H, 0, len(orders))
	for _, o := range orders {
		records = append(records, gin.H{
			"platform_biz_no": o.PlatformBizNo,
			"user_phone":      o.UserPhone,
			"name":            o.Name,
			"status":          o.Status,
			"cost":            o.Cost,
			"created_at":      o.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list": records,
		},
	})
}

// ManualRegisterUser 管理员手动注册用户
func (h *AdminHandler) ManualRegisterUser(c *gin.Context) {
	var req struct {
		Phone    string  `json:"phone" binding:"required"`
		Password string  `json:"password" binding:"required"`
		KYCPrice float64 `json:"kyc_price"` // 可选，不传则使用系统默认价格
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	// 验证手机号格式
	if err := adminValidator.ValidatePhone(req.Phone); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid phone format",
		})
		return
	}

	// 验证密码长度
	if len(req.Password) < 6 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "password must be at least 6 characters",
		})
		return
	}

	// 检查手机号是否已存在
	exists, err := h.userRepo.CheckPhoneExists(req.Phone)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to check phone existence",
		})
		return
	}
	if exists {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "phone already exists",
		})
		return
	}

	// 密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to hash password",
		})
		return
	}

	// 生成 API Key 和 API Secret
	apiKey := generateRandomKey(32)
	apiSecret := generateRandomKey(32)

	// 确定KYC单价
	kycPrice := req.KYCPrice
	if kycPrice <= 0 {
		// 使用系统默认价格
		priceStr, err := h.configRepo.GetConfig("kyc_price")
		if err == nil && priceStr != "" {
			if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
				kycPrice = price
			}
		}
		if kycPrice <= 0 {
			kycPrice = 1.00 // 兜底默认价格
		}
	}

	// 创建用户
	user := &model.PlatformUser{
		Phone:         req.Phone,
		PasswordHash:  string(hashedPassword),
		Balance:       0,
		APIKey:        apiKey,
		APISecret:     apiSecret,
		IsKYCVerified: 0,
		KYCPrice:      kycPrice,
		Status:        1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = h.userRepo.CreateUser(user)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to create user",
		})
		return
	}

	adminID := c.GetInt64("user_id")
	log.Printf("Admin %d manually registered user: phone=%s, kyc_price=%.2f", adminID, req.Phone, kycPrice)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "user registered successfully",
		"data": gin.H{
			"id":        user.ID,
			"phone":     user.Phone,
			"api_key":   user.APIKey,
			"kyc_price": user.KYCPrice,
		},
	})
}

// GetAuthOrderList 获取认证订单列表
func (h *AdminHandler) GetAuthOrderList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	statusStr := c.Query("status")
	userIDStr := c.Query("user_id")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var status *int
	var userID *int64

	if statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err == nil {
			status = &s
		}
	}

	if userIDStr != "" {
		uid, err := strconv.ParseInt(userIDStr, 10, 64)
		if err == nil {
			userID = &uid
		}
	}

	orders, total, err := h.authRepo.GetAllOrders(page, pageSize, status, userID)
	if err != nil {
		log.Printf("Failed to get auth order list: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to get order list",
		})
		return
	}

	orderList := make([]gin.H, 0, len(orders))
	for _, order := range orders {
		orderList = append(orderList, gin.H{
			"id":              order.ID,
			"platform_biz_no": order.PlatformBizNo,
			"biz_no":          order.BizNo,
			"user_id":         order.UserID,
			"user_phone":      order.UserPhone,
			"status":          order.Status,
			"cost":            order.Cost,
			"result_code":     order.ResultCode,
			"result_message":  order.ResultMessage,
			"is_refunded":     order.IsRefunded,
			"created_at":      order.CreatedAt,
			"finished_at":     order.FinishedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":      orderList,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetAuthOrderDetail 获取认证订单详情
func (h *AdminHandler) GetAuthOrderDetail(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid order id",
		})
		return
	}

	order, err := h.authRepo.GetOrderByID(orderID)
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
			"id":              order.ID,
			"platform_biz_no": order.PlatformBizNo,
			"biz_no":          order.BizNo,
			"user_id":         order.UserID,
			"return_url":      order.ReturnURL,
			"notify_url":      order.NotifyURL,
			"biz_extra_data":  order.BizExtraData,
			"up_token":        order.UpToken,
			"up_biz_id":       order.UpBizID,
			"up_request_id":   order.UpRequestID,
			"result_code":     order.ResultCode,
			"result_message":  order.ResultMessage,
			"result_data":     order.ResultData,
			"status":          order.Status,
			"cost":            order.Cost,
			"is_refunded":     order.IsRefunded,
			"notify_times":    order.NotifyTimes,
			"notify_status":   order.NotifyStatus,
			"created_at":      order.CreatedAt,
			"updated_at":      order.UpdatedAt,
			"finished_at":     order.FinishedAt,
		},
	})
}

// GetPaymentOrderList 获取支付订单列表
func (h *AdminHandler) GetPaymentOrderList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	statusStr := c.Query("status")
	userIDStr := c.Query("user_id")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var status *int
	var userID *int64

	if statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err == nil {
			status = &s
		}
	}

	if userIDStr != "" {
		uid, err := strconv.ParseInt(userIDStr, 10, 64)
		if err == nil {
			userID = &uid
		}
	}

	orders, total, err := h.paymentRepo.GetAllOrders(page, pageSize, status, userID)
	if err != nil {
		log.Printf("Failed to get payment order list: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to get order list",
		})
		return
	}

	orderList := make([]gin.H, 0, len(orders))
	for _, order := range orders {
		orderList = append(orderList, gin.H{
			"id":               order.ID,
			"pay_order_no":     order.PayOrderNo,
			"user_id":          order.UserID,
			"amount":           order.Amount,
			"channel":          order.Channel,
			"channel_trade_no": order.ChannelTradeNo,
			"status":           order.Status,
			"refund_status":    order.RefundStatus,
			"refund_amount":    order.RefundAmount,
			"paid_at":          order.PaidAt,
			"refunded_at":      order.RefundedAt,
			"created_at":       order.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":      orderList,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// generateRandomKey 生成随机密钥
func generateRandomKey(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GetUserDetail 获取用户详情
func (h *AdminHandler) GetUserDetail(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid user id",
		})
		return
	}

	user, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "user not found",
		})
		return
	}

	// 提取 NullString 的值
	kycName := ""
	if user.KYCName.Valid {
		kycName = user.KYCName.String
	}
	kycIDCard := ""
	if user.KYCIDCard.Valid {
		kycIDCard = user.KYCIDCard.String
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"id":              user.ID,
			"phone":           user.Phone,
			"balance":         user.Balance,
			"api_key":         user.APIKey,
			"is_kyc_verified": user.IsKYCVerified,
			"kyc_name":        kycName,
			"kyc_id_card":     kycIDCard,
			"status":          user.Status,
			"last_login_at":   user.LastLoginAt,
			"created_at":      user.CreatedAt,
			"updated_at":      user.UpdatedAt,
		},
	})
}

// UpdateUserStatus 更新用户状态
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid user id",
		})
		return
	}

	var req struct {
		Status int `json:"status" binding:"required"` // 0-禁用 1-正常
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	if req.Status != 0 && req.Status != 1 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid status value",
		})
		return
	}

	err = h.userRepo.UpdateUserStatus(userID, req.Status)
	if err != nil {
		log.Printf("Failed to update user status: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to update user status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// UpdateUserDiscount 更新用户KYC单价（原折扣功能）
func (h *AdminHandler) UpdateUserDiscount(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid user id",
		})
		return
	}

	var req struct {
		KYCPrice float64 `json:"kyc_price" binding:"required"` // KYC单价（元）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	// 验证单价范围（0.01 - 100元）
	if req.KYCPrice < 0.01 || req.KYCPrice > 100.0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "KYC单价必须在0.01-100.00元之间",
		})
		return
	}

	err = h.userRepo.UpdateUserKYCPrice(userID, req.KYCPrice)
	if err != nil {
		log.Printf("Failed to update user KYC price: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to update kyc price",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "KYC单价更新成功",
		"data": gin.H{
			"user_id":   userID,
			"kyc_price": req.KYCPrice,
		},
	})
}

// DeleteUser 删除用户
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid user id",
		})
		return
	}

	err = h.userRepo.DeleteUser(userID)
	if err != nil {
		log.Printf("Failed to delete user: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to delete user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// RechargeUserBalance 为用户手动充值
func (h *AdminHandler) RechargeUserBalance(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid user id",
		})
		return
	}

	var req struct {
		Amount float64 `json:"amount" binding:"required"`
		Remark string  `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "amount must be greater than 0",
		})
		return
	}

	// 验证金额范围
	if err := adminValidator.ValidateAmount(req.Amount); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "amount must be between 0 and 1,000,000",
		})
		return
	}

	// 清理备注中的潜在XSS
	req.Remark = adminValidator.SanitizeString(req.Remark)

	adminID := c.GetInt64("user_id")
	remark := req.Remark
	if remark == "" {
		remark = fmt.Sprintf("管理员人工充值(admin_id:%d)", adminID)
	} else {
		remark = fmt.Sprintf("管理员人工充值(admin_id:%d): %s", adminID, req.Remark)
	}

	// 人工充值（type=1，增加余额）
	err = h.balanceSvc.ManualRechargeBalance(userID, req.Amount, remark)
	if err != nil {
		log.Printf("Failed to recharge user balance: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to recharge balance",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "充值成功",
		"data": gin.H{
			"user_id": userID,
			"amount":  req.Amount,
			"remark":  remark,
		},
	})
}

// GiftUserBalance 赠送用户余额（促销、补偿等）
func (h *AdminHandler) GiftUserBalance(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid user id",
		})
		return
	}

	var req struct {
		Amount float64 `json:"amount" binding:"required"`
		Reason string  `json:"reason" binding:"required"` // 赠送原因（必填）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "赠送金额必须大于0",
		})
		return
	}

	// 验证金额范围
	if err := adminValidator.ValidateAmount(req.Amount); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "amount must be between 0 and 1,000,000",
		})
		return
	}

	// 清理原因中的潜在XSS
	req.Reason = adminValidator.SanitizeString(req.Reason)

	if len(req.Reason) < 2 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "赠送原因至少2个字符",
		})
		return
	}

	// 验证用户是否存在
	user, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		log.Printf("Failed to get user: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to get user",
		})
		return
	}
	if user == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "user not found",
		})
		return
	}

	adminID := c.GetInt64("user_id")
	remark := fmt.Sprintf("管理员余额赠送(admin_id:%d): %s", adminID, req.Reason)

	// 赠送余额
	err = h.balanceSvc.GiftBalance(userID, req.Amount, remark)
	if err != nil {
		log.Printf("Failed to gift user balance: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "赠送失败",
		})
		return
	}

	// 查询用户最新余额
	user, err = h.userRepo.GetUserByID(userID)
	if err != nil {
		log.Printf("Failed to get user info after gift: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "赠送成功",
			"data": gin.H{
				"user_id":     userID,
				"gift_amount": req.Amount,
				"reason":      req.Reason,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "赠送成功",
		"data": gin.H{
			"user_id":     userID,
			"gift_amount": req.Amount,
			"new_balance": user.Balance,
			"reason":      req.Reason,
		},
	})
}

// GetUserList 获取用户列表
func (h *AdminHandler) GetUserList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	phone := c.Query("phone")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	users, total, err := h.userRepo.GetAllUsers(page, pageSize, phone)
	if err != nil {
		log.Printf("Failed to get user list: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to get user list",
		})
		return
	}

	// 数据脱敏处理
	userList := make([]gin.H, 0, len(users))
	for _, user := range users {
		// 提取 NullString 的值，避免序列化为 {"String":"...","Valid":true}
		kycName := ""
		if user.KYCName.Valid {
			kycName = user.KYCName.String
		}
		kycIDCard := ""
		if user.KYCIDCard.Valid {
			kycIDCard = user.KYCIDCard.String
		}

		userList = append(userList, gin.H{
			"id":              user.ID,
			"phone":           user.Phone,
			"balance":         user.Balance,
			"is_kyc_verified": user.IsKYCVerified,
			"kyc_name":        kycName,
			"kyc_id_card":     kycIDCard,
			"kyc_price":       user.KYCPrice,
			"api_key":         user.APIKey,
			"status":          user.Status,
			"last_login_at":   user.LastLoginAt,
			"created_at":      user.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":      userList,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetUserFinanceStats 获取用户财务统计
func (h *AdminHandler) GetUserFinanceStats(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	stats, err := h.balanceLogRepo.GetUserFinanceStats(userID)
	if err != nil {
		log.Printf("Failed to get user finance stats: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to get finance stats",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

// GetUserBalanceLogs 获取用户余额流水
func (h *AdminHandler) GetUserBalanceLogs(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	logs, total, err := h.balanceLogRepo.GetUserBalanceLogs(userID, page, pageSize)
	if err != nil {
		log.Printf("Failed to get user balance logs: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to get balance logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":      logs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetUserAuthOrders 获取用户认证订单
func (h *AdminHandler) GetUserAuthOrders(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	orders, total, err := h.authRepo.GetUserAuthOrders(userID, page, pageSize)
	if err != nil {
		log.Printf("Failed to get user auth orders: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to get auth orders",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":      orders,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// ChangePassword 修改管理员密码
func (h *AdminHandler) ChangePassword(c *gin.Context) {
	adminIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "unauthorized",
		})
		return
	}
	adminID, ok := adminIDVal.(int64)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "invalid admin identity",
		})
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	// 查询当前管理员
	admin, err := h.adminRepo.GetAdminByID(adminID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "admin not found",
		})
		return
	}

	// 校验旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid old password",
		})
		return
	}

	// 校验新密码（与创建管理员一致，至少6位）
	if len(req.NewPassword) < 6 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "password must be at least 6 characters",
		})
		return
	}

	// 生成新密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash new password: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to hash password",
		})
		return
	}

	// 更新密码
	if err := h.adminRepo.UpdateAdminPassword(adminID, string(hashedPassword)); err != nil {
		log.Printf("Failed to update admin password: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to change password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}
