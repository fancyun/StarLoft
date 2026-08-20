package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"starloftrpa/internal/config"
	"starloftrpa/internal/model"
	"starloftrpa/internal/repository"
	"starloftrpa/internal/upstream"
)

// StartAuthResult StartAuth 返回结果
type StartAuthResult struct {
	Order   *model.AuthOrder
	AuthURL string // 上游认证 URL
	Token   string // 上游 token
	BizID   string // 上游 biz_id
}

// AuthService 认证服务
type AuthService struct {
	finAuthClient  upstream.FinAuthInterface
	orderRepo      *repository.AuthOrderRepository
	userRepo       *repository.UserRepository
	kycRecordRepo  *repository.KycRecordRepository
	balanceService *BalanceService
	config         *config.FinAuthConfig
}

// NewAuthService 创建认证服务
func NewAuthService(
	finAuthClient upstream.FinAuthInterface,
	orderRepo *repository.AuthOrderRepository,
	userRepo *repository.UserRepository,
	kycRecordRepo *repository.KycRecordRepository,
	balanceService *BalanceService,
	finAuthConfig *config.FinAuthConfig,
) *AuthService {
	return &AuthService{
		finAuthClient:  finAuthClient,
		orderRepo:      orderRepo,
		userRepo:       userRepo,
		kycRecordRepo:  kycRecordRepo,
		balanceService: balanceService,
		config:         finAuthConfig,
	}
}

// StartAuth 发起认证
func (s *AuthService) StartAuth(
	userID int64,
	name, idCard, bizNo, returnURL, notifyURL string,
	bizExtraData string,
	usePlatformURLs bool,
) (*StartAuthResult, error) {
	// 使用配置中的默认值（如果未传入）
	if returnURL == "" {
		returnURL = s.config.ReturnURL
	}
	if notifyURL == "" {
		notifyURL = s.config.NotifyURL
	}

	// 检查用户余额并确定 KYC 单价（提前处理，确保订单记录扣费金额正确）
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 使用用户的KYC单价
	kycPrice := user.KYCPrice
	if kycPrice <= 0 {
		kycPrice = 1.00 // 默认1元
	}

	if user.Balance < kycPrice {
		return nil, fmt.Errorf("余额不足，需要%.2f元，当前余额%.2f元", kycPrice, user.Balance)
	}

	// 生成平台流水号
	platformBizNo := fmt.Sprintf("%s_%d_%d", bizNo, userID, time.Now().UnixNano())

	// 创建 KYC 认证记录（保存账户实名信息）
	kycRecord := &model.KycRecord{
		UserID: userID,
		Name:   name,
		IDCard: idCard,
		Status: 1, // 认证中
	}
	err = s.kycRecordRepo.Create(kycRecord)
	if err != nil {
		return nil, fmt.Errorf("create kyc record failed: %w", err)
	}

	// 创建订单（不保存姓名和身份证号，cost 记录用户实际 KYC 单价）
	order := &model.AuthOrder{
		PlatformBizNo: platformBizNo,
		BizNo:         bizNo,
		UserID:        userID,
		ReturnURL:     returnURL,
		NotifyURL:     notifyURL,
		BizExtraData:  bizExtraData,
		Cost:          kycPrice,
		Status:        0, // 待认证
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = s.orderRepo.CreateOrder(order)
	if err != nil {
		return nil, fmt.Errorf("create order failed: %w", err)
	}

	// 扣除余额（使用用户KYC单价）
	err = s.balanceService.DeductBalance(userID, kycPrice, order.ID, "KYC认证消费")
	if err != nil {
		log.Printf("扣除余额失败 [order_id=%d]: %v", order.ID, err)
		// 余额未扣（事务已回滚），仅将订单和实名记录标记为失败，避免产生悬挂订单
		_ = s.orderRepo.UpdateOrderResult(order.ID, "DEDUCT_FAILED", "扣费失败", 3)
		_ = s.kycRecordRepo.UpdateResult(kycRecord.ID, 3, "DEDUCT_FAILED", "扣费失败", "", nil)
		return nil, fmt.Errorf("扣除余额失败: %w", err)
	}

	// 传给上游 get_token 的地址：下游调用时使用平台地址（平台中转），否则透传
	upstreamReturnURL := returnURL
	upstreamNotifyURL := notifyURL
	if usePlatformURLs {
		upstreamReturnURL = s.buildPlatformReturnURL(platformBizNo)
		upstreamNotifyURL = s.config.NotifyURL
	}

	// 调用上游 get_token
	req := &upstream.GetTokenRequest{
		SignVersion:    upstream.SignVersionHMACSHA256,
		ReturnURL:      upstreamReturnURL,
		NotifyURL:      upstreamNotifyURL,
		BizNo:          platformBizNo,
		SceneID:        s.config.SceneID,
		ComparisonType: "1", // 人脸核身模式
		UUID:           fmt.Sprintf("%d", userID),
		BizExtraData:   bizExtraData,
		IDCardMode:     "0", // 不拍摄身份证，直接传入姓名和身份证号
		IDCardName:     name,
		IDCardNumber:   idCard,
	}

	tokenResp, err := s.finAuthClient.GetToken(req)
	if err != nil {
		log.Printf("获取 FinAuth Token 失败: %v", err)
		// 将结果标记为连接超时并退还预扣余额
		s.revertStartAuth(order, kycRecord, userID, "连接超时退款")
		return nil, fmt.Errorf("get token failed: %w", err)
	}

	// 更新订单上游信息
	err = s.orderRepo.UpdateOrderUpstreamInfo(order.ID, tokenResp.Token, tokenResp.BizID, tokenResp.RequestID)
	if err != nil {
		log.Printf("更新订单上游信息失败 [order_id=%d]: %v", order.ID, err)
		// 余额已扣但 token 未写入，退还预扣余额并将订单标记为超时退款，避免悬挂
		s.revertStartAuth(order, kycRecord, userID, "写入token失败退款")
		return nil, fmt.Errorf("update order upstream info failed: %w", err)
	}

	// 生成认证 URL（基于配置中的 base_url）
	authURL := fmt.Sprintf("%s/finauth/lite/do?token=%s", s.config.BaseURL, tokenResp.Token)

	return &StartAuthResult{
		Order:   order,
		AuthURL: authURL,
		Token:   tokenResp.Token,
		BizID:   tokenResp.BizID,
	}, nil
}

// buildPlatformReturnURL 构造平台中转 return 页面地址（/kyc?biz_no=平台流水号）
func (s *AuthService) buildPlatformReturnURL(platformBizNo string) string {
	base := s.config.ReturnURL
	if base == "" {
		return ""
	}
	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}
	return base + sep + "biz_no=" + url.QueryEscape(platformBizNo)
}

