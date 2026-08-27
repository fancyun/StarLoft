package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"starloftrpa/internal/config"
	"starloftrpa/internal/model"
	"starloftrpa/internal/repository"
	"starloftrpa/internal/upstream"
	"starloftrpa/internal/utils"
)

// StartAuthResult StartAuth 返回结果
type StartAuthResult struct {
	Order   *model.AuthOrder
	AuthURL string // 上游认证 URL
	Token   string // 上游 token
	BizID   string // 上游 biz_id
}

// freeFailLimit 账户实名（Web）免费认证失败次数上限（账号终身累计，写死为3次）
const freeFailLimit = 3

// AuthService 认证服务
type AuthService struct {
	finAuthClient    upstream.FinAuthInterface
	orderRepo        *repository.AuthOrderRepository
	userRepo         *repository.UserRepository
	kycRecordRepo    *repository.KycRecordRepository
	resourcePackRepo *repository.ResourcePackRepository
	balanceService   *BalanceService
	configRepo       *repository.SystemConfigRepository
	config           *config.FinAuthConfig
}

// NewAuthService 创建认证服务
func NewAuthService(
	finAuthClient upstream.FinAuthInterface,
	orderRepo *repository.AuthOrderRepository,
	userRepo *repository.UserRepository,
	kycRecordRepo *repository.KycRecordRepository,
	resourcePackRepo *repository.ResourcePackRepository,
	balanceService *BalanceService,
	configRepo *repository.SystemConfigRepository,
	finAuthConfig *config.FinAuthConfig,
) *AuthService {
	return &AuthService{
		finAuthClient:    finAuthClient,
		orderRepo:        orderRepo,
		userRepo:         userRepo,
		kycRecordRepo:    kycRecordRepo,
		resourcePackRepo: resourcePackRepo,
		balanceService:   balanceService,
		configRepo:       configRepo,
		config:           finAuthConfig,
	}
}

