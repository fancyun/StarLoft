package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"starloftrpa/internal/repository"
	"starloftrpa/internal/runtime"
	"starloftrpa/internal/service"
	"starloftrpa/internal/upstream"

	"github.com/gin-gonic/gin"
)

// CallbackHandler 回调处理器
type CallbackHandler struct {
	authService    *service.AuthService
	balanceService *service.BalanceService
	paymentRepo    *repository.PaymentOrderRepository
	rt             *runtime.Runtime
}

// NewCallbackHandler 创建回调处理器
func NewCallbackHandler(
	authService *service.AuthService,
	balanceService *service.BalanceService,
	paymentRepo *repository.PaymentOrderRepository,
	rt *runtime.Runtime,
) *CallbackHandler {
	return &CallbackHandler{
		authService:    authService,
		balanceService: balanceService,
		paymentRepo:    paymentRepo,
		rt:             rt,
	}
}

func (h *CallbackHandler) alipay() *upstream.AlipayClient    { return h.rt.Alipay() }
func (h *CallbackHandler) wechat() *upstream.WeChatPayClient { return h.rt.Wechat() }

// FinAuthCallback 处理 FinAuth 异步回调（notify_url）
// 文档: https://www.yljz.com/document/finauth-guide-docs/h5_will_plus_return_notify_url
// POST 请求，Content-Type: application/x-www-form-urlencoded
// 参数: data (JSON 字符串), sign (HMAC 签名)
func (h *CallbackHandler) FinAuthCallback(c *gin.Context) {
	// 读取原始请求体（脱敏记录：body 含签名与可能的人脸/证件数据，仅记录长度）
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	log.Printf("收到 FinAuth 回调: Content-Type=%s, BodyLen=%d", c.GetHeader("Content-Type"), len(bodyBytes))

	// 解析 form data（重置 body）
	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	c.Request.ParseForm()
	data := c.PostForm("data")
	sign := c.PostForm("sign")

	if data == "" {
		log.Printf("FinAuth 回调: 缺少 data 参数")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少 data 参数",
		})
		return
	}

	if sign == "" {
		log.Printf("FinAuth 回调: 缺少 sign 参数")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少 sign 参数",
		})
		return
	}

	// 处理回调（data 可能含人脸图片等敏感数据，仅记录截断片段）
	log.Printf("FinAuth 回调: DataLen=%d, SignLen=%d, DataHead=%s", len(data), len(sign), truncateStr(data, 120))
	err := h.authService.HandleUpstreamCallback(data, sign)
	if err != nil {
		log.Printf("FinAuth 回调处理失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "回调处理失败",
		})
		return
	}

	// 成功响应（必须返回 200）
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// AlipayCallback 处理支付宝异步通知（notify_url，RSA2 验签）
// 支付宝要求返回纯文本 "success" 表示通知处理成功
func (h *CallbackHandler) AlipayCallback(c *gin.Context) {
	if h.alipay() == nil {
		log.Printf("支付宝回调: 支付宝支付未配置")
		c.String(http.StatusOK, "failure")
		return
	}

	if err := c.Request.ParseForm(); err != nil {
		log.Printf("支付宝回调: 解析表单失败: %v", err)
		c.String(http.StatusOK, "failure")
		return
	}

	params := make(map[string]string)
	for k, v := range c.Request.PostForm {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	outTradeNo := params["out_trade_no"]
	tradeNo := params["trade_no"]
	tradeStatus := params["trade_status"]
	log.Printf("收到支付宝回调: out_trade_no=%s, trade_no=%s, trade_status=%s", outTradeNo, tradeNo, tradeStatus)

	// 验证签名
	if !h.alipay().VerifyNotify(params) {
		log.Printf("支付宝回调签名验证失败: out_trade_no=%s", outTradeNo)
		c.String(http.StatusOK, "failure")
		return
	}

	// 查询订单
	order, err := h.paymentRepo.GetOrderByPayOrderNo(outTradeNo)
	if err != nil {
		log.Printf("支付宝回调订单不存在: out_trade_no=%s, err=%v", outTradeNo, err)
		c.String(http.StatusOK, "failure")
		return
	}

	// 非支付成功状态，忽略（返回 success 停止支付宝重试）
	if tradeStatus != "TRADE_SUCCESS" && tradeStatus != "TRADE_FINISHED" {
		log.Printf("支付宝回调非成功状态: out_trade_no=%s, trade_status=%s", outTradeNo, tradeStatus)
		c.String(http.StatusOK, "success")
		return
	}

	// 幂等处理：仅当订单待支付时才入账
	changed, err := h.paymentRepo.MarkOrderPaidIfPending(order.ID, tradeNo)
	if err != nil {
		log.Printf("更新支付订单状态失败: order_id=%d, err=%v", order.ID, err)
		c.String(http.StatusOK, "failure")
		return
	}
	if !changed {
		// 已处理过，直接返回成功
		c.String(http.StatusOK, "success")
		return
	}

	// 增加用户余额
	if err := h.balanceService.RechargeBalance(order.UserID, order.Amount, order.ID); err != nil {
		log.Printf("支付宝回调入账失败: order_id=%d, err=%v", order.ID, err)
		c.String(http.StatusOK, "failure")
		return
	}

	log.Printf("支付宝回调处理成功: out_trade_no=%s, user_id=%d, amount=%.2f", outTradeNo, order.UserID, order.Amount)
	c.String(http.StatusOK, "success")
}

// WeChatCallback 处理微信支付异步通知（APIv3，公钥验签 + AES-GCM 解密）
// 微信要求返回 200 且响应体为 "SUCCESS" 表示处理成功，否则会重试
func (h *CallbackHandler) WeChatCallback(c *gin.Context) {
	if h.wechat() == nil {
		log.Printf("微信回调: 微信支付未配置")
		c.String(http.StatusInternalServerError, "FAIL")
		return
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("微信回调: 读取请求体失败: %v", err)
		c.String(http.StatusInternalServerError, "FAIL")
		return
	}
	bodyStr := string(bodyBytes)

	timestamp := c.GetHeader("Wechatpay-Timestamp")
	nonce := c.GetHeader("Wechatpay-Nonce")
	signature := c.GetHeader("Wechatpay-Signature")
	serial := c.GetHeader("Wechatpay-Serial")
	log.Printf("收到微信回调: BodyLen=%d, serial=%s", len(bodyBytes), serial)

	// 验证签名
	if !h.wechat().VerifyNotify(timestamp, nonce, bodyStr, signature) {
		log.Printf("微信回调签名验证失败: serial=%s", serial)
		c.String(http.StatusInternalServerError, "FAIL")
		return
	}

	// 解析通知并解密资源
	var notify upstream.WechatNotify
	if err := json.Unmarshal(bodyBytes, &notify); err != nil {
		log.Printf("微信回调报文解析失败: %v", err)
		c.String(http.StatusInternalServerError, "FAIL")
		return
	}
	plain, err := h.wechat().DecryptNotify(notify.Resource.Ciphertext, notify.Resource.Nonce, notify.Resource.AssociatedData)
	if err != nil {
		log.Printf("微信回调报文解密失败: %v", err)
		c.String(http.StatusInternalServerError, "FAIL")
		return
	}

	var tx upstream.WechatTransaction
	if err := json.Unmarshal(plain, &tx); err != nil {
		log.Printf("微信回调交易数据解析失败: %v", err)
		c.String(http.StatusInternalServerError, "FAIL")
		return
	}
	log.Printf("微信回调交易: out_trade_no=%s, transaction_id=%s, trade_state=%s", tx.OutTradeNo, tx.TransactionID, tx.TradeState)

	// 查询订单
	order, err := h.paymentRepo.GetOrderByPayOrderNo(tx.OutTradeNo)
	if err != nil {
		log.Printf("微信回调订单不存在: out_trade_no=%s, err=%v", tx.OutTradeNo, err)
		c.String(http.StatusInternalServerError, "FAIL")
		return
	}

	// 非支付成功状态，忽略（返回 SUCCESS 停止微信重试）
	if tx.TradeState != "SUCCESS" {
		log.Printf("微信回调非成功状态: out_trade_no=%s, trade_state=%s", tx.OutTradeNo, tx.TradeState)
		c.String(http.StatusOK, "SUCCESS")
		return
	}

	// 幂等处理：仅当订单待支付时才入账
	changed, err := h.paymentRepo.MarkOrderPaidIfPending(order.ID, tx.TransactionID)
	if err != nil {
		log.Printf("更新支付订单状态失败: order_id=%d, err=%v", order.ID, err)
		c.String(http.StatusInternalServerError, "FAIL")
		return
	}
	if !changed {
		// 已处理过，直接返回成功
		c.String(http.StatusOK, "SUCCESS")
		return
	}

	// 增加用户余额
	if err := h.balanceService.RechargeBalance(order.UserID, order.Amount, order.ID); err != nil {
		log.Printf("微信回调入账失败: order_id=%d, err=%v", order.ID, err)
		c.String(http.StatusInternalServerError, "FAIL")
		return
	}

	log.Printf("微信回调处理成功: out_trade_no=%s, user_id=%d, amount=%.2f", tx.OutTradeNo, order.UserID, order.Amount)
	c.String(http.StatusOK, "SUCCESS")
}

// truncateStr 截断长字符串用于日志输出，避免敏感/大体积数据刷屏
func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
