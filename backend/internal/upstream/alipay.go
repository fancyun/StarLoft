package upstream

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	alipayDefaultGateway = "https://openapi.alipay.com/gateway.do"
	alipayNotifyURL      = "https://www.starloft.cn/api/callback/alipay" // 异步通知地址（写死，与后端回调路由绑定）
	alipayReturnURL      = "https://console.starloft.cn/balance"            // 同步跳转地址（写死，指向控制台余额页）
)

// AlipayClient 支付宝开放平台支付客户端（电脑网站支付 alipay.trade.page.pay，RSA2 签名）
type AlipayClient struct {
	AppID      string
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

// NewAlipayClient 创建支付宝支付客户端；未配置 AppID 时返回 (nil, nil)
func NewAlipayClient(appID, privateKeyPEM, publicKeyPEM string) (*AlipayClient, error) {
	if appID == "" {
		return nil, nil
	}
	priv, err := parseRSAPrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("解析支付宝应用私钥失败: %w", err)
	}
	pub, err := parseRSAPublicKey(publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("解析支付宝公钥失败: %w", err)
	}
	return &AlipayClient{
		AppID:      appID,
		PrivateKey: priv,
		PublicKey:  pub,
	}, nil
}

// BuildPagePayURL 构建电脑网站支付跳转链接（浏览器 GET 访问即可拉起支付宝收银台）
func (c *AlipayClient) BuildPagePayURL(outTradeNo string, amount float64) (string, error) {
	bizContent, err := json.Marshal(map[string]string{
		"out_trade_no": outTradeNo,
		"product_code": "FAST_INSTANT_TRADE_PAY",
		"total_amount": fmt.Sprintf("%.2f", amount),
		"subject":      "账户余额充值",
	})
	if err != nil {
		return "", err
	}

	params := map[string]string{
		"app_id":      c.AppID,
		"method":      "alipay.trade.page.pay",
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  alipayNotifyURL,
		"return_url":  alipayReturnURL,
		"biz_content": string(bizContent),
	}

	sign, err := c.Sign(params)
	if err != nil {
		return "", err
	}

	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	values.Set("sign", sign)

	return alipayDefaultGateway + "?" + values.Encode(), nil
}

// Sign 对参数按 key 升序拼接 key=value（以 & 连接），使用应用私钥做 RSA2 签名
func (c *AlipayClient) Sign(params map[string]string) (string, error) {
	content := buildSignContent(params)
	hash := sha256.Sum256([]byte(content))
	sig, err := rsa.SignPKCS1v15(rand.Reader, c.PrivateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// VerifyNotify 验证异步通知签名（排除 sign 与 sign_type 字段）
func (c *AlipayClient) VerifyNotify(params map[string]string) bool {
	sign := params["sign"]
	if sign == "" {
		return false
	}
	sigBytes, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false
	}

	filtered := make(map[string]string)
	for k, v := range params {
		if k == "sign" || k == "sign_type" {
			continue
		}
		filtered[k] = v
	}
	hash := sha256.Sum256([]byte(buildSignContent(filtered)))
	return rsa.VerifyPKCS1v15(c.PublicKey, crypto.SHA256, hash[:], sigBytes) == nil
}

// buildSignContent 按 key 升序拼接 key=value，以 & 连接（剔除空值与 sign）
func buildSignContent(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if k == "sign" || v == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(params[k])
	}
	return sb.String()
}

// parseRSAPrivateKey 解析 PEM 私钥（支持 PKCS1 与 PKCS8 格式）
func parseRSAPrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("pem 解码失败")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, errors.New("不是 RSA 私钥")
	}
	return nil, errors.New("不支持的私钥格式")
}

// parseRSAPublicKey 解析 PEM 公钥（支持 PKCS1 与 PKIX 格式）
func parseRSAPublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("pem 解码失败")
	}
	if key, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PublicKey); ok {
			return rsaKey, nil
		}
		return nil, errors.New("不是 RSA 公钥")
	}
	return nil, errors.New("不支持的公钥格式")
}
