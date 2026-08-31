package handler

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
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
	adminRepo           *repository.AdminRepository
	userRepo            *repository.UserRepository
	authRepo            *repository.AuthOrderRepository
	paymentRepo         *repository.PaymentOrderRepository
	configRepo          *repository.SystemConfigRepository
	balanceLogRepo      *repository.BalanceLogRepository
	resourcePackRepo    *repository.ResourcePackRepository
	internalAccountRepo *repository.InternalAccountRepository
	loginLogRepo        *repository.LoginLogRepository
	balanceSvc          *service.BalanceService
	authSvc             *service.AuthService
	jwtSecret           string
}

func NewAdminHandler(
	adminRepo *repository.AdminRepository,
	userRepo *repository.UserRepository,
	authRepo *repository.AuthOrderRepository,
	paymentRepo *repository.PaymentOrderRepository,
	configRepo *repository.SystemConfigRepository,
	balanceLogRepo *repository.BalanceLogRepository,
	resourcePackRepo *repository.ResourcePackRepository,
	internalAccountRepo *repository.InternalAccountRepository,
	loginLogRepo *repository.LoginLogRepository,
	balanceSvc *service.BalanceService,
	authSvc *service.AuthService,
	jwtSecret string,
) *AdminHandler {
	return &AdminHandler{
		adminRepo:           adminRepo,
		userRepo:            userRepo,
		authRepo:            authRepo,
		paymentRepo:         paymentRepo,
		configRepo:          configRepo,
		balanceLogRepo:      balanceLogRepo,
		resourcePackRepo:    resourcePackRepo,
		internalAccountRepo: internalAccountRepo,
		loginLogRepo:        loginLogRepo,
		balanceSvc:          balanceSvc,
		authSvc:             authSvc,
		jwtSecret:           jwtSecret,
	}
}

