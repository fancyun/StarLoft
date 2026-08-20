package upstream

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// 支付渠道常量
const (
	ChannelAlipay   = "alipay"   // 支付宝
	ChannelWechat   = "wechat"   // 微信支付
	ChannelUnionPay = "unionpay" // 云闪付（银联在线支付）
)

// UnionPayClient 银联商务天满支付客户端
type UnionPayClient struct {
	MerchantID  string
	TerminalID  string
	AccessToken string
	SecretKey   string
	BaseURL     string
	NotifyURL   string
}

// NewUnionPayClient 创建银联商务支付客户端
func NewUnionPayClient(merchantID, terminalID, accessToken, secretKey, baseURL, notifyURL string) *UnionPayClient {
	return &UnionPayClient{
		MerchantID:  merchantID,
		TerminalID:  terminalID,
		AccessToken: accessToken,
		SecretKey:   secretKey,
		BaseURL:     baseURL,
		NotifyURL:   notifyURL,
	}
}

// BarcodePayRequest 条码支付请求参数
type BarcodePayRequest struct {
	MerchantID     string `json:"merchantId"`     // 商户号
	TerminalID     string `json:"terminalId"`     // 终端号
	MerOrderID     string `json:"merOrderId"`     // 商户订单号
	TotalAmount    string `json:"totalAmount"`    // 订单金额（分）
	OrderDesc      string `json:"orderDesc"`      // 订单描述
	AuthCode       string `json:"authCode"`       // 付款码
	NotifyURL      string `json:"notifyUrl"`      // 异步通知地址
	OrderTimestamp string `json:"orderTimestamp"` // 订单时间戳
	Sign           string `json:"sign"`           // 签名
}

// BarcodePayResponse 条码支付响应
type BarcodePayResponse struct {
	ErrCode       string `json:"errCode"`       // 错误码
	ErrMsg        string `json:"errMsg"`        // 错误信息
	Status        string `json:"status"`        // 订单状态：SUCCESS-成功 FAIL-失败 PAYING-支付中
	MerOrderID    string `json:"merOrderId"`    // 商户订单号
	SeqID         string `json:"seqId"`         // 银联订单号
	TargetSys     string `json:"targetSys"`     // 目标系统：ALIPAY-支付宝 WECHAT-微信
	TargetOrderID string `json:"targetOrderId"` // 第三方订单号
	TotalAmount   string `json:"totalAmount"`   // 订单金额
}

// QueryOrderRequest 查询订单请求参数
type QueryOrderRequest struct {
	MerchantID string `json:"merchantId"` // 商户号
	TerminalID string `json:"terminalId"` // 终端号
	MerOrderID string `json:"merOrderId"` // 商户订单号
	Sign       string `json:"sign"`       // 签名
}

// QueryOrderResponse 查询订单响应
type QueryOrderResponse struct {
	ErrCode       string `json:"errCode"`       // 错误码
	ErrMsg        string `json:"errMsg"`        // 错误信息
	Status        string `json:"status"`        // 订单状态
	MerOrderID    string `json:"merOrderId"`    // 商户订单号
	SeqID         string `json:"seqId"`         // 银联订单号
	TargetSys     string `json:"targetSys"`     // 目标系统
	TargetOrderID string `json:"targetOrderId"` // 第三方订单号
	TotalAmount   string `json:"totalAmount"`   // 订单金额
	OrderDesc     string `json:"orderDesc"`     // 订单描述
}

// RefundRequest 退款请求参数
type RefundRequest struct {
	MerchantID    string `json:"merchantId"`    // 商户号
	TerminalID    string `json:"terminalId"`    // 终端号
	RefundOrderID string `json:"refundOrderId"` // 退款订单号
	MerOrderID    string `json:"merOrderId"`    // 原商户订单号
	RefundAmount  string `json:"refundAmount"`  // 退款金额（分）
	NotifyURL     string `json:"notifyUrl"`     // 异步通知地址
	Sign          string `json:"sign"`          // 签名
}

