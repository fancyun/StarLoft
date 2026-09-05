package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"starloftrpa/internal/model"
	"starloftrpa/internal/repository"
	"starloftrpa/internal/runtime"
	"starloftrpa/internal/service"
	"starloftrpa/internal/utils"
)

var validator = utils.NewInputValidator()

type UserHandler struct {
	userService  *service.UserService
	rt           *runtime.Runtime
	jwtManager   *utils.JWTManager
	authService  *service.AuthService
	loginLogRepo *repository.LoginLogRepository
}

func NewUserHandler(
	userService *service.UserService,
	rt *runtime.Runtime,
	jwtManager *utils.JWTManager,
	authService *service.AuthService,
	loginLogRepo *repository.LoginLogRepository,
) *UserHandler {
	return &UserHandler{
		userService:  userService,
		rt:           rt,
		jwtManager:   jwtManager,
		authService:  authService,
		loginLogRepo: loginLogRepo,
	}
}

func (h *UserHandler) sms() *service.SMSService     { return h.rt.SMS() }
func (h *UserHandler) email() *service.EmailService { return h.rt.Email() }
func (h *UserHandler) captcha() *service.CaptchaService {
	return h.rt.Captcha()
}

// SendCode 发送短信验证码
func (h *UserHandler) SendCode(c *gin.Context) {
	var req struct {
		Phone         string `json:"phone" binding:"required"`
		CaptchaTicket string `json:"captcha_ticket" binding:"required"`  // 腾讯验证码票据
		CaptchaRand   string `json:"captcha_randstr" binding:"required"` // 腾讯验证码随机串
		Scene         string `json:"scene" binding:"required"`           // register / login / change_password
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	// 验证人机验证码
	remoteIP := c.ClientIP()
	err := h.captcha().VerifyCaptcha(req.CaptchaTicket, req.CaptchaRand, remoteIP)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "captcha verification failed",
		})
		return
	}

	// 发送短信验证码
	err = h.sms().SendVerificationCode(req.Phone)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	// 验证码应存入 Redis，设置 5 分钟过期（当前版本使用第三方短信服务商的验证功能）

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"expire_in": 300,
		},
	})
}

// SendEmailCode 发送邮箱验证码
func (h *UserHandler) SendEmailCode(c *gin.Context) {
	var req struct {
		Email          string `json:"email" binding:"required"`
		CaptchaTicket  string `json:"captcha_ticket" binding:"required"`  // 腾讯验证码票据
		CaptchaRandstr string `json:"captcha_randstr" binding:"required"` // 腾讯验证码随机串
		Scene          string `json:"scene" binding:"required"`           // register
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	// 校验邮箱格式
	if !service.ValidateEmail(req.Email) {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "邮箱格式不正确",
		})
		return
	}

	// 验证人机验证码
	remoteIP := c.ClientIP()
	err := h.captcha().VerifyCaptcha(req.CaptchaTicket, req.CaptchaRandstr, remoteIP)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "captcha verification failed",
		})
		return
	}

	// 邮件服务未配置时不允许发送
	if !h.email().Enabled() {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "邮件服务未配置，无法发送邮箱验证码",
		})
		return
	}

	// 发送邮箱验证码
	err = h.email().SendVerificationCode(req.Email)
	if err != nil {
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
			"expire_in": 300,
		},
	})
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Phone          string `json:"phone" binding:"required"`
		Username       string `json:"username" binding:"required"`
		Email          string `json:"email" binding:"required"`
		SMSCode        string `json:"sms_code" binding:"required"`
		EmailCode      string `json:"email_code" binding:"required"`
		Password       string `json:"password" binding:"required"`
		CaptchaTicket  string `json:"captcha_ticket" binding:"required"`  // 腾讯验证码票据
		CaptchaRandstr string `json:"captcha_randstr" binding:"required"` // 腾讯验证码随机串
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	// 校验用户名格式（仅支持英文+数字+下划线）
	if !service.ValidateUsername(req.Username) {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "用户名仅支持英文、数字、下划线，长度3-32位",
		})
		return
	}

	// 校验邮箱格式
	if !service.ValidateEmail(req.Email) {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "邮箱格式不正确",
		})
		return
	}

	// 验证人机验证码
	remoteIP := c.ClientIP()
	err := h.captcha().VerifyCaptcha(req.CaptchaTicket, req.CaptchaRandstr, remoteIP)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "captcha verification failed",
		})
		return
	}

	// 验证短信验证码
	valid, err := h.sms().VerifyCode(req.Phone, req.SMSCode)
	if err != nil || !valid {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "短信验证码错误或已过期",
		})
		return
	}

	// 验证邮箱验证码
	if !h.email().Enabled() {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "邮件服务未配置，无法注册",
		})
		return
	}
	valid, err = h.email().VerifyCode(req.Email, req.EmailCode)
	if err != nil || !valid {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "邮箱验证码错误或已过期",
		})
		return
	}

	// 注册用户
	user, err := h.userService.Register(req.Phone, req.Username, req.Email, req.Password)
	if err != nil {
		switch err {
		case service.ErrPhoneAlreadyExists:
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"message": "该手机号已被注册",
			})
			return
		case service.ErrUsernameAlreadyExists:
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"message": "该用户名已被占用",
			})
			return
		case service.ErrEmailAlreadyExists:
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"message": "该邮箱已被注册",
			})
			return
		case service.ErrInvalidUsername:
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"message": "用户名仅支持英文、数字、下划线，长度3-32位",
			})
			return
		case service.ErrInvalidEmail:
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"message": "邮箱格式不正确",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "registration failed",
		})
		return
	}

	// 生成 Token
	token, err := h.jwtManager.GenerateToken(user.ID, user.Phone, "user", 24*time.Hour)
	if err != nil {
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
			"user_id":   user.ID,
			"phone":     user.Phone,
			"username":  user.Username,
			"email":     user.Email,
			"token":     token,
			"expire_in": 86400,
		},
	})
}

