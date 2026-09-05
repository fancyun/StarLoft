package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"starloftrpa/internal/model"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	FinAuth   FinAuthConfig
	Tencent   TencentConfig
	Alipay    AlipayConfig
	WeChatPay WeChatPayConfig
	Email     EmailConfig
	Log       LogConfig
	KycPrice  float64 // 平台 KYC 认证单价（元/次），API 认证计费与前端展示使用
}

// EmailConfig 腾讯云 SES 邮件发送配置（用于发送邮箱验证码）
type EmailConfig struct {
	From       string
	TemplateID uint64
	Region     string
}

// LogConfig 日志配置
type LogConfig struct {
	Dir string
}

type ServerConfig struct {
	Host string
	Port string
	Mode string
}

type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string // 系统库名（连接默认库）
	KycDBName    string // 实名认证产品库名（跨库访问）
	MaxIdleConns int
	MaxOpenConns int
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
}

type JWTConfig struct {
	Secret      string
	ExpireHours int
	AdminSecret string
}

type FinAuthConfig struct {
	APIKey    string
	APISecret string
	SceneID   string
	BaseURL   string
}

type TencentConfig struct {
	SecretID  string
	SecretKey string
	SMS       TencentSMSConfig
	Captcha   TencentCaptchaConfig
}

type TencentSMSConfig struct {
	SDKAppID   string
	SignName   string
	TemplateID string
}

type TencentCaptchaConfig struct {
	CaptchaAppID string
	AppSecretKey string
}

// AlipayConfig 支付宝开放平台支付配置（电脑网站支付 alipay.trade.page.pay）
type AlipayConfig struct {
	AppID      string // 应用AppID
	PrivateKey string // 应用私钥（RSA2，PEM）
	PublicKey  string // 支付宝公钥（PEM，用于回调验签）
}

// WeChatPayConfig 微信支付（APIv3 Native支付）配置
type WeChatPayConfig struct {
	AppID          string // 商户绑定的AppID
	MchID          string // 商户号
	APIv3Key       string // APIv3密钥（用于回调报文解密）
	MchSerialNo    string // 商户API证书序列号
	MchPrivateKey  string // 商户API私钥（PEM）
	PlatformPubKey string // 微信支付公钥（PEM，用于回调验签）
}

