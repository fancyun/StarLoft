package service

import (
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

// SMSService 短信服务
type SMSService struct {
	client        *sms.Client
	sdkAppId      string
	signName      string
	templateId    string
	verifyCodeSvc *VerificationCodeService
}

// NewSMSService 创建短信服务
func NewSMSService(secretId, secretKey, sdkAppId, signName, templateId string) (*SMSService, error) {
	credential := common.NewCredential(secretId, secretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "sms.tencentcloudapi.com"

	client, err := sms.NewClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return &SMSService{
		client:        client,
		sdkAppId:      sdkAppId,
		signName:      signName,
		templateId:    templateId,
		verifyCodeSvc: NewVerificationCodeService(),
	}, nil
}

// SendVerificationCode 发送验证码
func (s *SMSService) SendVerificationCode(phone string) error {
	// 检查发送频率限制
	canSend, remainSeconds, err := s.verifyCodeSvc.CheckSendRateLimit(phone)
	if err != nil {
		return fmt.Errorf("检查发送频率失败: %w", err)
	}
	if !canSend {
		return fmt.Errorf("发送过于频繁，请 %d 秒后再试", remainSeconds)
	}

	// 生成验证码
	code := s.verifyCodeSvc.GenerateCode()

	// 发送短信
	request := sms.NewSendSmsRequest()
	request.SmsSdkAppId = common.StringPtr(s.sdkAppId)
	request.SignName = common.StringPtr(s.signName)
	request.TemplateId = common.StringPtr(s.templateId)
	request.TemplateParamSet = common.StringPtrs([]string{code, "5"})
	request.PhoneNumberSet = common.StringPtrs([]string{"+86" + phone})

	response, err := s.client.SendSms(request)
	if err != nil {
		return fmt.Errorf("发送短信失败: %w", err)
	}

	// 检查发送结果
	if response.Response.SendStatusSet == nil || len(response.Response.SendStatusSet) == 0 {
		return fmt.Errorf("短信发送失败：无响应")
	}

	status := response.Response.SendStatusSet[0]
	if *status.Code != "Ok" {
		return fmt.Errorf("短信发送失败：%s", *status.Message)
	}

	// 保存验证码到 Redis
	if err := s.verifyCodeSvc.SaveCode(phone, code); err != nil {
		return fmt.Errorf("保存验证码失败: %w", err)
	}

	// 设置发送频率限制
	if err := s.verifyCodeSvc.SetSendRateLimit(phone); err != nil {
		return fmt.Errorf("设置频率限制失败: %w", err)
	}

	return nil
}

// VerifyCode 验证验证码
func (s *SMSService) VerifyCode(phone, code string) (bool, error) {
	return s.verifyCodeSvc.VerifyCode(phone, code)
}

// GenerateVerificationCode 生成验证码（供测试使用）
func (s *SMSService) GenerateVerificationCode() string {
	return s.verifyCodeSvc.GenerateCode()
}