// Login 用户登录（支持用户名/手机号/邮箱）
func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Account        string `json:"account" binding:"required"`
		Password       string `json:"password"`
		SMSCode        string `json:"sms_code"`
		LoginType      string `json:"login_type" binding:"required"`      // password / sms_code
		CaptchaTicket  string `json:"captcha_ticket" binding:"required"`  // 腾讯验证码票据
		CaptchaRandstr string `json:"captcha_randstr" binding:"required"` // 腾讯验证码随机串
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	// 检测SQL注入
	if err := validator.DetectSQLInjection(req.Account); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid input detected",
		})
		return
	}

	// 验证登录类型
	if req.LoginType != "password" && req.LoginType != "sms_code" {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid login type",
		})
		return
	}

	// 验证码登录必须使用手机号（短信发送依赖手机号）
	if req.LoginType == "sms_code" {
		if err := validator.ValidatePhone(req.Account); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"message": "验证码登录请输入正确的手机号",
			})
			return
		}
	}

	// 验证人机验证码
	remoteIP := c.ClientIP()
	err := h.captcha().VerifyCaptcha(req.CaptchaTicket, req.CaptchaRandstr, remoteIP)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "captcha verification failed",
		})
		return
	}

	// 根据登录方式验证（login_type 已在上方校验，仅 password/sms_code 两种）
	var user *model.User
	var loginErr error
	switch req.LoginType {
	case "password":
		user, loginErr = h.userService.Login(req.Account, req.Password)
	case "sms_code":
		// 验证短信验证码
		valid, verr := h.sms().VerifyCode(req.Account, req.SMSCode)
		if verr != nil || !valid {
			loginErr = errors.New("短信验证码错误或已过期")
		} else {
			user, loginErr = h.userService.GetUserByPhone(req.Account)
			if loginErr != nil {
				loginErr = service.ErrInvalidPassword
			} else if user.Status == 0 {
				loginErr = service.ErrUserDisabled
			}
		}
	default:
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid login type",
		})
		return
	}

	// 记录登录日志（成功与失败均记录）
	if loginErr != nil {
		h.loginLogRepo.InsertUserLoginLog(&model.UserLoginLog{
			Account:    req.Account,
			LoginType:  req.LoginType,
			IP:         c.ClientIP(),
			UserAgent:  c.GetHeader("User-Agent"),
			Status:     0,
			FailReason: loginErr.Error(),
		})
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": loginErr.Error(),
		})
		return
	}

	h.loginLogRepo.InsertUserLoginLog(&model.UserLoginLog{
		UserID:    user.ID,
		Account:   req.Account,
		LoginType: req.LoginType,
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Status:    1,
	})

	// 生成 Token
	token, err := h.jwtManager.GenerateToken(user.ID, user.Phone, "user", 24*time.Hour)
	if err != nil {
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
			"user_id":    user.ID,
			"phone":      user.Phone,
			"username":   user.Username,
			"email":      user.Email,
			"token":      token,
			"expire_in":  86400,
		},
	})
}

