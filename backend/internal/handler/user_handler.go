package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"starloftrpa/internal/model"
	"starloftrpa/internal/service"
	"starloftrpa/internal/utils"
)

var validator = utils.NewInputValidator()

type UserHandler struct {
	userService    *service.UserService
	smsService     *service.SMSService
	captchaService *service.CaptchaService
	jwtManager     *utils.JWTManager
}

func NewUserHandler(
	userService *service.UserService,
	smsService *service.SMSService,
	captchaService *service.CaptchaService,
	jwtManager *utils.JWTManager,
) *UserHandler {
	return &UserHandler{
		userService:    userService,
		smsService:     smsService,
		captchaService: captchaService,
		jwtManager:     jwtManager,
	}
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
	err := h.captchaService.VerifyCaptcha(req.CaptchaTicket, req.CaptchaRand, remoteIP)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "captcha verification failed",
		})
		return
	}

	// 发送短信验证码
	err = h.smsService.SendVerificationCode(req.Phone)
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

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Phone          string `json:"phone" binding:"required"`
		SMSCode        string `json:"sms_code" binding:"required"`
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

	// 验证人机验证码
	remoteIP := c.ClientIP()
	err := h.captchaService.VerifyCaptcha(req.CaptchaTicket, req.CaptchaRandstr, remoteIP)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "captcha verification failed",
		})
		return
	}

	// 验证短信验证码
	valid, err := h.smsService.VerifyCode(req.Phone, req.SMSCode)
	if err != nil || !valid {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "短信验证码错误或已过期",
		})
		return
	}

	// 注册用户
	user, err := h.userService.Register(req.Phone, req.Password)
	if err != nil {
		if err == service.ErrPhoneAlreadyExists {
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"message": "phone already exists",
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
			"token":     token,
			"expire_in": 86400,
		},
	})
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Phone          string `json:"phone" binding:"required"`
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

	// 验证手机号格式
	if err := validator.ValidatePhone(req.Phone); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid phone number format",
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

	// 检测SQL注入
	if err := validator.DetectSQLInjection(req.Phone); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid input detected",
		})
		return
	}

	// 验证人机验证码
	remoteIP := c.ClientIP()
	err := h.captchaService.VerifyCaptcha(req.CaptchaTicket, req.CaptchaRandstr, remoteIP)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "captcha verification failed",
		})
		return
	}

	// 根据登录方式验证
	var user *model.PlatformUser
	if req.LoginType == "password" {
		user, err = h.userService.Login(req.Phone, req.Password)
	} else if req.LoginType == "sms_code" {
		// 验证短信验证码
		valid, verr := h.smsService.VerifyCode(req.Phone, req.SMSCode)
		if verr != nil || !valid {
			c.JSON(http.StatusOK, gin.H{
				"code":    401,
				"message": "短信验证码错误或已过期",
			})
			return
		}
		user, err = h.userService.GetUserByPhone(req.Phone)
		if err != nil {
			err = service.ErrInvalidPassword
		} else if user.Status == 0 {
			err = service.ErrUserDisabled
		}
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "invalid login type",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": err.Error(),
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
			"user_id":    user.ID,
			"phone":      user.Phone,
			"token":      token,
			"expire_in":  86400,
			"api_key":    user.APIKey,
			"api_secret": user.APISecret,
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
	kycIDCard := ""
	if user.IsKYCVerified == 1 {
		if user.KYCName.Valid {
			kycName = maskName(user.KYCName.String)
		}
		if user.KYCIDCard.Valid {
			kycIDCard = maskIDCard(user.KYCIDCard.String)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"user_id":         user.ID,
			"phone":           user.Phone,
			"is_kyc_verified": user.IsKYCVerified,
			"kyc_name":        kycName,
			"kyc_id_card":     kycIDCard,
			"balance":         user.Balance,
			"api_key":         user.APIKey,
			// api_secret 仅返回给用户本人（用于 API 管理页展示/复制）；实名前为空字符串
			"api_secret": user.APISecret,
			"created_at": user.CreatedAt,
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
	if user.IsKYCVerified != 1 {
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
	err := h.captchaService.VerifyCaptcha(req.CaptchaTicket, req.CaptchaRandstr, remoteIP)
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
	valid, err := h.smsService.VerifyCode(user.Phone, req.SMSCode)
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