// logAdminOperation 记录管理员操作日志（仅写入 admin.log 文件，不额外建表）
func (h *AdminHandler) logAdminOperation(c *gin.Context, operation, resourceType string, resourceID int64, details string) {
	adminID, _ := c.Get("user_id")
	adminIDInt, _ := adminID.(int64)

	// 写入文件日志
	utils.AdminLogger.Printf("admin_id=%d operation=%s resource_type=%s resource_id=%d ip=%s details=%s",
		adminIDInt, operation, resourceType, resourceID, c.ClientIP(), details)
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
		h.loginLogRepo.InsertAdminLoginLog(&model.AdminLoginLog{
			Username:   req.Username,
			IP:         c.ClientIP(),
			UserAgent:  c.GetHeader("User-Agent"),
			Status:     0,
			FailReason: "account not found",
		})
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
		h.loginLogRepo.InsertAdminLoginLog(&model.AdminLoginLog{
			AdminID:    admin.ID,
			Username:   req.Username,
			IP:         c.ClientIP(),
			UserAgent:  c.GetHeader("User-Agent"),
			Status:     0,
			FailReason: "password mismatch",
		})
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "invalid username or password",
		})
		return
	}

	// 检查账号状态
	if admin.Status != 1 {
		h.loginLogRepo.InsertAdminLoginLog(&model.AdminLoginLog{
			AdminID:    admin.ID,
			Username:   req.Username,
			IP:         c.ClientIP(),
			UserAgent:  c.GetHeader("User-Agent"),
			Status:     0,
			FailReason: "account disabled",
		})
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

	// 记录管理员登录日志（成功）
	h.loginLogRepo.InsertAdminLoginLog(&model.AdminLoginLog{
		AdminID:   admin.ID,
		Username:  req.Username,
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Status:    1,
	})

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

	// 记录管理员登录操作日志（登录发生在鉴权之前，需手动设置 admin_id）
	c.Set("user_id", admin.ID)
	h.logAdminOperation(c, "login", "admin", admin.ID, admin.Username)

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

	// 记录配置更新操作日志（脱敏：不记录敏感配置值）
	h.logAdminOperation(c, "config_update", "config", 0, "keys="+strings.Join(sortedKeys(stringReq), ","))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// sortedKeys 返回 map 的键，按字典序排序
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
		Phone    string `json:"phone" binding:"required"`
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
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

	// 校验用户名格式（仅支持英文+数字+下划线）
	req.Username = strings.TrimSpace(req.Username)
	if !service.ValidateUsername(req.Username) {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "用户名仅支持英文、数字、下划线，长度3-32位",
		})
		return
	}

	// 校验邮箱格式
	req.Email = strings.TrimSpace(req.Email)
	if !service.ValidateEmail(req.Email) {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "邮箱格式不正确",
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

	// 检查用户名是否已存在
	exists, err = h.userRepo.CheckUsernameExists(req.Username)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to check username existence",
		})
		return
	}
	if exists {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "username already exists",
		})
		return
	}

	// 检查邮箱是否已存在
	exists, err = h.userRepo.CheckEmailExists(req.Email)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to check email existence",
		})
		return
	}
	if exists {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "email already exists",
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

	// 使用系统默认KYC单价（已取消个人单价设置，统一按平台价格扣费）
	kycPrice := 1.00
	priceStr, err := h.configRepo.GetConfig("kyc_price")
	if err == nil && priceStr != "" {
		if price, err := strconv.ParseFloat(priceStr, 64); err == nil && price > 0 {
			kycPrice = price
		}
	}

	// API Key 创建时自动生成（唯一），API Secret 需完成账户实名后再生成下发
	user := &model.PlatformUser{
		Phone:         req.Phone,
		Username:      req.Username,
		Email:         req.Email,
		PasswordHash:  string(hashedPassword),
		Balance:       0,
		APIKey:        utils.GenerateRandomKey(32),
		APISecret:     "",
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
	log.Printf("Admin %d manually registered user: phone=%s", adminID, req.Phone)

	// 记录操作日志
	h.logAdminOperation(c, "register", "user", user.ID, fmt.Sprintf("phone=%s username=%s email=%s", req.Phone, req.Username, req.Email))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "user registered successfully",
		"data": gin.H{
			"id":       user.ID,
			"phone":    user.Phone,
			"username": user.Username,
			"email":    user.Email,
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
			"username":        user.Username,
			"email":           user.Email,
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

	// 记录操作日志
	statusText := "禁用"
	if req.Status == 1 {
		statusText = "启用"
	}
	h.logAdminOperation(c, "user_status", "user", userID, "status="+statusText)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
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

	// 记录操作日志
	h.logAdminOperation(c, "user_delete", "user", userID, "")

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
		Amount       float64 `json:"amount" binding:"required"`
		BankSerialNo string  `json:"bank_serial_no" binding:"required"` // 银行流水单号（必填，用于对账）
		Remark       string  `json:"remark"`
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

	// 校验银行流水单号
	req.BankSerialNo = adminValidator.SanitizeString(strings.TrimSpace(req.BankSerialNo))
	if len(req.BankSerialNo) < 4 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "请填写有效的银行流水单号（至少4位）",
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
	err = h.balanceSvc.ManualRechargeBalance(userID, req.Amount, remark, req.BankSerialNo)
	if err != nil {
		log.Printf("Failed to recharge user balance: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to recharge balance",
		})
		return
	}

	// 记录操作日志
	h.logAdminOperation(c, "recharge", "user", userID, fmt.Sprintf("amount=%.2f bank_serial_no=%s", req.Amount, req.BankSerialNo))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "充值成功",
		"data": gin.H{
			"user_id":        userID,
			"amount":         req.Amount,
			"bank_serial_no": req.BankSerialNo,
			"remark":         remark,
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
			"username":        user.Username,
			"email":           user.Email,
			"balance":         user.Balance,
			"is_kyc_verified": user.IsKYCVerified,
			"kyc_name":        kycName,
			"kyc_id_card":     kycIDCard,
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

	// 记录操作日志
	h.logAdminOperation(c, "change_password", "admin", adminID, "")

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// ---------- 资源包管理（管理后台） ----------

// GetResourcePackList 获取资源包列表（管理后台，含已下架）
func (h *AdminHandler) GetResourcePackList(c *gin.Context) {
	packs, err := h.resourcePackRepo.ListPacks(nil)
	if err != nil {
		log.Printf("Failed to get resource pack list: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to get resource pack list",
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

// CreateResourcePack 创建资源包
func (h *AdminHandler) CreateResourcePack(c *gin.Context) {
	var req struct {
		Name        string  `json:"name" binding:"required"`
		TotalCount  int     `json:"total_count" binding:"required"`
		Price       float64 `json:"price" binding:"required"`
		Stock       int     `json:"stock"`  // 库存：-1-不限量，>=0-限量
		Status      int     `json:"status"` // 1-上架 0-下架
		Description string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	req.Name = adminValidator.SanitizeString(strings.TrimSpace(req.Name))
	if len(req.Name) < 1 || len(req.Name) > 100 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "资源包名称长度必须在1-100个字符之间",
		})
		return
	}
	if req.TotalCount <= 0 || req.TotalCount > 100000 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "认证次数必须大于0且不超过100000",
		})
		return
	}
	if req.Price <= 0 || req.Price > 100000 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "售价必须大于0且不超过100000",
		})
		return
	}
	if req.Stock < -1 {
		req.Stock = -1
	}
	if req.Status != 0 && req.Status != 1 {
		req.Status = 1
	}

	pack := &model.ResourcePack{
		Name:        req.Name,
		TotalCount:  req.TotalCount,
		Price:       req.Price,
		Stock:       req.Stock,
		Status:      req.Status,
		Description: adminValidator.SanitizeString(req.Description),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := h.resourcePackRepo.CreatePack(pack); err != nil {
		log.Printf("Failed to create resource pack: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to create resource pack",
		})
		return
	}

	// 记录操作日志
	h.logAdminOperation(c, "pack_create", "pack", pack.ID, fmt.Sprintf("name=%s total_count=%d price=%.2f stock=%d", pack.Name, pack.TotalCount, pack.Price, pack.Stock))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "资源包创建成功",
		"data":    pack,
	})
}

