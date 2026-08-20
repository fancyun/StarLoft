package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"starloftrpa/internal/repository"
)

var (
	ErrCallbackFailed = errors.New("callback failed")
)

type CallbackService struct {
	authOrderRepo *repository.AuthOrderRepository
}

func NewCallbackService(authOrderRepo *repository.AuthOrderRepository) *CallbackService {
	return &CallbackService{
		authOrderRepo: authOrderRepo,
	}
}

// NotifyDownstream 通知下游用户
func (s *CallbackService) NotifyDownstream(notifyURL string, data interface{}, apiSecret string) error {
	// 构建回调数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// 生成签名
	sign := s.generateSign(string(jsonData), apiSecret)

	// 构建请求体
	payload := map[string]interface{}{
		"data": data,
		"sign": sign,
	}
	payloadJSON, _ := json.Marshal(payload)

	// 发送 POST 请求
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", notifyURL, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return ErrCallbackFailed
	}

	// 解析响应
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}

	// 下游应该返回 {"code": 0} 表示成功接收
	if code, ok := result["code"].(float64); !ok || code != 0 {
		return ErrCallbackFailed
	}

	return nil
}

// RetryNotify 重试通知（指数退避）
func (s *CallbackService) RetryNotify(notifyURL string, data interface{}, apiSecret string, maxRetries int) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = s.NotifyDownstream(notifyURL, data, apiSecret)
		if err == nil {
			return nil
		}

		// 指数退避：1s, 2s, 4s
		waitTime := time.Duration(1<<uint(i)) * time.Second
		time.Sleep(waitTime)
	}

	return err
}

// generateSign 生成回调签名
func (s *CallbackService) generateSign(jsonData, apiSecret string) string {
	// 使用 HMAC-SHA256 签名算法
	// sign = hmac_sha256(api_secret, json_data)
	mac := hmac.New(sha256.New, []byte(apiSecret))
	mac.Write([]byte(jsonData))
	return hex.EncodeToString(mac.Sum(nil))
}
