package runtime

import (
	"log"
	"strconv"
	"strings"
	"sync"

	"starloftrpa/internal/config"
	"starloftrpa/internal/repository"
	"starloftrpa/internal/service"
	"starloftrpa/internal/upstream"
)

// Runtime 持有第三方业务配置对应的上游客户端快照。
// 管理员在后台修改业务配置后，可通过 Reload 从数据库重新读取并原子替换快照，即时生效（无需重启）。
// 数据库未配置的业务项兜底回退到环境变量（config.Load 加载的初始值）。
type Runtime struct {
	mu         sync.RWMutex
	configRepo *repository.SystemConfigRepository
	base       *config.Config
	snp        *snapshot
}

// snapshot 当前生效的第三方业务客户端快照，Reload 时整体替换。
type snapshot struct {
	finAuth      upstream.FinAuthInterface
	finAuthCfg   config.FinAuthConfig
	alipay       *upstream.AlipayClient
	wechat       *upstream.WeChatPayClient
	sms          *service.SMSService
	email        *service.EmailService
	captcha      *service.CaptchaService
	captchaAppID string
}

// New 创建运行时控制器，并根据数据库业务配置构建初始快照。
func New(cfg *config.Config, configRepo *repository.SystemConfigRepository) (*Runtime, error) {
	rt := &Runtime{
		configRepo: configRepo,
		base:       cfg,
	}
	if err := rt.Reload(); err != nil {
		return nil, err
	}
	return rt, nil
}

// Reload 从数据库重新读取第三方业务配置并重建客户端快照，原子替换即时生效。
func (rt *Runtime) Reload() error {
	dbConfigs, err := rt.configRepo.GetAllConfigs()
	if err != nil {
		return err
	}

	// get 优先取数据库值，未配置/为空时回退到环境变量初始值
	get := func(key, envFallback string) string {
		if v, ok := dbConfigs[key]; ok && v != "" {
			return v
		}
		return envFallback
	}

	b := rt.base
	s := &snapshot{}

	// FinAuth
	s.finAuthCfg = config.FinAuthConfig{
		APIKey:    get("finauth_api_key", b.FinAuth.APIKey),
		APISecret: get("finauth_api_secret", b.FinAuth.APISecret),
		SceneID:   get("finauth_scene_id", b.FinAuth.SceneID),
		BaseURL:   get("finauth_base_url", b.FinAuth.BaseURL),
	}
	s.finAuth = upstream.NewFinAuthClient(s.finAuthCfg.BaseURL, s.finAuthCfg.APIKey, s.finAuthCfg.APISecret)

	// 腾讯云账号密钥（短信/验证码/邮件复用）
	secretID := get("tencent_secret_id", b.Tencent.SecretID)
	secretKey := get("tencent_secret_key", b.Tencent.SecretKey)

	// 短信
	smsSvc, err := service.NewSMSService(
		secretID,
		secretKey,
		get("tencent_sms_sdk_app_id", b.Tencent.SMS.SDKAppID),
		get("tencent_sms_sign_name", b.Tencent.SMS.SignName),
		get("tencent_sms_template_id", b.Tencent.SMS.TemplateID),
	)
	if err != nil {
		log.Printf("构建短信服务失败，短信功能暂时不可用: %v", err)
	} else {
		s.sms = smsSvc
	}

	// 人机验证码
	s.captchaAppID = get("tencent_captcha_app_id", b.Tencent.Captcha.CaptchaAppID)
	s.captcha = service.NewCaptchaService(
		secretID,
		secretKey,
		s.captchaAppID,
		get("tencent_captcha_secret", b.Tencent.Captcha.AppSecretKey),
	)

	// 邮件（腾讯云 SES，未配置时返回 nil 不启用）
	tmplStr := get("ses_template_id", "")
	tmplID := b.Email.TemplateID
	if tmplStr != "" {
		if n, e := strconv.ParseUint(tmplStr, 10, 64); e == nil {
			tmplID = n
		}
	}
	s.email = service.NewEmailService(
		secretID,
		secretKey,
		get("ses_from", b.Email.From),
		tmplID,
		get("ses_region", b.Email.Region),
	)

	// 支付宝
	alipayAppID := get("alipay_app_id", b.Alipay.AppID)
	alipayPriv := normalizePEM(get("alipay_private_key", b.Alipay.PrivateKey))
	alipayPub := normalizePEM(get("alipay_public_key", b.Alipay.PublicKey))
	if alipayAppID != "" && alipayPriv != "" && alipayPub != "" {
		alipayClient, e := upstream.NewAlipayClient(alipayAppID, alipayPriv, alipayPub)
		if e != nil {
			log.Printf("重建支付宝支付客户端失败，支付宝充值暂时不可用: %v", e)
		} else {
			s.alipay = alipayClient
		}
	}

	// 微信支付
	wechatAppID := get("wechat_app_id", b.WeChatPay.AppID)
	wechatMchID := get("wechat_mch_id", b.WeChatPay.MchID)
	wechatV3Key := get("wechat_api_v3_key", b.WeChatPay.APIv3Key)
	wechatSerial := get("wechat_mch_serial_no", b.WeChatPay.MchSerialNo)
	wechatPriv := normalizePEM(get("wechat_mch_private_key", b.WeChatPay.MchPrivateKey))
	wechatPub := normalizePEM(get("wechat_platform_public_key", b.WeChatPay.PlatformPubKey))
	if wechatAppID != "" && wechatMchID != "" && wechatV3Key != "" && wechatSerial != "" && wechatPriv != "" && wechatPub != "" {
		wechatClient, e := upstream.NewWeChatPayClient(wechatAppID, wechatMchID, wechatV3Key, wechatSerial, wechatPriv, wechatPub)
		if e != nil {
			log.Printf("重建微信支付客户端失败，微信充值暂时不可用: %v", e)
		} else {
			s.wechat = wechatClient
		}
	}

	rt.mu.Lock()
	rt.snp = s
	rt.mu.Unlock()
	return nil
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

// normalizePEM 将 PEM 内容中的字面 \n 还原为换行（数据库中存储的单行 PEM 需还原为多行）。
func normalizePEM(s string) string {
	return strings.ReplaceAll(s, "\\n", "\n")
}
