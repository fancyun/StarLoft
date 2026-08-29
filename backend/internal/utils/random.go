package utils

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

// GenerateRandomKey 生成指定字节长度的随机十六进制密钥（使用密码学安全随机源）
// 例如 length=32 时生成 64 位十六进制字符（256 位熵）
func GenerateRandomKey(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GenerateRandomDigits 生成指定长度的随机数字字符串（使用密码学安全随机源）
// 例如 length=20 时生成 20 位随机数字
func GenerateRandomDigits(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	var sb strings.Builder
	sb.Grow(length)
	for _, by := range b {
		sb.WriteByte('0' + by%10)
	}
	return sb.String()
}