// GetProfile 获取用户信息
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetInt64("user_id")

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "user not found",
		})
		return
	}

	// 脱敏处理
	kycName := ""
	kycNumber := ""
	if user.IsKYCVerified == 1 || user.IsKYCVerified == 2 {
		if user.KYCName.Valid {
			kycName = maskName(user.KYCName.String)
		}
		if user.KYCNumber.Valid {
			kycNumber = maskIDCard(user.KYCNumber.String)
		}
	}

	// 账户实名剩余免费认证次数（终身累计失败达3次后转为计费）
	freeAuthRemaining := 0
	if h.authService != nil {
		if remaining, err := h.authService.GetFreeAuthRemaining(userID); err == nil {
			freeAuthRemaining = remaining
		}
	}

	// 当前生效的 API 密钥（密钥对存于 api 表；完整 api_secret 仅返回给用户本人展示/复制）
	apiKey, apiSecret, apiPermission := "", "", ""
	if ak, err := h.userService.GetAPIKeyByUser(userID); err == nil {
		apiKey = ak.APIKey
		apiPermission = ak.Permission
		if user.IsKYCVerified > 0 {
			apiSecret = ak.APISecret
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"user_id":         user.ID,
			"phone":           user.Phone,
			"username":        user.Username,
			"email":           user.Email,
			"is_kyc_verified": user.IsKYCVerified,
			"kyc_name":        kycName,
			"kyc_number":      kycNumber,
			"balance":         user.Balance,
			"api_key":         apiKey,
			"api_secret":      apiSecret,
			"api_permission":  apiPermission,
			"created_at": user.CreatedAt,
			// 账户实名免费认证次数（free_auth_limit 终身免费上限，free_auth_remaining 剩余免费次数）
			"free_auth_limit":     3,
			"free_auth_remaining": freeAuthRemaining,
		},
	})
}

// ResetAPIKey 重置 API 密钥（需先完成账户实名认证）
func (h *UserHandler) ResetAPIKey(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// 开通 API 需先完成实名认证
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "user not found",
		})
		return
	}
	if user.IsKYCVerified == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    403,
			"message": "请先完成实名认证后开通 API",
		})
		return
	}

	apiKey, apiSecret, err := h.userService.ResetAPIKey(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "failed to reset api key",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"api_key":    apiKey,
			"api_secret": apiSecret,
		},
	})
}

// SetAPIKeyPermission 设置 API 密钥权限范围（all-全部服务，或单个服务标识如 kyc）
func (h *UserHandler) SetAPIKeyPermission(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		Permission string `json:"permission" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "invalid request parameters"})
		return
	}

	if err := h.userService.SetAPIKeyPermission(userID, req.Permission); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "设置 API 密钥权限失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"permission": req.Permission}})
}

// ChangePassword 修改密码
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		SMSCode        string `json:"sms_code" binding:"required"`
		NewPassword    string `json:"new_password" binding:"required"`
		CaptchaTicket  string `json:"captcha_ticket" binding:"required"`  // 腾讯验证码票据
		CaptchaRandstr string `json:"captcha_randstr" binding:"required"` // 腾讯验证码随机串
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid request parameters",
		})
		return
	}

	// 验证新密码强度
	if err := validator.ValidatePassword(req.NewPassword); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "password must be 8-32 characters with letters and numbers",
		})
		return
	}

	// 验证人机验证码
	remoteIP := c.ClientIP()
	err := h.captcha().VerifyCaptcha(req.CaptchaTicket, req.CaptchaRandstr, remoteIP)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "captcha verification failed",
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
	valid, err := h.sms().VerifyCode(user.Phone, req.SMSCode)
	if err != nil || !valid {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "短信验证码错误或已过期",
		})
		return
	}

	err = h.userService.ChangePassword(userID, req.NewPassword)
	if err != nil {
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

// 辅助函数
func maskName(name string) string {
	if len(name) == 0 {
		return ""
	}
	if len(name) == 1 {
		return name
	}
	if len(name) == 2 {
		return name[0:1] + "*"
	}
	return name[0:1] + "**" + name[len(name)-1:]
}

func maskIDCard(idCard string) string {
	if len(idCard) < 8 {
		return idCard
	}
	return idCard[0:3] + "***********" + idCard[len(idCard)-4:]
}