// StartAuth 发起认证
// source: 1-账户实名（Web） 2-API调用
// free: true 表示账户实名免费路径（账号终身累计失败次数达到上限（写死3次）后转为计费）
func (s *AuthService) StartAuth(
	userID int64,
	name, idCard, bizNo, returnURL, notifyURL string,
	bizExtraData string,
	source int,
	free bool,
) (*StartAuthResult, error) {
	// 使用配置中的默认值（如果未传入）
	if returnURL == "" {
		returnURL = s.config.ReturnURL
	}
	if notifyURL == "" {
		notifyURL = s.config.NotifyURL
	}

	// 幂等去重：同一下游 biz_no 若已存在进行中订单，直接复用，避免重复创建任务/重复扣费
	if existing, err := s.orderRepo.GetOrderByBizNo(userID, bizNo); err == nil && existing != nil {
		if existing.Status == 0 || existing.Status == 1 {
			authURL := s.BuildAuthURL(existing.UpToken)
			return &StartAuthResult{
				Order:   existing,
				AuthURL: authURL,
				Token:   existing.UpToken,
				BizID:   existing.UpBizID,
			}, nil
		}
	}

	// 获取用户信息
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 确定本次认证是否计费：
	// - API 调用（source=2）始终计费；
	// - 账户实名（source=1）默认免费，但账号终身累计失败达到上限（写死3次）后转为计费（防止反复失败刷上游次数）
	charge := false
	if source == 2 {
		charge = true
	} else if source == 1 {
		failCount, err := s.orderRepo.CountUserFreeFailures(userID)
		if err != nil {
			return nil, fmt.Errorf("查询免费认证失败次数失败: %w", err)
		}
		if failCount >= freeFailLimit {
			charge = true
		}
	}

	// 确定扣费金额与扣费方式（有资源包时优先扣资源包，否则按平台单价从余额扣除）
	kycPrice := 0.0
	var userPack *model.UserResourcePack
	payType := 0 // 0-免费 1-余额 2-资源包
	if charge {
		kycPrice = s.getPlatformKycPrice()
		if kycPrice <= 0 {
			kycPrice = 1.00 // 兜底默认价格
		}

		userPack, err = s.resourcePackRepo.GetUserActivePack(userID)
		if err != nil {
			return nil, fmt.Errorf("查询用户资源包失败: %w", err)
		}
		if userPack != nil {
			payType = 2
		} else {
			payType = 1
			if user.Balance < kycPrice {
				return nil, fmt.Errorf("余额不足，需要%.2f元，当前余额%.2f元", kycPrice, user.Balance)
			}
		}
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

	// 创建订单（不保存姓名和身份证号，cost 记录平台KYC单价）
	order := &model.AuthOrder{
		PlatformBizNo: platformBizNo,
		BizNo:         bizNo,
		UserID:        userID,
		ReturnURL:     returnURL,
		NotifyURL:     notifyURL,
		BizExtraData:  bizExtraData,
		Cost:          kycPrice,
		Source:        source,
		PayType:       payType,
		Status:        0, // 待认证
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if payType == 2 {
		order.UserPackID = userPack.ID
	}

	err = s.orderRepo.CreateOrder(order)
	if err != nil {
		return nil, fmt.Errorf("create order failed: %w", err)
	}

	// 执行扣费（先扣资源包，再扣余额；扣费失败需回滚订单，避免悬挂订单）
	if charge {
		if err := s.deductOrderCharge(order, userPack, userID, kycPrice, kycRecord); err != nil {
			return nil, err
		}
	}

	// 传给上游 get_token 的地址：
	// - return_url 透传用户/下游地址：用户在上游完成认证后浏览器直接回到其 return_url，不再经平台 /kyc 中转页
	// - notify_url 始终为平台回调地址：认证结果由平台异步接收并落地（更新订单、扣费/退款、下发 API Secret、通知下游）
	upstreamReturnURL := returnURL
	upstreamNotifyURL := s.config.NotifyURL

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
		// 将结果标记为连接超时并退还预扣费用（资源包/余额）
		s.revertStartAuth(order, kycRecord, userID, "连接超时退款")
		return nil, fmt.Errorf("get token failed: %w", err)
	}

	// 更新订单上游信息
	err = s.orderRepo.UpdateOrderUpstreamInfo(order.ID, tokenResp.Token, tokenResp.BizID, tokenResp.RequestID)
	if err != nil {
		log.Printf("更新订单上游信息失败 [order_id=%d]: %v", order.ID, err)
		// 费用已扣但 token 未写入，退还预扣费用并将订单标记为超时退款，避免悬挂
		s.revertStartAuth(order, kycRecord, userID, "写入token失败退款")
		return nil, fmt.Errorf("update order upstream info failed: %w", err)
	}

	// 生成认证 URL（基于配置中的 base_url）
	authURL := s.BuildAuthURL(tokenResp.Token)

	return &StartAuthResult{
		Order:   order,
		AuthURL: authURL,
		Token:   tokenResp.Token,
		BizID:   tokenResp.BizID,
	}, nil
}

// GetPlatformKycPrice 获取平台KYC认证单价（系统配置 kyc_price）
func (s *AuthService) GetPlatformKycPrice() float64 {
	return s.getPlatformKycPrice()
}

// getPlatformKycPrice 获取平台KYC认证单价（系统配置 kyc_price）
func (s *AuthService) getPlatformKycPrice() float64 {
	if s.configRepo == nil {
		return 1.00
	}
	priceStr, err := s.configRepo.GetConfig("kyc_price")
	if err == nil && priceStr != "" {
		if price, err := strconv.ParseFloat(priceStr, 64); err == nil && price > 0 {
			return price
		}
	}
	return 1.00
}

// deductOrderCharge 执行订单扣费（先扣资源包，再扣余额）
// 资源包并发耗尽时回退到余额扣费；扣费失败将订单与实名记录标记为失败
func (s *AuthService) deductOrderCharge(
	order *model.AuthOrder,
	userPack *model.UserResourcePack,
	userID int64,
	kycPrice float64,
	kycRecord *model.KycRecord,
) error {
	if order.PayType == 2 && userPack != nil {
		ok, err := s.resourcePackRepo.DeductUserPackCount(userPack.ID, userID)
		if err != nil {
			log.Printf("扣减资源包失败 [order_id=%d, pack_id=%d]: %v", order.ID, userPack.ID, err)
			s.markOrderDeductFailed(order, kycRecord)
			return fmt.Errorf("扣减资源包失败: %w", err)
		}
		if ok {
			return nil
		}
		// 资源包并发耗尽/无剩余次数，回退到余额扣费
		log.Printf("资源包已耗尽，回退余额扣费 [order_id=%d, pack_id=%d]", order.ID, userPack.ID)
		if err := s.orderRepo.UpdateOrderPayType(order.ID, 1, 0); err != nil {
			log.Printf("回退余额扣费时更新订单扣费方式失败 [order_id=%d]: %v", order.ID, err)
		}
		order.PayType = 1
		order.UserPackID = 0
	}

	err := s.balanceService.DeductBalance(userID, kycPrice, order.ID, "KYC认证消费")
	if err != nil {
		log.Printf("扣除余额失败 [order_id=%d]: %v", order.ID, err)
		// 余额未扣（事务已回滚），仅将订单和实名记录标记为失败，避免产生悬挂订单
		s.markOrderDeductFailed(order, kycRecord)
		return fmt.Errorf("扣除余额失败: %w", err)
	}
	return nil
}

// markOrderDeductFailed 扣费失败时将订单与实名记录标记为失败，避免悬挂订单
func (s *AuthService) markOrderDeductFailed(order *model.AuthOrder, kycRecord *model.KycRecord) {
	_ = s.orderRepo.UpdateOrderResult(order.ID, "DEDUCT_FAILED", "扣费失败", 3)
	if kycRecord != nil {
		_ = s.kycRecordRepo.UpdateResult(kycRecord.ID, 3, "DEDUCT_FAILED", "扣费失败", "", nil)
	}
}

// refundOrderCharge 根据订单扣费方式退还费用（资源包退资源包次数，余额退余额）
func (s *AuthService) refundOrderCharge(order *model.AuthOrder, remark string) error {
	if order.PayType == 2 && order.UserPackID > 0 {
		return s.resourcePackRepo.RefundUserPackCount(order.UserPackID)
	}
	if order.PayType == 1 {
		return s.balanceService.RefundBalance(order.UserID, order.Cost, order.ID, remark)
	}
	return nil
}

// BuildAuthURL 根据上游 token 构造认证页面地址（用于跳转上游完成认证）
func (s *AuthService) BuildAuthURL(token string) string {
	if token == "" {
		return ""
	}
	return fmt.Sprintf("%s/finauth/lite/do?token=%s", s.config.BaseURL, token)
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

// ensureAPISecret 实名成功后生成并下发 API Secret（API Key 已于注册时自动生成）
// 仅对历史存量用户（注册时未生成 Key）一并补全 API Key
func (s *AuthService) ensureAPISecret(userID int64) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return
	}
	// 已有 Secret 则不重复生成
	if user.APISecret != "" {
		return
	}
	// 历史存量用户可能在注册时未生成 API Key（此前设计注册不生成 Key），实名成功后一并补全
	if user.APIKey == "" {
		apiKey := utils.GenerateRandomKey(32)
		if err := s.userRepo.UpdateUserAPIKey(userID, apiKey, ""); err != nil {
			log.Printf("实名成功后补发 API Key 失败 [user_id=%d]: %v", userID, err)
			return
		}
	}
	apiSecret := utils.GenerateRandomKey(32)
	if err := s.userRepo.UpdateUserAPISecret(userID, apiSecret); err != nil {
		log.Printf("实名成功后生成 API Secret 失败 [user_id=%d]: %v", userID, err)
		return
	}
	log.Printf("实名成功，已为用户生成 API Secret [user_id=%d]", userID)
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
		// 实名成功后自动生成下发 API Secret（开通 API 需先完成实名）
		s.ensureAPISecret(order.UserID)
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
		// 实名成功后自动生成下发 API Secret（开通 API 需先完成实名）
		s.ensureAPISecret(order.UserID)
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

// revertStartAuth 发起认证失败时退还预扣费用并将结果标记为连接超时（免费流程不涉及退费）
func (s *AuthService) revertStartAuth(order *model.AuthOrder, kycRecord *model.KycRecord, userID int64, remark string) {
	if order.Cost > 0 {
		if err := s.refundOrderCharge(order, remark); err != nil {
			log.Printf("发起认证失败退费失败 [order_id=%d]: %v", order.ID, err)
		}
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

// refundIfNotChargeable 不计费结果（6000/6100）退还预扣费用（资源包/余额）
func (s *AuthService) refundIfNotChargeable(order *model.AuthOrder, resultCode int) {
	if order.Cost <= 0 {
		return
	}
	if resultCode != 6000 && resultCode != 6100 {
		return
	}
	if order.IsRefunded == 1 {
		return
	}
	if err := s.refundOrderCharge(order, "认证不计费退款"); err != nil {
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