// UpdateResourcePack 更新资源包
func (h *AdminHandler) UpdateResourcePack(c *gin.Context) {
	packID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid pack id",
		})
		return
	}

	pack, err := h.resourcePackRepo.GetPackByID(packID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "resource pack not found",
		})
		return
	}

	var req struct {
		Name        string  `json:"name"`
		TotalCount  int     `json:"total_count"`
		Price       float64 `json:"price"`
		Stock       int     `json:"stock"`
		Status      int     `json:"status"`
		Description string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	// 仅更新传入的字段
	if req.Name != "" {
		req.Name = adminValidator.SanitizeString(strings.TrimSpace(req.Name))
		if len(req.Name) > 100 {
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"message": "资源包名称长度不能超过100个字符",
			})
			return
		}
		pack.Name = req.Name
	}
	if req.TotalCount > 0 {
		if req.TotalCount > 100000 {
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"message": "认证次数不能超过100000",
			})
			return
		}
		pack.TotalCount = req.TotalCount
	}
	if req.Price > 0 {
		if req.Price > 100000 {
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"message": "售价不能超过100000",
			})
			return
		}
		pack.Price = req.Price
	}
	pack.Stock = req.Stock
	if pack.Stock < -1 {
		pack.Stock = -1
	}
	if req.Status == 0 || req.Status == 1 {
		pack.Status = req.Status
	}
	pack.Description = adminValidator.SanitizeString(req.Description)
	pack.UpdatedAt = time.Now()

	if err := h.resourcePackRepo.UpdatePack(pack); err != nil {
		log.Printf("Failed to update resource pack: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to update resource pack",
		})
		return
	}

	// 记录操作日志
	h.logAdminOperation(c, "pack_update", "pack", packID, fmt.Sprintf("name=%s total_count=%d price=%.2f stock=%d status=%d", pack.Name, pack.TotalCount, pack.Price, pack.Stock, pack.Status))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "资源包更新成功",
		"data":    pack,
	})
}

// DeleteResourcePack 下架资源包（软删除：status=0，保留已售用户资源包记录）
func (h *AdminHandler) DeleteResourcePack(c *gin.Context) {
	packID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid pack id",
		})
		return
	}

	pack, err := h.resourcePackRepo.GetPackByID(packID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "resource pack not found",
		})
		return
	}

	pack.Status = 0
	pack.UpdatedAt = time.Now()
	if err := h.resourcePackRepo.UpdatePack(pack); err != nil {
		log.Printf("Failed to delete resource pack: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to delete resource pack",
		})
		return
	}

	// 记录操作日志
	h.logAdminOperation(c, "pack_delete", "pack", packID, "name="+pack.Name)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "资源包已下架",
	})
}