// GetAuthResult 查询认证结果
func (s *AuthService) GetAuthResult(userID int64, bizNo, platformBizNo string) (*model.AuthOrder, error) {
	var order *model.AuthOrder
	var err error

	if platformBizNo != "" {
		order, err = s.orderRepo.GetOrderByPlatformBizNo(platformBizNo)
	} else {
		order, err = s.orderRepo.GetOrderByBizNo(userID, bizNo)
	}
	if err != nil {
		return nil, err
	}

	// 如果订单还在处理中，尝试从上游查询最新结果
	if order.Status == 0 || order.Status == 1 {
		s.syncOrderResult(order)
		// 重新查询
		order, _ = s.orderRepo.GetOrderByPlatformBizNo(order.PlatformBizNo)
	}

	return order, nil
}

// GetPublicOrderResult 公开查询认证结果（/kyc 中转页按平台流水号查询，处理中时主动同步上游）
func (s *AuthService) GetPublicOrderResult(platformBizNo string) (*model.AuthOrder, error) {
	order, err := s.orderRepo.GetOrderByPlatformBizNo(platformBizNo)
	if err != nil {
		return nil, err
	}

	// 如果订单还在处理中，尝试从上游查询最新结果
	if order.Status == 0 || order.Status == 1 {
		s.syncOrderResult(order)
		order, _ = s.orderRepo.GetOrderByPlatformBizNo(platformBizNo)
	}

	return order, nil
}

// GetUserByID 查询用户
func (s *AuthService) GetUserByID(userID int64) (*model.PlatformUser, error) {
	return s.userRepo.GetUserByID(userID)
}

// GetLatestKycRecord 获取用户最新 KYC 记录
func (s *AuthService) GetLatestKycRecord(userID int64) (*model.KycRecord, error) {
	return s.kycRecordRepo.GetLatestByUserID(userID)
}

// GetLatestPendingOrder 获取用户最新进行中的认证订单（用于继续认证）
func (s *AuthService) GetLatestPendingOrder(userID int64) (*model.AuthOrder, error) {
	return s.orderRepo.GetLatestPendingOrder(userID)
}

