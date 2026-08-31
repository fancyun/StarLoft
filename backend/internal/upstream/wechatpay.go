package upstream

import (
	"bytes"
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"starloftrpa/internal/utils"
)

const (
	wechatAPIBase   = "https://api.mch.weixin.qq.com"
	wechatNotifyURL = "https://kyc.starloft.cn/api/v1/callback/wechat" // 异步通知地址（写死，与后端回调路由绑定）
)

// WeChatPayClient 微信支付（APIv3 Native支付）客户端
type WeChatPayClient struct {
	AppID             string
	MchID             string
	APIv3Key          string
	MchSerialNo       string
	MchPrivateKey     *rsa.PrivateKey
	PlatformPublicKey *rsa.PublicKey // 微信支付公钥（用于回调验签）
	httpClient        *http.Client
}

// NewWeChatPayClient 创建微信支付客户端；未配置（商户号或密钥未填写）时返回 (nil, nil)
func NewWeChatPayClient(appID, mchID, apiV3Key, mchSerialNo, mchPrivateKeyPEM, platformPublicKeyPEM string) (*WeChatPayClient, error) {
	if !isConfiguredValue(mchID) || !isConfiguredValue(mchPrivateKeyPEM) || !isConfiguredValue(platformPublicKeyPEM) {
		return nil, nil
	}
	priv, err := parseRSAPrivateKey(mchPrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("解析微信商户API私钥失败: %w", err)
	}
	pub, err := parseRSAPublicKey(platformPublicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("解析微信支付公钥失败: %w", err)
	}
	return &WeChatPayClient{
		AppID:             appID,
		MchID:             mchID,
		APIv3Key:          apiV3Key,
		MchSerialNo:       mchSerialNo,
		MchPrivateKey:     priv,
		PlatformPublicKey: pub,
		httpClient:        &http.Client{Timeout: 15 * time.Second},
	}, nil
}

// CreateNativeOrder 调用 Native 下单接口，返回 code_url（用于生成支付二维码）
func (c *WeChatPayClient) CreateNativeOrder(outTradeNo string, amount float64, description string) (string, error) {
	path := "/v3/pay/transactions/native"
	body := map[string]interface{}{
		"appid":        c.AppID,
		"mchid":        c.MchID,
		"description":  description,
		"out_trade_no": outTradeNo,
		"notify_url":   wechatNotifyURL,
		"amount": map[string]interface{}{
			"total":    int(math.Round(amount * 100)), // 金额单位：分
			"currency": "CNY",
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonce := utils.GenerateRandomKey(16)
	message := fmt.Sprintf("POST\n%s\n%s\n%s\n%s\n", path, timestamp, nonce, string(bodyBytes))
	signature, err := c.signMessage(message)
	if err != nil {
		return "", err
	}

	auth := fmt.Sprintf(
		`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",signature="%s",timestamp="%s",serial_no="%s"`,
		c.MchID, nonce, signature, timestamp, c.MchSerialNo,
	)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, wechatAPIBase+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", auth)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("微信Native下单失败: status=%d, body=%s", resp.StatusCode, string(respBytes))
	}

	var result struct {
		CodeURL string `json:"code_url"`
	}
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return "", err
	}
	if result.CodeURL == "" {
		return "", errors.New("微信Native下单返回 code_url 为空")
	}
	return result.CodeURL, nil
}

// VerifyNotify 验证微信支付回调签名（微信支付公钥验签）
// 签名串为 timestamp\nnonce\nbody\n，使用微信支付公钥做 SHA256withRSA 验证
func (c *WeChatPayClient) VerifyNotify(timestamp, nonce, body, signature string) bool {
	if c.PlatformPublicKey == nil || signature == "" {
		return false
	}
	message := fmt.Sprintf("%s\n%s\n%s\n", timestamp, nonce, body)
	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}
	hash := sha256.Sum256([]byte(message))
	return rsa.VerifyPKCS1v15(c.PlatformPublicKey, crypto.SHA256, hash[:], sigBytes) == nil
}

// DecryptNotify 解密微信支付回调资源（AES-256-GCM，密钥为 APIv3 密钥）
func (c *WeChatPayClient) DecryptNotify(ciphertext, nonce, associatedData string) ([]byte, error) {
	key := []byte(c.APIv3Key)
	if len(key) != 32 {
		return nil, errors.New("APIv3密钥长度必须为32字节")
	}
	cipherBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, []byte(nonce), cipherBytes, []byte(associatedData))
}

// signMessage 使用商户API私钥对签名串做 SHA256withRSA 签名
func (c *WeChatPayClient) signMessage(message string) (string, error) {
	hash := sha256.Sum256([]byte(message))
	sig, err := rsa.SignPKCS1v15(rand.Reader, c.MchPrivateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// WechatNotify 微信支付回调通知
type WechatNotify struct {
	ID           string               `json:"id"`
	EventType    string               `json:"event_type"`
	ResourceType string               `json:"resource_type"`
	Resource     WechatNotifyResource `json:"resource"`
	Summary      string               `json:"summary"`
}

// WechatNotifyResource 回调资源（密文）
type WechatNotifyResource struct {
	Algorithm      string `json:"algorithm"`
	Ciphertext     string `json:"ciphertext"`
	Nonce          string `json:"nonce"`
	AssociatedData string `json:"associated_data"`
}

// WechatTransaction 支付成功回调解密后的交易信息
type WechatTransaction struct {
	OutTradeNo    string `json:"out_trade_no"`
	TransactionID string `json:"transaction_id"`
	TradeState    string `json:"trade_state"`
	Amount        struct {
		Total int `json:"total"`
	} `json:"amount"`
	MchID string `json:"mchid"`
	AppID string `json:"appid"`
}
