package service

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

// EmailService 邮件服务（SMTP 发送验证码）
type EmailService struct {
	host         string
	port         int
	user         string
	password     string
	from         string
	verifyCodeSvc *VerificationCodeService
}

// NewEmailService 创建邮件服务
// 未配置 SMTP 时返回 nil（不启用邮箱验证码功能）
func NewEmailService(host string, port int, user, password, from string) *EmailService {
	if host == "" {
		return nil
	}
	if from == "" {
		from = user
	}
	if from == "" {
		from = "noreply@example.com"
	}
	return &EmailService{
		host:          host,
		port:          port,
		user:          user,
		password:      password,
		from:          from,
		verifyCodeSvc: NewVerificationCodeService(),
	}
}

// Enabled 邮件服务是否可用
func (s *EmailService) Enabled() bool {
	return s != nil && s.host != ""
}

// SendVerificationCode 发送邮箱验证码
func (s *EmailService) SendVerificationCode(email string) error {
	if s == nil {
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

	// 发送邮件
	subject := "【星楼KYC】邮箱验证码"
	body := fmt.Sprintf("您的验证码为：%s，5分钟内有效。若非本人操作，请忽略本邮件。", code)
	if err := s.Send(email, subject, body); err != nil {
		return err
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

// Send 发送普通邮件（支持 465 隐式 TLS 与 587/25 STARTTLS）
func (s *EmailService) Send(to, subject, body string) error {
	msg := buildMessage(s.from, to, subject, body)

	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	// 465 端口使用隐式 TLS
	if s.port == 465 {
		return s.sendTLS(addr, to, msg)
	}

	// 其他端口（587/25）使用标准 smtp（含 STARTTLS）
	auth := smtp.PlainAuth("", s.user, s.password, s.host)
	return smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg))
}

// sendTLS 通过隐式 TLS 发送
func (s *EmailService) sendTLS(addr, to, msg string) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: s.host})
	if err != nil {
		return fmt.Errorf("TLS 连接失败: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("SMTP 客户端创建失败: %w", err)
	}
	defer client.Close()

	auth := smtp.PlainAuth("", s.user, s.password, s.host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP 认证失败: %w", err)
	}
	if err := client.Mail(s.from); err != nil {
		return fmt.Errorf("SMTP Mail 失败: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("SMTP Rcpt 失败: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP Data 失败: %w", err)
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("SMTP 写入失败: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("SMTP 关闭失败: %w", err)
	}
	return client.Quit()
}

// buildMessage 构造邮件内容（含头与正文，UTF-8 编码）
func buildMessage(from, to, subject, body string) string {
	msg := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: =?UTF-8?B?" + base64Encode(subject) + "?=\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"Content-Transfer-Encoding: base64\r\n" +
		"\r\n" +
		base64Encode(body)
	return msg
}

// base64Encode 对字符串进行 Base64 编码（用于邮件头与正文）
func base64Encode(s string) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	data := []byte(s)
	var sb strings.Builder
	for i := 0; i < len(data); i += 3 {
		var b0, b1, b2 byte
		b0 = data[i]
		if i+1 < len(data) {
			b1 = data[i+1]
		}
		if i+2 < len(data) {
			b2 = data[i+2]
		}
		sb.WriteByte(chars[b0>>2])
		sb.WriteByte(chars[((b0&0x03)<<4)|(b1>>4)])
		if i+1 < len(data) {
			sb.WriteByte(chars[((b1&0x0F)<<2)|(b2>>6)])
		} else {
			sb.WriteByte('=')
		}
		if i+2 < len(data) {
			sb.WriteByte(chars[b2&0x3F])
		} else {
			sb.WriteByte('=')
		}
	}
	return sb.String()
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