// GetLatestOrder 获取用户最新订单（不限状态）
func (s *AuthService) GetLatestOrder(userID int64) (*model.AuthOrder, error) {
	return s.orderRepo.GetLatestOrderByUserID(userID)
}

// SyncOrderByToken 根据 token 同步上游结果（保留兼容；/kyc 页面现统一按 biz_no 查询）
func (s *AuthService) SyncOrderByToken(userID int64) (*model.AuthOrder, error) {
	order, err := s.orderRepo.GetLatestPendingOrder(userID)
	if err != nil {
		return nil, err
	}
	// 同步上游结果
	s.syncOrderResult(order)
	// 重新查询，获取最新状态
	order, err = s.orderRepo.GetOrderByPlatformBizNo(order.PlatformBizNo)
	if err != nil {
		return nil, err
	}
	return order, nil
}

// CancelKycRecord 取消用户的最新认证记录
func (s *AuthService) CancelKycRecord(userID int64) error {
	kycRecord, err := s.kycRecordRepo.GetLatestByUserID(userID)
	if err != nil {
		return err
	}
	if kycRecord.Status != 1 {
		return fmt.Errorf("no pending kyc record")
	}
	return s.kycRecordRepo.Cancel(kycRecord.ID)
}

// ReplaceKycRecord 更换实名：将最新已实名记录标记为"已更换"，清除用户实名状态
func (s *AuthService) ReplaceKycRecord(userID int64) error {
	// 1. 将最新已实名的 kyc_record 标记为"已更换"
	err := s.kycRecordRepo.Replace(userID)
	if err != nil {
		return fmt.Errorf("replace kyc record failed: %w", err)
	}

	// 2. 清除 platform_user 的实名信息
	err = s.userRepo.ClearUserKYCInfo(userID)
	if err != nil {
		return fmt.Errorf("clear user kyc info failed: %w", err)
	}

	return nil
}

// GetUserAuthRecords 查询用户认证记录
func (s *AuthService) GetUserAuthRecords(userID int64, page, pageSize int) ([]*model.AuthOrder, int64, error) {
	return s.orderRepo.GetUserOrders(userID, page, pageSize)
}

// GetUserAuthCallStats 统计用户近 N 天的认证调用次数（按天，缺失日期补零）
func (s *AuthService) GetUserAuthCallStats(userID int64, days int) ([]string, []int64, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -(days - 1))

	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")

	stats, err := s.orderRepo.GetUserDailyAuthCount(userID, startStr, endStr)
	if err != nil {
		return nil, nil, err
	}

	dates := make([]string, 0, days)
	counts := make([]int64, 0, days)
	for i := 0; i < days; i++ {
		d := startDate.AddDate(0, 0, i).Format("2006-01-02")
		dates = append(dates, d)
		counts = append(counts, stats[d])
	}

	return dates, counts, nil
}

// isInProgressMessage 判断上游返回的 result_message 是否表示「认证尚未开始/进行中」。
// 该状态下认证流程尚未完结，不应视为失败。
func isInProgressMessage(msg string) bool {
	m := strings.ToUpper(strings.TrimSpace(msg))
	switch m {
	case "NOT_STARTED", "PROCESSING", "NOT_START", "IN_PROGRESS", "PENDING":
		return true
	}
	return false
}

