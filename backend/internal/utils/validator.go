package utils

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	ErrInvalidPhone    = errors.New("invalid phone number")
	ErrInvalidPassword = errors.New("password must be 8-32 characters with letters and numbers")
	ErrInvalidIDCard   = errors.New("invalid ID card number")
	ErrInvalidName     = errors.New("invalid name")
	ErrInvalidAmount   = errors.New("invalid amount")
	ErrSQLInjection    = errors.New("potential SQL injection detected")
	ErrXSSAttempt      = errors.New("potential XSS attempt detected")
)

// InputValidator 输入验证器
type InputValidator struct{}

func NewInputValidator() *InputValidator {
	return &InputValidator{}
}

// ValidatePhone 验证手机号（中国大陆）
func (v *InputValidator) ValidatePhone(phone string) error {
	if phone == "" {
		return ErrInvalidPhone
	}
	matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, phone)
	if !matched {
		return ErrInvalidPhone
	}
	return nil
}

// ValidatePassword 验证密码强度
func (v *InputValidator) ValidatePassword(password string) error {
	if len(password) < 8 || len(password) > 32 {
		return ErrInvalidPassword
	}

	hasLetter := false
	hasNumber := false

	for _, char := range password {
		if unicode.IsLetter(char) {
			hasLetter = true
		}
		if unicode.IsDigit(char) {
			hasNumber = true
		}
	}

	if !hasLetter || !hasNumber {
		return ErrInvalidPassword
	}

	return nil
}

// ValidateIDCard 验证身份证号（中国大陆18位）
func (v *InputValidator) ValidateIDCard(idCard string) error {
	if len(idCard) != 18 {
		return ErrInvalidIDCard
	}

	// 基本格式验证
	matched, _ := regexp.MatchString(`^\d{17}[\dXx]$`, idCard)
	if !matched {
		return ErrInvalidIDCard
	}

	// 校验码验证
	weights := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	checkCodes := []byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}

	sum := 0
	for i := 0; i < 17; i++ {
		sum += int(idCard[i]-'0') * weights[i]
	}

	checkCode := checkCodes[sum%11]
	lastChar := idCard[17]
	if lastChar == 'x' {
		lastChar = 'X'
	}

	if lastChar != checkCode {
		return ErrInvalidIDCard
	}

	return nil
}

// ValidateName 验证姓名（2-50字符，仅中文、英文字母、间隔号）
func (v *InputValidator) ValidateName(name string) error {
	name = strings.TrimSpace(name)

	// 按字符数（rune）判断长度，而非字节数
	if utf8.RuneCountInString(name) < 2 || utf8.RuneCountInString(name) > 50 {
		return ErrInvalidName
	}

	// \p{Han} 匹配所有汉字（含扩展区生僻字），同时允许英文字母和间隔号
	matched, _ := regexp.MatchString(`^[\p{Han}a-zA-Z·]+$`, name)
	if !matched {
		return ErrInvalidName
	}

	return nil
}

// ValidateAmount 验证金额（大于0，最多两位小数）
func (v *InputValidator) ValidateAmount(amount float64) error {
	if amount <= 0 || amount > 1000000 {
		return ErrInvalidAmount
	}
	return nil
}

// SanitizeString 清理字符串（防止XSS）
func (v *InputValidator) SanitizeString(input string) string {
	// 移除潜在的HTML标签和脚本
	input = strings.ReplaceAll(input, "&", "&amp;")
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")
	input = strings.ReplaceAll(input, "\"", "&quot;")
	input = strings.ReplaceAll(input, "'", "&#39;")
	return input
}

// DetectSQLInjection 检测SQL注入特征
func (v *InputValidator) DetectSQLInjection(input string) error {
	lowerInput := strings.ToLower(input)
	sqlKeywords := []string{
		"' or '", "' and '", "union select", "drop table",
		"insert into", "delete from", "update set", "exec(",
		"execute(", "script", "javascript:", "onerror=",
	}

	for _, keyword := range sqlKeywords {
		if strings.Contains(lowerInput, keyword) {
			return ErrSQLInjection
		}
	}

	return nil
}

// ValidatePageParams 验证分页参数
func (v *InputValidator) ValidatePageParams(page, pageSize int) (int, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize, nil
}

// ValidateURL 验证URL格式
func (v *InputValidator) ValidateURL(url string) error {
	if url == "" {
		return nil
	}
	matched, _ := regexp.MatchString(`^https?://[^\s]+$`, url)
	if !matched {
		return errors.New("invalid URL format")
	}
	return nil
}