// Load 从环境变量加载所有配置
func Load() (*Config, error) {
	cfg := &Config{}

	// 从环境变量加载所有配置
	loadFromEnv(cfg)

	// 校验关键密钥（启动早失败，避免带空/弱密钥运行泄露账户或数据）
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate 校验关键必填密钥非空且满足最小长度
func (cfg *Config) validate() error {
	checks := []struct {
		name string
		val  string
		min  int
	}{
		{"JWT_SECRET", cfg.JWT.Secret, 32},
		{"JWT_ADMIN_SECRET", cfg.JWT.AdminSecret, 32},
		{"DB_PASSWORD", cfg.Database.Password, 8},
		{"REDIS_PASSWORD", cfg.Redis.Password, 8},
		{"TENCENT_SECRET_ID", cfg.Tencent.SecretID, 8},
		{"TENCENT_SECRET_KEY", cfg.Tencent.SecretKey, 16},
	}
	for _, c := range checks {
		if len(c.val) < c.min {
			return fmt.Errorf("关键配置 %s 缺失或强度不足：至少需要 %d 个字符", c.name, c.min)
		}
	}
	return nil
}

// loadFromEnv 从环境变量加载所有配置
func loadFromEnv(cfg *Config) {
	// 服务器配置
	if host := os.Getenv("SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		cfg.Server.Port = port
	}
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		cfg.Server.Mode = mode
	}

	// 数据库配置
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		// 转换为int
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Database.Port = p
		}
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		cfg.Database.Password = password
	}
	// 分库库名写死为固定常量（与 init.sql 建库一致，不支持通过环境变量修改）
	cfg.Database.DBName = model.SysDB
	cfg.Database.KycDBName = model.KycDB
	if maxIdle := os.Getenv("DB_MAX_IDLE_CONNS"); maxIdle != "" {
		if n, err := strconv.Atoi(maxIdle); err == nil {
			cfg.Database.MaxIdleConns = n
		}
	}
	if maxOpen := os.Getenv("DB_MAX_OPEN_CONNS"); maxOpen != "" {
		if n, err := strconv.Atoi(maxOpen); err == nil {
			cfg.Database.MaxOpenConns = n
		}
	}

	// Redis配置
	if host := os.Getenv("REDIS_HOST"); host != "" {
		cfg.Redis.Host = host
	}
	if port := os.Getenv("REDIS_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Redis.Port = p
		}
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		cfg.Redis.Password = password
	}
	if db := os.Getenv("REDIS_DB"); db != "" {
		if d, err := strconv.Atoi(db); err == nil {
			cfg.Redis.DB = d
		}
	}
	if poolSize := os.Getenv("REDIS_POOL_SIZE"); poolSize != "" {
		if n, err := strconv.Atoi(poolSize); err == nil {
			cfg.Redis.PoolSize = n
		}
	}

	// JWT配置
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.JWT.Secret = secret
	}
	if adminSecret := os.Getenv("JWT_ADMIN_SECRET"); adminSecret != "" {
		cfg.JWT.AdminSecret = adminSecret
	}
	if expireHours := os.Getenv("JWT_EXPIRE_HOURS"); expireHours != "" {
		if n, err := strconv.Atoi(expireHours); err == nil {
			cfg.JWT.ExpireHours = n
		}
	}

	// FinAuth配置
	if apiKey := os.Getenv("FINAUTH_API_KEY"); apiKey != "" {
		cfg.FinAuth.APIKey = apiKey
	}
	if apiSecret := os.Getenv("FINAUTH_API_SECRET"); apiSecret != "" {
		cfg.FinAuth.APISecret = apiSecret
	}
	if sceneID := os.Getenv("FINAUTH_SCENE_ID"); sceneID != "" {
		cfg.FinAuth.SceneID = sceneID
	}
	if baseURL := os.Getenv("FINAUTH_BASE_URL"); baseURL != "" {
		cfg.FinAuth.BaseURL = baseURL
	}

	// 腾讯云配置
	if secretID := os.Getenv("TENCENT_SECRET_ID"); secretID != "" {
		cfg.Tencent.SecretID = secretID
	}
	if secretKey := os.Getenv("TENCENT_SECRET_KEY"); secretKey != "" {
		cfg.Tencent.SecretKey = secretKey
	}
	if signName := os.Getenv("TENCENT_SMS_SIGN_NAME"); signName != "" {
		cfg.Tencent.SMS.SignName = signName
	}
	if sdkAppID := os.Getenv("TENCENT_SMS_SDK_APP_ID"); sdkAppID != "" {
		cfg.Tencent.SMS.SDKAppID = sdkAppID
	}
	if templateID := os.Getenv("TENCENT_SMS_TEMPLATE_ID"); templateID != "" {
		cfg.Tencent.SMS.TemplateID = templateID
	}
	if captchaAppID := os.Getenv("TENCENT_CAPTCHA_APP_ID"); captchaAppID != "" {
		cfg.Tencent.Captcha.CaptchaAppID = captchaAppID
	}
	if appSecretKey := os.Getenv("TENCENT_CAPTCHA_SECRET"); appSecretKey != "" {
		cfg.Tencent.Captcha.AppSecretKey = appSecretKey
	}

	// 支付宝支付配置
	if v := os.Getenv("ALIPAY_APP_ID"); v != "" {
		cfg.Alipay.AppID = v
	}
	if v := os.Getenv("ALIPAY_PRIVATE_KEY"); v != "" {
		cfg.Alipay.PrivateKey = normalizePEM(v)
	}
	if v := os.Getenv("ALIPAY_PUBLIC_KEY"); v != "" {
		cfg.Alipay.PublicKey = normalizePEM(v)
	}

	// 微信支付配置
	if v := os.Getenv("WECHAT_APP_ID"); v != "" {
		cfg.WeChatPay.AppID = v
	}
	if v := os.Getenv("WECHAT_MCH_ID"); v != "" {
		cfg.WeChatPay.MchID = v
	}
	if v := os.Getenv("WECHAT_API_V3_KEY"); v != "" {
		cfg.WeChatPay.APIv3Key = v
	}
	if v := os.Getenv("WECHAT_MCH_SERIAL_NO"); v != "" {
		cfg.WeChatPay.MchSerialNo = v
	}
	if v := os.Getenv("WECHAT_MCH_PRIVATE_KEY"); v != "" {
		cfg.WeChatPay.MchPrivateKey = normalizePEM(v)
	}
	if v := os.Getenv("WECHAT_PLATFORM_PUBLIC_KEY"); v != "" {
		cfg.WeChatPay.PlatformPubKey = normalizePEM(v)
	}

	// 邮件（腾讯云 SES）配置
	if from := os.Getenv("SES_FROM"); from != "" {
		cfg.Email.From = from
	}
	if templateID := os.Getenv("SES_TEMPLATE_ID"); templateID != "" {
		if n, err := strconv.ParseUint(templateID, 10, 64); err == nil {
			cfg.Email.TemplateID = n
		}
	}
	if region := os.Getenv("SES_REGION"); region != "" {
		cfg.Email.Region = region
	}

	// 日志配置
	if dir := os.Getenv("LOG_DIR"); dir != "" {
		cfg.Log.Dir = dir
	}

	// 平台 KYC 认证单价（元/次），未配置时兜底 1.00
	if kycPrice := os.Getenv("KYC_PRICE"); kycPrice != "" {
		if n, err := strconv.ParseFloat(kycPrice, 64); err == nil && n > 0 {
			cfg.KycPrice = n
		}
	}
}

// normalizePEM 将 PEM 内容中的字面 \n 还原为换行
// （.env 与 docker-compose 环境变量中的多行 PEM 需以单行 \n 形式书写）
func normalizePEM(s string) string {
	return strings.ReplaceAll(s, "\\n", "\n")
}
