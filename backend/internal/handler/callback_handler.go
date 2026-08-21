package handler

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"

	"starloftrpa/internal/repository"
	"starloftrpa/internal/service"
	"starloftrpa/internal/upstream"

	"github.com/gin-gonic/gin"
)

// CallbackHandler 回调处理器
type CallbackHandler struct {
	authService     *service.AuthService
	balanceService  *service.BalanceService
	callbackService *service.CallbackService
	paymentRepo     *repository.PaymentOrderRepository
	unionPayClient  *upstream.UnionPayClient
}

// NewCallbackHandler 创建回调处理器
func NewCallbackHandler(
	authService *service.AuthService,
	balanceService *service.BalanceService,
	callbackService *service.CallbackService,
	paymentRepo *repository.PaymentOrderRepository,
	unionPayClient *upstream.UnionPayClient,
) *CallbackHandler {
	return &CallbackHandler{
		authService:     authService,
		balanceService:  balanceService,
		callbackService: callbackService,
		paymentRepo:     paymentRepo,
		unionPayClient:  unionPayClient,
	}
}

// FinAuthCallback 处理 FinAuth 异步回调（notify_url）
// 文档: https://www.yljz.com/document/finauth-guide-docs/h5_will_plus_return_notify_url
// POST 请求，Content-Type: application/x-www-form-urlencoded
// 参数: data (JSON 字符串), sign (HMAC 签名)
func (h *CallbackHandler) FinAuthCallback(c *gin.Context) {
	// 打印原始请求体用于调试
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	bodyStr := string(bodyBytes)
	log.Printf("收到 FinAuth 回调: Content-Type=%s, Body=%s", c.GetHeader("Content-Type"), bodyStr)

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

	// 处理回调
	log.Printf("FinAuth 回调: Data=%s, Sign=%s", data, sign)
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

// PaymentCallback 处理支付回调
// 银联天满支付异步通知（application/x-www-form-urlencoded）
func (h *CallbackHandler) PaymentCallback(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": "PARAM_ERROR"})
		return
	}

	params := make(map[string]string)
	for k, v := range c.Request.PostForm {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	merOrderID := params["merOrderId"]
	status := params["status"]
	seqID := params["seqId"]
	targetSys := params["targetSys"]

	log.Printf("收到支付回调: merOrderId=%s, status=%s, seqId=%s, targetSys=%s", merOrderID, status, seqID, targetSys)

	// 验证签名
	if h.unionPayClient == nil || !h.unionPayClient.VerifyNotify(params) {
		log.Printf("支付回调签名验证失败: merOrderId=%s", merOrderID)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": "SIGN_ERROR"})
		return
	}

	// 查询订单
	order, err := h.paymentRepo.GetOrderByPayOrderNo(merOrderID)
	if err != nil {
		log.Printf("支付回调订单不存在: merOrderId=%s, err=%v", merOrderID, err)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": "ORDER_NOT_EXIST"})
		return
	}

	// 非支付成功状态，忽略（等待后续通知）
	if !strings.EqualFold(status, "SUCCESS") {
		log.Printf("支付回调非成功状态: merOrderId=%s, status=%s", merOrderID, status)
		c.JSON(http.StatusOK, gin.H{"errCode": "SUCCESS"})
		return
	}

	// 幂等处理：仅当订单待支付时才入账
	changed, err := h.paymentRepo.MarkOrderPaidIfPending(order.ID, seqID)
	if err != nil {
		log.Printf("更新支付订单状态失败: order_id=%d, err=%v", order.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": "FAIL"})
		return
	}
	if !changed {
		// 已处理过，直接返回成功
		c.JSON(http.StatusOK, gin.H{"errCode": "SUCCESS"})
		return
	}

	// 增加用户余额
	if err := h.balanceService.RechargeBalance(order.UserID, order.Amount, order.ID); err != nil {
		log.Printf("支付回调入账失败: order_id=%d, err=%v", order.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": "FAIL"})
		return
	}

	log.Printf("支付回调处理成功: merOrderId=%s, user_id=%d, amount=%.2f", merOrderID, order.UserID, order.Amount)
	c.JSON(http.StatusOK, gin.H{"errCode": "SUCCESS"})
}
