package router

import (
	"starloftrpa/internal/config"
	"starloftrpa/internal/database"
	"starloftrpa/internal/handler"
	"starloftrpa/internal/middleware"
	"starloftrpa/internal/repository"
	"starloftrpa/internal/service"
	"starloftrpa/internal/upstream"
	"starloftrpa/internal/utils"

	"github.com/gin-gonic/gin"
)

func Setup(cfg *config.Config) (*gin.Engine, *service.AuthService) {
	r := gin.New()

	// 全局中间件
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.Recovery())

	// 获取数据库连接
	db := database.DB

	// 初始化 Repository
	userRepo := repository.NewUserRepository(db)
	authRepo := repository.NewAuthOrderRepository(db)
	paymentRepo := repository.NewPaymentOrderRepository(db)
	balanceLogRepo := repository.NewBalanceLogRepository(db)
	adminRepo := repository.NewAdminRepository(db)
	configRepo := repository.NewSystemConfigRepository(db)
	kycRecordRepo := repository.NewKycRecordRepository(db)
	resourcePackRepo := repository.NewResourcePackRepository(db)
	loginLogRepo := repository.NewLoginLogRepository(db)

	// 初始化上游服务客户端
	finAuthClient := upstream.NewFinAuthClient(
		cfg.FinAuth.BaseURL,
		cfg.FinAuth.APIKey,
		cfg.FinAuth.APISecret,
	)

	// 初始化支付客户端
	alipayClient, err := upstream.NewAlipayClient(
		cfg.Alipay.AppID,
		cfg.Alipay.PrivateKey,
		cfg.Alipay.PublicKey,
	)
	if err != nil {
		panic("初始化支付宝支付客户端失败: " + err.Error())
	}

	wechatClient, err := upstream.NewWeChatPayClient(
		cfg.WeChatPay.AppID,
		cfg.WeChatPay.MchID,
		cfg.WeChatPay.APIv3Key,
		cfg.WeChatPay.MchSerialNo,
		cfg.WeChatPay.MchPrivateKey,
		cfg.WeChatPay.PlatformPubKey,
	)
	if err != nil {
		panic("初始化微信支付客户端失败: " + err.Error())
	}

	// 初始化短信服务
	smsService, err := service.NewSMSService(
		cfg.Tencent.SecretID,
		cfg.Tencent.SecretKey,
		cfg.Tencent.SMS.SDKAppID,
		cfg.Tencent.SMS.SignName,
		cfg.Tencent.SMS.TemplateID,
	)
	if err != nil {
		panic("初始化短信服务失败: " + err.Error())
	}

	// 初始化验证码服务
	captchaService := service.NewCaptchaService(
		cfg.Tencent.SecretID,
		cfg.Tencent.SecretKey,
		cfg.Tencent.Captcha.CaptchaAppID,
		cfg.Tencent.Captcha.AppSecretKey,
	)

	// 初始化邮箱服务（腾讯云 SES，未配置时返回 nil，不启用邮箱验证码）
	emailService := service.NewEmailService(
		cfg.Tencent.SecretID,
		cfg.Tencent.SecretKey,
		cfg.Email.From,
		cfg.Email.TemplateID,
		cfg.Email.Region,
	)

	// 初始化 Service
	userService := service.NewUserService(userRepo, configRepo)
	balanceService := service.NewBalanceService(userRepo, balanceLogRepo, paymentRepo, resourcePackRepo, configRepo, db)

	authService := service.NewAuthService(
		finAuthClient,
		authRepo,
		userRepo,
		kycRecordRepo,
		resourcePackRepo,
		balanceService,
		configRepo,
		&cfg.FinAuth,
	)

	// 初始化 JWT Manager
	jwtManager := utils.NewJWTManager(cfg.JWT.Secret)
	signMgr := utils.NewSignatureManager()

	// 初始化 Handler
	publicHandler := handler.NewPublicHandler(cfg, configRepo)

	userHandler := handler.NewUserHandler(
		userService,
		smsService,
		emailService,
		captchaService,
		jwtManager,
		authService,
		loginLogRepo,
	)

	authHandler := handler.NewAuthHandler(authService, balanceService, resourcePackRepo, alipayClient, wechatClient)

	adminHandler := handler.NewAdminHandler(adminRepo,
		userRepo,
		authRepo,
		paymentRepo,
		configRepo,
		balanceLogRepo,
		resourcePackRepo,
		loginLogRepo,
		balanceService,
		authService,
		cfg.JWT.AdminSecret,
	)

	callbackHandler := handler.NewCallbackHandler(
		authService,
		balanceService,
		paymentRepo,
		alipayClient,
		wechatClient,
	)

	dashboardHandler := handler.NewDashboardHandler(db)

	// 路由注册

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 公开接口（无需认证）
		public := v1.Group("/public")
		{
			public.GET("/config", publicHandler.GetPublicConfig)
			public.GET("/qr", publicHandler.GetQRCode)
		}

		// 用户相关路由（Web前端调用，JWT鉴权）
		user := v1.Group("/user")
		{
			user.POST("/send-code", userHandler.SendCode)
			user.POST("/send-email-code", userHandler.SendEmailCode)
			user.POST("/register", userHandler.Register)
			user.POST("/login", userHandler.Login)

			// 需要JWT认证的路由
			auth := user.Group("", middleware.JWTAuth(cfg.JWT.Secret))
			{
				auth.GET("/profile", userHandler.GetProfile)
				auth.GET("/kyc/status", authHandler.GetUserAuthStatus)
				auth.POST("/kyc/sync", authHandler.SyncKycResult)
				auth.POST("/kyc", authHandler.StartAuthForWeb)
				auth.DELETE("/kyc", authHandler.CancelKycRecord)
				auth.GET("/records", authHandler.GetUserAuthRecords)
				auth.GET("/stats/calls", authHandler.GetUserAuthCallStats)
				auth.POST("/recharge", authHandler.CreateRecharge)
				auth.GET("/recharge/result", authHandler.GetRechargeResult)
				auth.POST("/api-key/reset", userHandler.ResetAPIKey)
				auth.POST("/change-password", userHandler.ChangePassword)
				// 资源包（使用余额购买，不支持直接为资源包付费）
				auth.GET("/packs", authHandler.ListResourcePacks)                  // 在售资源包列表
				auth.POST("/packs/:id/purchase", authHandler.PurchaseResourcePack) // 使用余额购买资源包
				auth.GET("/packs/mine", authHandler.MyResourcePacks)               // 我的资源包
			}
		}

		// KYC认证API（API Key鉴权）
		kyc := v1.Group("/kyc", middleware.APIKeyMiddleware(userRepo, signMgr))
		{
			kyc.POST("/start", authHandler.StartAuth)
			kyc.POST("/result", authHandler.GetAuthResult)
			kyc.POST("/balance/query", authHandler.QueryBalance)
		}

		// 管理后台路由
		admin := v1.Group("/admin")
		{
			// 管理员登录 - 限流每分钟5次
			admin.POST("/login", middleware.RateLimiterForIP(5), adminHandler.AdminLogin)

			// 需要管理员JWT认证的路由 - 每分钟100次
			adminAuth := admin.Group("", middleware.JWTAuth(cfg.JWT.AdminSecret), middleware.RateLimiter(100))
			{
				// 用户管理
				adminAuth.GET("/users", adminHandler.GetUserList)
				adminAuth.GET("/users/:id", adminHandler.GetUserDetail)
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

				// 系统配置
				adminAuth.GET("/config", adminHandler.GetSystemConfig)
				adminAuth.PUT("/config", adminHandler.UpdateSystemConfig)

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

		// 回调接口（无需认证，通过签名验证）
		callback := v1.Group("/callback")
		{
			callback.POST("/finauth", callbackHandler.FinAuthCallback)
			callback.POST("/alipay", callbackHandler.AlipayCallback)
			callback.POST("/wechat", callbackHandler.WeChatCallback)
		}
	}

	return r, authService
}
