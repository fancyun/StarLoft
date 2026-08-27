package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	FinAuth  FinAuthConfig
	Tencent  TencentConfig
	UnionPay UnionPayConfig
	Email    EmailConfig
	Log      LogConfig
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
	DBName       string
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
	ReturnURL string
	NotifyURL string
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

// UnionPayConfig 银联商务天满支付配置
type UnionPayConfig struct {
	MerchantNo  string
	TerminalNo  string
	AccessToken string
	SignKey     string
	ApiUrl      string
	NotifyURL   string
}

// EncryptionConfig 数据加密配置
type EncryptionConfig struct {
	Key string
}

// Load 从环境变量加载所有配置
func Load() (*Config, error) {
	cfg := &Config{}

	// 从环境变量加载所有配置
	loadFromEnv(cfg)

	return cfg, nil
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
	if dbname := os.Getenv("DB_NAME"); dbname != "" {
		cfg.Database.DBName = dbname
	}
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
	if returnURL := os.Getenv("FINAUTH_RETURN_URL"); returnURL != "" {
		cfg.FinAuth.ReturnURL = returnURL
	}
	if notifyURL := os.Getenv("FINAUTH_NOTIFY_URL"); notifyURL != "" {
		cfg.FinAuth.NotifyURL = notifyURL
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

	// 银联支付配置
	if merchantNo := os.Getenv("UNIONPAY_MERCHANT_NO"); merchantNo != "" {
		cfg.UnionPay.MerchantNo = merchantNo
	}
	if terminalNo := os.Getenv("UNIONPAY_TERMINAL_NO"); terminalNo != "" {
		cfg.UnionPay.TerminalNo = terminalNo
	}
	if accessToken := os.Getenv("UNIONPAY_ACCESS_TOKEN"); accessToken != "" {
		cfg.UnionPay.AccessToken = accessToken
	}
	if signKey := os.Getenv("UNIONPAY_SIGN_KEY"); signKey != "" {
		cfg.UnionPay.SignKey = signKey
	}
	if apiURL := os.Getenv("UNIONPAY_API_URL"); apiURL != "" {
		cfg.UnionPay.ApiUrl = apiURL
	}
	if notifyURL := os.Getenv("UNIONPAY_NOTIFY_URL"); notifyURL != "" {
		cfg.UnionPay.NotifyURL = notifyURL
	}

	// 加密密钥
	if encKey := os.Getenv("ENCRYPTION_KEY"); encKey != "" {
		// 注意：需要确保Config结构体有Encryption字段
		// 如果没有，需要添加
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
}
