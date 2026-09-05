package router

import (
	"net/http"

	"starloftrpa/internal/config"
	"starloftrpa/internal/database"
	"starloftrpa/internal/handler"
	"starloftrpa/internal/middleware"
	"starloftrpa/internal/redis"
	"starloftrpa/internal/repository"
	"starloftrpa/internal/runtime"
	"starloftrpa/internal/service"
	"starloftrpa/internal/utils"

	"github.com/gin-gonic/gin"
)

func Setup(cfg *config.Config) (*gin.Engine, *service.AuthService, *service.BalanceService) {
	r := gin.New()

	// 全局中间件
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.Recovery())

	// 健康检查：/healthz 存活探针（进程存活即通过）；/readyz 就绪探针（依赖的 DB/Redis 可用才通过）
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		ctx := c.Request.Context()
		if err := database.DB.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "detail": "database"})
			return
		}
		if err := redis.Client.Ping(ctx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "detail": "redis"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// Prometheus 指标采集端点
	r.GET("/metrics", metricsHandler)

	// 获取数据库连接
	db := database.DB

	// 初始化 Repository
	userRepo := repository.NewUserRepository(db)
	apiRepo := repository.NewApiKeyRepository(db)
	authRepo := repository.NewAuthOrderRepository(db)
	paymentRepo := repository.NewPaymentOrderRepository(db)
	balanceLogRepo := repository.NewBalanceLogRepository(db)
	adminRepo := repository.NewAdminRepository(db)
	kycPersonalRepo := repository.NewKycPersonalRepository(db)
	kycEnterpriseRepo := repository.NewKycEnterpriseRepository(db)
	resourcePackRepo := repository.NewResourcePackRepository(db)
	loginLogRepo := repository.NewLoginLogRepository(db)

	// 初始化运行时：从环境变量配置构建第三方业务客户端快照
	rt, err := runtime.New(cfg)
	if err != nil {
		panic("初始化运行时配置失败: " + err.Error())
	}

	// 初始化 Service
	userService := service.NewUserService(userRepo, apiRepo)
	balanceService := service.NewBalanceService(userRepo, balanceLogRepo, paymentRepo, resourcePackRepo, db)

	authService := service.NewAuthService(
		rt.FinAuth,
		rt.FinAuthCfg,
		authRepo,
		userRepo,
		apiRepo,
		kycPersonalRepo,
		kycEnterpriseRepo,
		resourcePackRepo,
		balanceService,
		cfg.KycPrice,
	)

	// 初始化 JWT Manager
	jwtManager := utils.NewJWTManager(cfg.JWT.Secret)
	signMgr := utils.NewSignatureManager()

	// 初始化 Handler
	publicHandler := handler.NewPublicHandler(rt)

	userHandler := handler.NewUserHandler(
		userService,
		rt,
		jwtManager,
		authService,
		loginLogRepo,
	)

	authHandler := handler.NewAuthHandler(authService, balanceService, resourcePackRepo, rt)

	adminHandler := handler.NewAdminHandler(adminRepo,
		userRepo,
		authRepo,
		paymentRepo,
		balanceLogRepo,
		resourcePackRepo,
		loginLogRepo,
		balanceService,
		authService,
		rt,
		cfg.JWT.AdminSecret,
	)

	callbackHandler := handler.NewCallbackHandler(
		authService,
		balanceService,
		paymentRepo,
		rt,
	)

	dashboardHandler := handler.NewDashboardHandler(db)
	enterpriseAdminHandler := handler.NewEnterpriseAdminHandler(authService)

	// 路由注册（按前端分为三组）

	// ============ Console 前端路由 ============
	console := r.Group("/console")
	{
		// 公开接口（无需认证，Console 前端加载验证码配置、渲染支付二维码等）
		console.GET("/config", publicHandler.GetPublicConfig)
		console.GET("/qr", publicHandler.GetQRCode)

		// 用户相关路由（Web前端调用，JWT鉴权）
		console.POST("/send-code", userHandler.SendCode)
		console.POST("/send-email-code", userHandler.SendEmailCode)
		console.POST("/register", userHandler.Register)
		console.POST("/login", userHandler.Login)

		// 需要JWT认证的路由
		auth := console.Group("", middleware.JWTAuth(cfg.JWT.Secret))
		{
			auth.GET("/profile", userHandler.GetProfile)
			auth.GET("/kyc/status", authHandler.GetUserAuthStatus)
			auth.POST("/kyc/sync", authHandler.SyncKycResult)
			auth.POST("/kyc", authHandler.StartAuthForWeb)
			auth.DELETE("/kyc", authHandler.CancelKycRecord)
			// 企业实名（Web）
			auth.GET("/kyc/enterprise/status", authHandler.GetEnterpriseAuthStatus)
			auth.POST("/kyc/enterprise", authHandler.StartEnterpriseAuthForWeb)
			auth.GET("/records", authHandler.GetUserAuthRecords)
			auth.GET("/stats/calls", authHandler.GetUserAuthCallStats)
			auth.POST("/recharge", authHandler.CreateRecharge)
			auth.GET("/recharge/result", authHandler.GetRechargeResult)
			auth.POST("/api-key/reset", userHandler.ResetAPIKey)
			auth.POST("/api-key/permission", userHandler.SetAPIKeyPermission)
			auth.POST("/change-password", userHandler.ChangePassword)
			// 资源包（余额购买 / 在线组合支付）
			auth.GET("/packs", authHandler.ListResourcePacks)                   // 在售资源包列表
			auth.POST("/packs/:id/purchase", authHandler.PurchaseResourcePack)  // 使用余额购买资源包
			auth.POST("/packs/:id/pay", authHandler.PurchaseResourcePackOnline) // 在线购买（余额+支付宝/微信组合支付）
			auth.GET("/packs/mine", authHandler.MyResourcePacks)                // 我的资源包
		}
	}

	// ============ Admin 前端路由 ============
	admin := r.Group("/admin")
	{
		// 管理员登录 - 限流每分钟5次
		admin.POST("/login", middleware.RateLimiterForIP(5), adminHandler.AdminLogin)

		// 需要管理员JWT认证的路由 - 每分钟100次
		adminAuth := admin.Group("", middleware.JWTAuth(cfg.JWT.AdminSecret), middleware.RateLimiter(100))
		{
			// 用户管理
			adminAuth.GET("/users", adminHandler.GetUserList)
			adminAuth.GET("/users/:id", adminHandler.GetUserDetail)

			// 企业实名管理
			adminAuth.GET("/kyc/enterprise", enterpriseAdminHandler.ListEnterpriseRecords)
			adminAuth.POST("/kyc/enterprise/verify", enterpriseAdminHandler.VerifyEnterprise)
			adminAuth.PUT("/users/:id/status", adminHandler.UpdateUserStatus)
			adminAuth.DELETE("/users/:id", adminHandler.DeleteUser)
			adminAuth.POST("/users/:id/recharge", adminHandler.RechargeUserBalance)     // 人工充值（需银行流水单号）
			adminAuth.GET("/users/:id/finance/stats", adminHandler.GetUserFinanceStats) // 用户财务统计
			adminAuth.GET("/users/:id/balance-logs", adminHandler.GetUserBalanceLogs)   // 用户余额流水
			adminAuth.GET("/users/:id/auth-orders", adminHandler.GetUserAuthOrders)     // 用户认证订单
			adminAuth.POST("/users/register", adminHandler.ManualRegisterUser)          // 人工注册账号

			// 资源包管理
			adminAuth.GET("/packs", adminHandler.GetResourcePackList)
			adminAuth.POST("/packs", adminHandler.CreateResourcePack)
			adminAuth.PUT("/packs/:id", adminHandler.UpdateResourcePack)
			adminAuth.DELETE("/packs/:id", adminHandler.DeleteResourcePack)

			// 订单管理
			adminAuth.GET("/orders", adminHandler.GetAuthOrderList)
			adminAuth.GET("/orders/recent", adminHandler.GetRecentAuthOrders)
			adminAuth.GET("/orders/:id", adminHandler.GetAuthOrderDetail)
			adminAuth.GET("/payments", adminHandler.GetPaymentOrderList)

			// 管理员修改密码
			adminAuth.POST("/change-password", adminHandler.ChangePassword)

			// Dashboard数据
			adminAuth.GET("/dashboard", dashboardHandler.GetDashboard)

			// 财务统计
			adminAuth.GET("/finance/summary", dashboardHandler.GetFinanceSummary)
			adminAuth.GET("/finance/daily", dashboardHandler.GetDailyFinanceStats)

			// 数据统计
			adminAuth.GET("/stats/overview", adminHandler.GetStatisticsOverview)
			adminAuth.GET("/stats/orders", adminHandler.GetOrderStatistics)
			adminAuth.GET("/stats/revenue", adminHandler.GetIncomeStatistics)
		}
	}

	// ============ 外部 API 路由（面向下游/插件） ============
	api := r.Group("/api")
	{
		// KYC认证API（API Key 鉴权 + 服务访问权限/企业实名校验）
		kyc := api.Group("/kyc",
			middleware.APIKeyMiddleware(userRepo, apiRepo, signMgr),
			middleware.ServiceAccessGuard("kyc", 2),
		)
		{
			kyc.POST("/start", authHandler.StartAuth)
			kyc.POST("/result", authHandler.GetAuthResult)
			kyc.POST("/balance/query", authHandler.QueryBalance)
		}

		// 回调接口（无需认证，通过签名验证）
		callback := api.Group("/callback")
		{
			callback.POST("/finauth", callbackHandler.FinAuthCallback)
			callback.POST("/alipay", callbackHandler.AlipayCallback)
			callback.POST("/wechat", callbackHandler.WeChatCallback)
		}
	}

	return r, authService, balanceService
}