// ---------- 内部账号管理（本司其他系统专用，无需实名与计费，不可在用户端登录） ----------

// GetInternalAccountList 获取内部账号列表（分页 + 关键词搜索）
func (h *AdminHandler) GetInternalAccountList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	accounts, total, err := h.internalAccountRepo.List(strings.TrimSpace(keyword), page, pageSize)
	if err != nil {
		log.Printf("Failed to get internal account list: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to get internal account list",
		})
		return
	}

	accountList := make([]gin.H, 0, len(accounts))
	for _, acc := range accounts {
		accountList = append(accountList, gin.H{
			"id":         acc.ID,
			"name":       acc.Name,
			"remark":     acc.Remark,
			"api_key":    acc.APIKey,
			"api_secret": acc.APISecret,
			"status":     acc.Status,
			"created_at": acc.CreatedAt,
			"updated_at": acc.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":      accountList,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// CreateInternalAccount 创建内部账号（创建时即生成 API Key/Secret）
func (h *AdminHandler) CreateInternalAccount(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required"`
		Remark string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	req.Name = adminValidator.SanitizeString(strings.TrimSpace(req.Name))
	if len(req.Name) < 1 || len(req.Name) > 64 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "账号名称长度必须在1-64个字符之间",
		})
		return
	}

	// 名称唯一性校验
	if _, err := h.internalAccountRepo.GetByName(req.Name); err == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "内部账号名称已存在",
		})
		return
	}

	req.Remark = adminValidator.SanitizeString(req.Remark)
	if len(req.Remark) > 255 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "备注长度不能超过255个字符",
		})
		return
	}

	acc := &model.InternalAccount{
		Name:      req.Name,
		Remark:    req.Remark,
		APIKey:    utils.GenerateRandomKey(32),
		APISecret: utils.GenerateRandomKey(32),
		Status:    1, // 默认启用
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.internalAccountRepo.Create(acc); err != nil {
		log.Printf("Failed to create internal account: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to create internal account",
		})
		return
	}

	// 记录操作日志
	h.logAdminOperation(c, "internal_account_create", "internal_account", acc.ID, "name="+acc.Name)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "内部账号创建成功",
		"data": gin.H{
			"id":         acc.ID,
			"name":       acc.Name,
			"remark":     acc.Remark,
			"api_key":    acc.APIKey,
			"api_secret": acc.APISecret,
			"status":     acc.Status,
		},
	})
}

// UpdateInternalAccountStatus 启用/禁用内部账号
func (h *AdminHandler) UpdateInternalAccountStatus(c *gin.Context) {
	accID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accID <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid account id",
		})
		return
	}

	var req struct {
		Status int `json:"status" binding:"required"` // 0-禁用 1-启用
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

	if err := h.internalAccountRepo.UpdateStatus(accID, req.Status); err != nil {
		log.Printf("Failed to update internal account status: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "内部账号不存在",
		})
		return
	}

	// 记录操作日志
	statusText := "禁用"
	if req.Status == 1 {
		statusText = "启用"
	}
	h.logAdminOperation(c, "internal_account_status", "internal_account", accID, "status="+statusText)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// ResetInternalAccountAPI 重置内部账号 API Key/Secret（旧密钥立即失效）
func (h *AdminHandler) ResetInternalAccountAPI(c *gin.Context) {
	accID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accID <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid account id",
		})
		return
	}

	acc, err := h.internalAccountRepo.GetByID(accID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "内部账号不存在",
		})
		return
	}

	newKey := utils.GenerateRandomKey(32)
	newSecret := utils.GenerateRandomKey(32)
	if err := h.internalAccountRepo.ResetAPIKey(accID, newKey, newSecret); err != nil {
		log.Printf("Failed to reset internal account api key: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to reset internal account api key",
		})
		return
	}

	// 记录操作日志（不记录新密钥内容）
	h.logAdminOperation(c, "internal_account_reset_api", "internal_account", accID, "name="+acc.Name)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "重置成功",
		"data": gin.H{
			"id":         acc.ID,
			"name":       acc.Name,
			"api_key":    newKey,
			"api_secret": newSecret,
		},
	})
}
