package service

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ses "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ses/v20201002"
)

// EmailService 邮件服务（腾讯云 SES API 发送验证码）
type EmailService struct {
	client        *ses.Client
	from          string
	templateID    uint64
	verifyCodeSvc *VerificationCodeService
}

// NewEmailService 创建邮件服务
// 未配置 SES（发件人邮箱或模板ID为空）时返回 nil（不启用邮箱验证码功能）
// SecretId/SecretKey 复用腾讯云账号密钥，与短信服务一致
func NewEmailService(secretId, secretKey, from string, templateID uint64, region string) *EmailService {
	if from == "" || templateID == 0 {
		return nil
	}
	if region == "" {
		region = "ap-guangzhou"
	}

	credential := common.NewCredential(secretId, secretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "ses.tencentcloudapi.com"

	client, err := ses.NewClient(credential, region, cpf)
	if err != nil {
		return nil
	}

	return &EmailService{
		client:        client,
		from:          from,
		templateID:    templateID,
		verifyCodeSvc: NewVerificationCodeService(),
	}
}

// Enabled 邮件服务是否可用
func (s *EmailService) Enabled() bool {
	return s != nil && s.client != nil
}

// SendVerificationCode 发送邮箱验证码
func (s *EmailService) SendVerificationCode(email string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("邮件服务未配置")
	}

	// 检查发送频率限制
	canSend, remainSeconds, err := s.verifyCodeSvc.CheckEmailSendRateLimit(email)
	if err != nil {
		return fmt.Errorf("检查发送频率失败: %w", err)
	}
	if !canSend {
		return fmt.Errorf("发送过于频繁，请 %d 秒后再试", remainSeconds)
	}

	// 生成验证码
	code := s.verifyCodeSvc.GenerateCode()

	// 构造模板参数（模板中验证码变量名为 {{code}}）
	templateData, err := json.Marshal(map[string]string{"code": code})
	if err != nil {
		return fmt.Errorf("构造邮件模板参数失败: %w", err)
	}

	// 发送邮件（SES 模板发送）
	subject := "【星楼KYC】邮箱验证码"
	request := ses.NewSendEmailRequest()
	request.FromEmailAddress = common.StringPtr(s.from)
	request.Destination = common.StringPtrs([]string{email})
	request.Subject = common.StringPtr(subject)
	request.ReplyToAddresses = common.StringPtr(s.from)
	request.Template = &ses.Template{
		TemplateID:   common.Uint64Ptr(s.templateID),
		TemplateData: common.StringPtr(string(templateData)),
	}

	if _, err := s.client.SendEmail(request); err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	// 保存验证码到 Redis
	if err := s.verifyCodeSvc.SaveEmailCode(email, code); err != nil {
		return fmt.Errorf("保存验证码失败: %w", err)
	}

	// 设置发送频率限制
	if err := s.verifyCodeSvc.SetEmailSendRateLimit(email); err != nil {
		return fmt.Errorf("设置频率限制失败: %w", err)
	}

	log.Printf("邮箱验证码已发送: %s", maskEmail(email))
	return nil
}

// VerifyCode 验证邮箱验证码
func (s *EmailService) VerifyCode(email, code string) (bool, error) {
	if s == nil {
		return false, fmt.Errorf("邮件服务未配置")
	}
	return s.verifyCodeSvc.VerifyEmailCode(email, code)
}

// maskEmail 邮箱脱敏（仅记录用途）
func maskEmail(email string) string {
	idx := strings.Index(email, "@")
	if idx <= 1 {
		return "***" + email[idx:]
	}
	head := email[:idx]
	maskedHead := head[:1] + "***" + head[len(head)-1:]
	return maskedHead + email[idx:]
}