// syncOrderResult 同步订单结果（从上游查询）
func (s *AuthService) syncOrderResult(order *model.AuthOrder) {
	if order.UpBizID == "" {
		return
	}

	req := &upstream.GetResultRequest{
		BizID:       order.UpBizID,
		SignVersion: upstream.SignVersionHMACSHA256,
	}

	result, err := s.finAuthClient.GetResult(req)
	if err != nil {
		log.Printf("同步订单结果失败 [biz_id=%s]: %v", order.UpBizID, err)
		return
	}

	// 未开始/进行中：认证尚未完结，保持「认证中」，不结束流程
	if isInProgressMessage(result.ResultMessage) {
		log.Printf("认证尚未开始或进行中，保持认证中状态 [biz_id=%s, result_message=%s]", order.UpBizID, result.ResultMessage)
		return
	}

	// 判断结果状态
	status := order.Status
	switch result.ResultCode {
	case 1000:
		status = 2 // 认证成功
	case 2000, 3000, 4000:
		status = 3 // 认证失败（计费）
	case 6000, 6100:
		status = 3 // 认证失败（不计费）
	default:
		if result.ResultCode != 0 {
			status = 3
		}
	}

	resultCode := fmt.Sprintf("%d", result.ResultCode)
	err = s.orderRepo.UpdateOrderResult(order.ID, resultCode, result.ResultMessage, status)
	if err != nil {
		log.Printf("更新订单结果失败 [order_id=%d]: %v", order.ID, err)
		return
	}

	// 不计费结果（6000/6100）退还预扣余额
	s.refundIfNotChargeable(order, int(result.ResultCode))

	// 更新 kyc_record 状态
	if status == 2 {
		kycRecord, err := s.kycRecordRepo.GetLatestByUserID(order.UserID)
		if err == nil && kycRecord != nil {
			// ✅ 使用加密后的身份证号更新用户信息
			err = s.userRepo.UpdateUserKYCInfo(order.UserID, kycRecord.Name, kycRecord.IDCard)
			if err != nil {
				log.Printf("更新用户实名信息失败 [user_id=%d]: %v", order.UserID, err)
			}
			now := time.Now()
			_ = s.kycRecordRepo.UpdateResult(kycRecord.ID, 2, resultCode, result.ResultMessage, "", &now)
		}
	} else if status == 3 {
		kycRecord, err := s.kycRecordRepo.GetLatestByUserID(order.UserID)
		if err == nil && kycRecord != nil {
			_ = s.kycRecordRepo.UpdateResult(kycRecord.ID, 3, resultCode, result.ResultMessage, "", nil)
		}
	}

	// 通知下游
	order.Status = status
	order.ResultCode = resultCode
	order.ResultMessage = result.ResultMessage
	s.NotifyDownstream(order)
}

// HandleUpstreamCallback 处理上游异步回调
// data: JSON 字符串，sign: HMAC 签名
func (s *AuthService) HandleUpstreamCallback(data, sign string) error {
	// 1. 验证签名
	if !s.finAuthClient.VerifySign(data, sign) {
		log.Printf("回调签名验证失败: data=%s, sign=%s", data, sign)
		return errors.New("signature verification failed")
	}

	// 2. 解析回调数据
	var notifyData upstream.NotifyData
	if err := json.Unmarshal([]byte(data), &notifyData); err != nil {
		log.Printf("解析回调数据失败: %v, data=%s", err, data)
		return fmt.Errorf("parse notify data failed: %w", err)
	}

	// 3. 根据 biz_id 查找订单
	order, err := s.orderRepo.GetOrderByUpBizID(notifyData.BizInfo.BizID)
	if err != nil {
		log.Printf("查找订单失败 [biz_id=%s]: %v", notifyData.BizInfo.BizID, err)
		return fmt.Errorf("order not found: %w", err)
	}

	// 未开始/进行中：认证尚未完结，忽略本次回调，保持「认证中」
	if isInProgressMessage(notifyData.ResultMessage) {
		log.Printf("回调表示认证尚未开始或进行中，忽略 [biz_id=%s, result_message=%s]", notifyData.BizInfo.BizID, notifyData.ResultMessage)
		return nil
	}

	// 4. 判断认证结果
	status := order.Status
	switch notifyData.ResultCode {
	case 1000:
		status = 2 // 认证成功
	case 2000, 3000, 4000:
		status = 3 // 认证失败（计费）
	case 6000, 6100:
		status = 3 // 认证失败（不计费）
	default:
		status = 3
	}

	resultCode := fmt.Sprintf("%d", notifyData.ResultCode)
	err = s.orderRepo.UpdateOrderResult(order.ID, resultCode, notifyData.ResultMessage, status)
	if err != nil {
		log.Printf("更新订单结果失败 [order_id=%d]: %v", order.ID, err)
		return fmt.Errorf("update order result failed: %w", err)
	}

	// 不计费结果（6000/6100）退还预扣余额
	s.refundIfNotChargeable(order, int(notifyData.ResultCode))

	// 5. 更新 kyc_record 状态
	if status == 2 {
		kycRecord, err := s.kycRecordRepo.GetLatestByUserID(order.UserID)
		if err == nil && kycRecord != nil {
			err = s.userRepo.UpdateUserKYCInfo(order.UserID, kycRecord.Name, kycRecord.IDCard)
			if err != nil {
				log.Printf("更新用户实名信息失败 [user_id=%d]: %v", order.UserID, err)
			}
			now := time.Now()
			_ = s.kycRecordRepo.UpdateResult(kycRecord.ID, 2, resultCode, notifyData.ResultMessage, "", &now)
		}
	} else {
		kycRecord, err := s.kycRecordRepo.GetLatestByUserID(order.UserID)
		if err == nil && kycRecord != nil {
			_ = s.kycRecordRepo.UpdateResult(kycRecord.ID, 3, resultCode, notifyData.ResultMessage, "", nil)
		}
	}

	// 通知下游
	order.Status = status
	order.ResultCode = resultCode
	order.ResultMessage = notifyData.ResultMessage
	s.NotifyDownstream(order)

	log.Printf("回调处理成功 [biz_id=%s, result_code=%d]", notifyData.BizInfo.BizID, notifyData.ResultCode)
	return nil
}