// RefundResponse 退款响应
type RefundResponse struct {
	ErrCode       string `json:"errCode"`       // 错误码
	ErrMsg        string `json:"errMsg"`        // 错误信息
	Status        string `json:"status"`        // 退款状态
	RefundOrderID string `json:"refundOrderId"` // 退款订单号
	MerOrderID    string `json:"merOrderId"`    // 原商户订单号
	SeqID         string `json:"seqId"`         // 银联订单号
	RefundAmount  string `json:"refundAmount"`  // 退款金额
}

// GenerateSign 生成签名
func (c *UnionPayClient) GenerateSign(params map[string]string) string {
	// 1. 按照参数名ASCII码从小到大排序
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 2. 拼接参数：key1=value1&key2=value2
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString("&")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}

	// 3. 拼接密钥：string&key=secretKey
	builder.WriteString("&key=")
	builder.WriteString(c.SecretKey)

	// 4. MD5加密并转大写
	hash := md5.Sum([]byte(builder.String()))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// BarcodePay 条码支付（扫码枪扫用户付款码）
func (c *UnionPayClient) BarcodePay(orderID string, amount int64, desc string, authCode string) (*BarcodePayResponse, error) {
	params := map[string]string{
		"merchantId":     c.MerchantID,
		"terminalId":     c.TerminalID,
		"merOrderId":     orderID,
		"totalAmount":    fmt.Sprintf("%d", amount), // 金额单位：分
		"orderDesc":      desc,
		"authCode":       authCode,
		"notifyUrl":      c.NotifyURL,
		"orderTimestamp": time.Now().Format("20060102150405"),
	}

	// 生成签名
	params["sign"] = c.GenerateSign(params)

	// 构造请求
	reqURL := c.BaseURL + "/v1/netpay/bills/barcodepay"
	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	req, err := http.NewRequest("POST", reqURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "OPEN-ACCESS-TOKEN AccessToken="+c.AccessToken)

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result BarcodePayResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// QueryOrder 查询订单
func (c *UnionPayClient) QueryOrder(orderID string) (*QueryOrderResponse, error) {
	params := map[string]string{
		"merchantId": c.MerchantID,
		"terminalId": c.TerminalID,
		"merOrderId": orderID,
	}

	// 生成签名
	params["sign"] = c.GenerateSign(params)

	// 构造请求
	reqURL := c.BaseURL + "/v1/netpay/bills/query"
	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	req, err := http.NewRequest("POST", reqURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "OPEN-ACCESS-TOKEN AccessToken="+c.AccessToken)

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result QueryOrderResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Refund 退款
func (c *UnionPayClient) Refund(refundOrderID, originalOrderID string, amount int64) (*RefundResponse, error) {
	params := map[string]string{
		"merchantId":    c.MerchantID,
		"terminalId":    c.TerminalID,
		"refundOrderId": refundOrderID,
		"merOrderId":    originalOrderID,
		"refundAmount":  fmt.Sprintf("%d", amount), // 金额单位：分
		"notifyUrl":     c.NotifyURL,
	}

	// 生成签名
	params["sign"] = c.GenerateSign(params)

	// 构造请求
	reqURL := c.BaseURL + "/v1/netpay/bills/refund"
	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	req, err := http.NewRequest("POST", reqURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "OPEN-ACCESS-TOKEN AccessToken="+c.AccessToken)

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result RefundResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// VerifyNotify 验证异步通知签名
func (c *UnionPayClient) VerifyNotify(params map[string]string) bool {
	receivedSign := params["sign"]
	if receivedSign == "" {
		return false
	}

	calculatedSign := c.GenerateSign(params)
	return receivedSign == calculatedSign
}

// NormalizeChannel 将天满返回的 targetSys 标准化为平台内部使用的渠道标识
// （alipay / wechat / unionpay），用于订单入库和展示。
func NormalizeChannel(targetSys string) string {
	switch strings.ToUpper(targetSys) {
	case "ALIPAY":
		return ChannelAlipay
	case "WECHAT", "WEIXIN":
		return ChannelWechat
	case "UNIONPAY", "QUICKPASS", "UPSDK":
		return ChannelUnionPay
	default:
		return strings.ToLower(targetSys)
	}
}
