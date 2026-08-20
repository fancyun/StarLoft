package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var (
	ErrEncryptionFailed = errors.New("encryption failed")
	ErrDecryptionFailed = errors.New("decryption failed")
	ErrInvalidKey       = errors.New("invalid encryption key")
)

// EncryptionManager 加密管理器（用于敏感信息如身份证号）
type EncryptionManager struct {
	key []byte
}

// NewEncryptionManager 创建加密管理器
// key 必须是 16, 24 或 32 字节（AES-128, AES-192, AES-256）
func NewEncryptionManager(key string) (*EncryptionManager, error) {
	keyBytes := []byte(key)
	if len(keyBytes) != 16 && len(keyBytes) != 24 && len(keyBytes) != 32 {
		return nil, ErrInvalidKey
	}
	return &EncryptionManager{key: keyBytes}, nil
}

// Encrypt 加密数据（使用AES-GCM）
func (m *EncryptionManager) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(m.key)
	if err != nil {
		return "", ErrEncryptionFailed
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", ErrEncryptionFailed
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", ErrEncryptionFailed
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密数据
func (m *EncryptionManager) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	block, err := aes.NewCipher(m.key)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrDecryptionFailed
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}

// EncryptIDCard 加密身份证号（带额外验证）
func (m *EncryptionManager) EncryptIDCard(idCard string) (string, error) {
	validator := NewInputValidator()
	if err := validator.ValidateIDCard(idCard); err != nil {
		return "", err
	}
	return m.Encrypt(idCard)
}

// DecryptIDCard 解密身份证号
func (m *EncryptionManager) DecryptIDCard(encryptedIDCard string) (string, error) {
	return m.Decrypt(encryptedIDCard)
}

// HashSensitiveData 对敏感数据进行哈希（用于查询，不可逆）
func (m *EncryptionManager) HashSensitiveData(data string) string {
	// 使用HMAC-SHA256作为哈希函数
	signMgr := NewSignatureManager()
	return signMgr.GenerateHMACSHA256(string(m.key), data)
}