// NotifyDownstream 通知下游：将认证结果 POST 到下游的 notify_url
func (s *AuthService) NotifyDownstream(order *model.AuthOrder) {
	if order.NotifyURL == "" {
		return
	}

	payload := map[string]interface{}{
		"biz_no":          order.BizNo,
		"platform_biz_no": order.PlatformBizNo,
		"status":          order.Status,
		"result_code":     order.ResultCode,
		"result_message":  order.ResultMessage,
		"cost":            order.Cost,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("序列化通知数据失败 [order_id=%d]: %v", order.ID, err)
		return
	}

	resp, err := http.Post(order.NotifyURL, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		log.Printf("通知下游失败 [order_id=%d, url=%s]: %v", order.ID, order.NotifyURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("通知下游成功 [order_id=%d, url=%s, status=%d]", order.ID, order.NotifyURL, resp.StatusCode)
	} else {
		log.Printf("通知下游返回异常 [order_id=%d, url=%s, status=%d]", order.ID, order.NotifyURL, resp.StatusCode)
	}
}

// revertStartAuth 发起认证失败时退还预扣余额并将结果标记为连接超时
func (s *AuthService) revertStartAuth(order *model.AuthOrder, kycRecord *model.KycRecord, userID int64, remark string) {
	if err := s.balanceService.RefundBalance(userID, order.Cost, order.ID, remark); err != nil {
		log.Printf("发起认证失败退费失败 [order_id=%d]: %v", order.ID, err)
	}

	// 不删除订单与实名记录，将结果标记为连接超时（status=5 超时/已退款）
	if err := s.orderRepo.UpdateOrderResult(order.ID, "TIMEOUT", "连接超时", 5); err != nil {
		log.Printf("更新订单连接超时失败 [order_id=%d]: %v", order.ID, err)
	}
	if err := s.orderRepo.UpdateOrderRefundFlag(order.ID); err != nil {
		log.Printf("更新订单退款标记失败 [order_id=%d]: %v", order.ID, err)
	}

	// 更新实名记录状态为失败（连接超时）
	if err := s.kycRecordRepo.UpdateResult(kycRecord.ID, 3, "TIMEOUT", "连接超时", "", nil); err != nil {
		log.Printf("更新实名记录连接超时失败 [record_id=%d]: %v", kycRecord.ID, err)
	}
}

// refundIfNotChargeable 不计费结果（6000/6100）退还预扣余额
func (s *AuthService) refundIfNotChargeable(order *model.AuthOrder, resultCode int) {
	if resultCode != 6000 && resultCode != 6100 {
		return
	}
	if order.IsRefunded == 1 {
		return
	}
	if err := s.balanceService.RefundBalance(order.UserID, order.Cost, order.ID, "认证不计费退款"); err != nil {
		log.Printf("不计费退款失败 [order_id=%d]: %v", order.ID, err)
		return
	}
	if err := s.orderRepo.UpdateOrderRefundFlag(order.ID); err != nil {
		log.Printf("更新退款标记失败 [order_id=%d]: %v", order.ID, err)
		return
	}
	order.IsRefunded = 1
}

// SyncPendingOrders 同步所有处理中订单的上游结果，并根据上游返回的结果决定是否退款
func (s *AuthService) SyncPendingOrders() error {
	orders, err := s.orderRepo.GetPendingOrders()
	if err != nil {
		return err
	}

	for _, order := range orders {
		// syncOrderResult 会根据上游结果码更新订单状态，
		// 并在结果为不计费(6000/6100)时自动退还预扣余额
		s.syncOrderResult(order)
	}
	return nil
}
