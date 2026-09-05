package runtime

import (
	"log"
	"sync"

	"starloftrpa/internal/config"
	"starloftrpa/internal/service"
	"starloftrpa/internal/upstream"
)

// Runtime 持有第三方业务配置对应的上游客户端快照。
// 业务配置全部来自环境变量（config.Load 加载），启动时一次性构建，不依赖数据库配置。
type Runtime struct {
	mu  sync.RWMutex
	snp *snapshot
}

// snapshot 当前生效的第三方业务客户端快照。
type snapshot struct {
	finAuth      upstream.FinAuthInterface
	finAuthCfg   config.FinAuthConfig
	alipay       *upstream.AlipayClient
	wechat       *upstream.WeChatPayClient
	sms          *service.SMSService
	email        *service.EmailService
	captcha      *service.CaptchaService
	captchaAppID string
	kycPrice     float64
}

// New 根据环境变量配置构建第三方业务客户端快照。
func New(cfg *config.Config) (*Runtime, error) {
	rt := &Runtime{}
	s := &snapshot{}

	// FinAuth
	s.finAuthCfg = cfg.FinAuth
	s.finAuth = upstream.NewFinAuthClient(s.finAuthCfg.BaseURL, s.finAuthCfg.APIKey, s.finAuthCfg.APISecret)

	// 腾讯云账号密钥（短信/验证码/邮件复用）
	secretID := cfg.Tencent.SecretID
	secretKey := cfg.Tencent.SecretKey

	// 短信
	smsSvc, err := service.NewSMSService(
		secretID,
		secretKey,
		cfg.Tencent.SMS.SDKAppID,
		cfg.Tencent.SMS.SignName,
		cfg.Tencent.SMS.TemplateID,
	)
	if err != nil {
		log.Printf("构建短信服务失败，短信功能暂时不可用: %v", err)
	} else {
		s.sms = smsSvc
	}

	// 人机验证码
	s.captchaAppID = cfg.Tencent.Captcha.CaptchaAppID
	s.captcha = service.NewCaptchaService(
		secretID,
		secretKey,
		s.captchaAppID,
		cfg.Tencent.Captcha.AppSecretKey,
	)

	// 邮件（腾讯云 SES，未配置时返回 nil 不启用）
	s.email = service.NewEmailService(
		secretID,
		secretKey,
		cfg.Email.From,
		cfg.Email.TemplateID,
		cfg.Email.Region,
	)

	// 支付宝
	if cfg.Alipay.AppID != "" && cfg.Alipay.PrivateKey != "" && cfg.Alipay.PublicKey != "" {
		alipayClient, e := upstream.NewAlipayClient(cfg.Alipay.AppID, cfg.Alipay.PrivateKey, cfg.Alipay.PublicKey)
		if e != nil {
			log.Printf("构建支付宝支付客户端失败，支付宝充值暂时不可用: %v", e)
		} else {
			s.alipay = alipayClient
		}
	}

	// 微信支付
	if cfg.WeChatPay.AppID != "" && cfg.WeChatPay.MchID != "" && cfg.WeChatPay.APIv3Key != "" &&
		cfg.WeChatPay.MchSerialNo != "" && cfg.WeChatPay.MchPrivateKey != "" && cfg.WeChatPay.PlatformPubKey != "" {
		wechatClient, e := upstream.NewWeChatPayClient(
			cfg.WeChatPay.AppID, cfg.WeChatPay.MchID, cfg.WeChatPay.APIv3Key,
			cfg.WeChatPay.MchSerialNo, cfg.WeChatPay.MchPrivateKey, cfg.WeChatPay.PlatformPubKey,
		)
		if e != nil {
			log.Printf("构建微信支付客户端失败，微信充值暂时不可用: %v", e)
		} else {
			s.wechat = wechatClient
		}
	}

	// 平台 KYC 认证单价
	s.kycPrice = cfg.KycPrice

	rt.snp = s
	return rt, nil
}

// FinAuth 返回当前生效的 FinAuth 客户端。
func (rt *Runtime) FinAuth() upstream.FinAuthInterface {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.snp.finAuth
}

// FinAuthCfg 返回当前生效的 FinAuth 配置。
func (rt *Runtime) FinAuthCfg() config.FinAuthConfig {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.snp.finAuthCfg
}

// Alipay 返回当前生效的支付宝客户端（未配置时为 nil）。
func (rt *Runtime) Alipay() *upstream.AlipayClient {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.snp.alipay
}

// Wechat 返回当前生效的微信支付客户端（未配置时为 nil）。
func (rt *Runtime) Wechat() *upstream.WeChatPayClient {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.snp.wechat
}

// SMS 返回当前生效的短信服务。
func (rt *Runtime) SMS() *service.SMSService {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.snp.sms
}

// Email 返回当前生效的邮件服务（未配置时为 nil）。
func (rt *Runtime) Email() *service.EmailService {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.snp.email
}

// Captcha 返回当前生效的人机验证码服务。
func (rt *Runtime) Captcha() *service.CaptchaService {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.snp.captcha
}

// CaptchaAppID 返回当前生效的验证码 AppID（供前端渲染验证码组件）。
func (rt *Runtime) CaptchaAppID() string {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.snp.captchaAppID
}

// KycPrice 返回平台 KYC 认证单价（元/次）。
func (rt *Runtime) KycPrice() float64 {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.snp.kycPrice
}
