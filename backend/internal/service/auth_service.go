package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
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
// 账户实名（source=1）不产生认证订单，走 Record；API 调用（source=2）走 Order
type StartAuthResult struct {
	Order   *model.AuthOrder
	Record  *model.KycPersonal
	AuthURL string // 上游认证 URL
	Token   string // 上游 token
	BizID   string // 上游 biz_id
}

// freeFailLimit 账户实名（Web）免费认证失败次数上限（账号终身累计，写死为3次，仅用于前端提示展示）
const freeFailLimit = 3

// 上游（FinAuth）回调与跳转地址（写死，与后端回调路由/前端页面绑定）
const (
	finAuthNotifyURL = "https://www.starloft.cn/api/callback/finauth" // 上游异步通知地址：认证结果由平台接收并落地
	finAuthReturnURL = "/user/kyc"                                    // 上游同步跳转默认地址：未传入 return_url 时兜底，认证完成后返回账户实名页
)

// AuthService 认证服务
type AuthService struct {
	// finAuthClient / finAuthCfg 为取器闭包，从运行时快照获取当前生效的 FinAuth 客户端与配置，
	// 业务配置在后台修改后无需重启即对后续认证生效。
	finAuthClient    func() upstream.FinAuthInterface
	finAuthCfg       func() config.FinAuthConfig
	orderRepo        *repository.AuthOrderRepository
	userRepo         *repository.UserRepository
	kycRecordRepo    *repository.KycPersonalRepository
	kycEntRepo       *repository.KycEnterpriseRepository
	resourcePackRepo *repository.ResourcePackRepository
	balanceService   *BalanceService
	configRepo       *repository.SystemConfigRepository
}

// NewAuthService 创建认证服务
func NewAuthService(
	finAuthClient func() upstream.FinAuthInterface,
	finAuthCfg func() config.FinAuthConfig,
	orderRepo *repository.AuthOrderRepository,
	userRepo *repository.UserRepository,
	kycRecordRepo *repository.KycPersonalRepository,
	kycEntRepo *repository.KycEnterpriseRepository,
	resourcePackRepo *repository.ResourcePackRepository,
	balanceService *BalanceService,
	configRepo *repository.SystemConfigRepository,
) *AuthService {
	return &AuthService{
		finAuthClient:    finAuthClient,
		finAuthCfg:       finAuthCfg,
		orderRepo:        orderRepo,
		userRepo:         userRepo,
		kycRecordRepo:    kycRecordRepo,
		kycEntRepo:       kycEntRepo,
		resourcePackRepo: resourcePackRepo,
		balanceService:   balanceService,
		configRepo:       configRepo,
	}
}

// GetFreeAuthRemaining 返回账户实名（source=1）剩余免费认证次数
// 账户实名始终免费不计费，该次数仅作前端提示展示
func (s *AuthService) GetFreeAuthRemaining(userID int64) (int, error) {
	failCount, err := s.kycRecordRepo.CountUserFreeFailures(userID)
	if err != nil {
		return 0, err
	}
	remaining := freeFailLimit - failCount
	if remaining < 0 {
		remaining = 0
	}
	return remaining, nil
}

// StartAuth 发起认证
// source: 1-账户实名（Web） 2-API调用
//   - 账户实名（source=1）：完全不经过实名认证产品库（starloft_kyc），无认证订单、无计费，
//
// 认证信息单独储存在系统库的实名记录表（kyc_personal）中。
//   - API 调用（source=2）：走认证订单 + 计费流程。
//
// free: 保留参数，账户实名始终免费，不再使用。
func (s *AuthService) StartAuth(
	userID int64,
	name, idCard, returnURL, notifyURL string,
	bizExtraData string,
	source int,
	free bool,
) (*StartAuthResult, error) {
	// return_url 未传入时使用默认跳转地址（认证完成后返回账户实名页）
	if returnURL == "" {
		returnURL = finAuthReturnURL
	}

	// 生成全平台唯一业务流水号（纯随机数，唯一性由密码学随机源保证，调用时平台随机生成）
	bizNo := utils.GenerateRandomDigits(20)

	// 账户实名（source=1）：只写系统库实名记录，不产生认证订单、不计费
	if source == 1 {
		return s.startAccountAuth(userID, name, idCard, bizNo, returnURL, notifyURL, bizExtraData)
	}

	// 注意：notifyURL 由 API 流程（source=2）的下游显式传入。

	// 获取用户信息
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// API 调用（source=2）始终计费
	charge := true

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

	// 创建 KYC 认证记录（保存实名信息，关联订单流程）
	kycRecord := &model.KycPersonal{
		UserID: userID,
		Source: 2,
		BizNo:  bizNo,
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
		BizNo:        bizNo,
		UserID:       userID,
		ReturnURL:    returnURL,
		NotifyURL:    notifyURL,
		BizExtraData: bizExtraData,
		Cost:         kycPrice,
		Source:       source,
		PayType:      payType,
		Status:       0, // 待认证
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if payType == 2 {
		order.UserPackID = userPack.ID
	}

	err = s.orderRepo.CreateOrder(order)
	if err != nil {
		return nil, fmt.Errorf("create order failed: %w", err)
	}

	// 执行扣费（先扣资源包，再扣余额；扣费失败需回滚订单，避免悬挂订单）
	if err := s.deductOrderCharge(order, userPack, userID, kycPrice, kycRecord); err != nil {
		return nil, err
	}

	// 传给上游 get_token 的地址：
	// - return_url 透传下游地址：用户在上游完成认证后浏览器直接回到其 return_url
	// - notify_url 始终为平台回调地址：认证结果由平台异步接收并落地（更新订单、扣费/退款、下发 API Secret、通知下游）
	upstreamReturnURL := returnURL
	upstreamNotifyURL := finAuthNotifyURL

	// 调用上游 get_token
	req := &upstream.GetTokenRequest{
		SignVersion:    upstream.SignVersionHMACSHA256,
		ReturnURL:      upstreamReturnURL,
		NotifyURL:      upstreamNotifyURL,
		BizNo:          bizNo,
		SceneID:        s.finAuthCfg().SceneID,
		ComparisonType: "1", // 人脸核身模式
		UUID:           fmt.Sprintf("%d", userID),
		BizExtraData:   bizExtraData,
		IDCardMode:     "0", // 不拍摄身份证，直接传入姓名和身份证号
		IDCardName:     name,
		IDCardNumber:   idCard,
	}

	tokenResp, err := s.finAuthClient().GetToken(req)
	if err != nil {
		log.Printf("获取 FinAuth Token 失败: %v", err)
		// 将结果标记为连接超时并退还预扣费用（资源包/余额）
		s.revertStartAuth(order, kycRecord, "连接超时退款")
		return nil, fmt.Errorf("get token failed: %w", err)
	}

	// 更新订单上游信息
	err = s.orderRepo.UpdateOrderUpstreamInfo(order.ID, tokenResp.Token, tokenResp.BizID, tokenResp.RequestID)
	if err != nil {
		log.Printf("更新订单上游信息失败 [order_id=%d]: %v", order.ID, err)
		// 费用已扣但 token 未写入，退还预扣费用并将订单标记为超时退款，避免悬挂
		s.revertStartAuth(order, kycRecord, "写入token失败退款")
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

// startAccountAuth 账户实名（source=1）发起认证：
// 不写认证订单、不做任何计费，实名信息单独储存在系统库的实名记录表（kyc_record）中。
func (s *AuthService) startAccountAuth(
	userID int64,
	name, idCard, bizNo, returnURL, notifyURL, bizExtraData string,
) (*StartAuthResult, error) {
	// 创建实名记录（状态：认证中）
	kycRecord := &model.KycPersonal{
		UserID:       userID,
		Source:       1,
		BizNo:        bizNo,
		ReturnURL:    returnURL,
		NotifyURL:    notifyURL,
		BizExtraData: bizExtraData,
		Name:         name,
		IDCard:       idCard,
		Status:       1,
	}
	if err := s.kycRecordRepo.Create(kycRecord); err != nil {
		return nil, fmt.Errorf("create kyc record failed: %w", err)
	}

	// 传给上游 get_token 的地址：
	// - return_url 透传用户地址：用户在上游完成认证后浏览器直接回到其 return_url
	// - notify_url 始终为平台回调地址：认证结果由平台异步接收并落地（更新实名记录、下发 API Secret）
	upstreamReturnURL := returnURL
	upstreamNotifyURL := finAuthNotifyURL

	req := &upstream.GetTokenRequest{
		SignVersion:    upstream.SignVersionHMACSHA256,
		ReturnURL:      upstreamReturnURL,
		NotifyURL:      upstreamNotifyURL,
		BizNo:          bizNo,
		SceneID:        s.finAuthCfg().SceneID,
		ComparisonType: "1", // 人脸核身模式
		UUID:           fmt.Sprintf("%d", userID),
		BizExtraData:   bizExtraData,
		IDCardMode:     "0", // 不拍摄身份证，直接传入姓名和身份证号
		IDCardName:     name,
		IDCardNumber:   idCard,
	}

	tokenResp, err := s.finAuthClient().GetToken(req)
	if err != nil {
		log.Printf("获取 FinAuth Token 失败: %v", err)
		// 账户实名不计费，仅将实名记录标记为失败（连接超时）
		_ = s.kycRecordRepo.UpdateResult(kycRecord.ID, 3, "TIMEOUT", "连接超时", "", nil)
		return nil, fmt.Errorf("get token failed: %w", err)
	}

	// 写入记录上游信息（token/biz_id/request_id）
	if err := s.kycRecordRepo.UpdateUpstreamInfo(kycRecord.ID, tokenResp.Token, tokenResp.BizID, tokenResp.RequestID); err != nil {
		log.Printf("更新实名记录上游信息失败 [record_id=%d]: %v", kycRecord.ID, err)
		_ = s.kycRecordRepo.UpdateResult(kycRecord.ID, 3, "TIMEOUT", "写入token失败", "", nil)
		return nil, fmt.Errorf("update kyc record upstream info failed: %w", err)
	}

	authURL := s.BuildAuthURL(tokenResp.Token)

	return &StartAuthResult{
		Record:  kycRecord,
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
	kycRecord *model.KycPersonal,
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
func (s *AuthService) markOrderDeductFailed(order *model.AuthOrder, kycRecord *model.KycPersonal) {
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
	return fmt.Sprintf("%s/finauth/lite/do?token=%s", s.finAuthCfg().BaseURL, token)
}

// GetAuthResult 查询认证结果
func (s *AuthService) GetAuthResult(userID int64, bizNo string) (*model.AuthOrder, error) {
	order, err := s.orderRepo.GetOrderByBizNo(bizNo)
	if err != nil {
		return nil, err
	}

	// 越权防护：按 biz_no 查询不受 user_id 过滤，必须显式校验订单归属
	if order.UserID != userID {
		return nil, errors.New("order not found")
	}

	// 如果订单还在处理中，尝试从上游查询最新结果
	if order.Status == 0 || order.Status == 1 {
		s.syncOrderResult(order)
		// 重新查询
		order, _ = s.orderRepo.GetOrderByBizNo(order.BizNo)
	}

	return order, nil
}

// GetUserByID 查询用户
func (s *AuthService) GetUserByID(userID int64) (*model.User, error) {
	return s.userRepo.GetUserByID(userID)
}

// GetLatestKycRecord 获取用户最新 KYC 记录
func (s *AuthService) GetLatestKycRecord(userID int64) (*model.KycPersonal, error) {
	return s.kycRecordRepo.GetLatestByUserID(userID)
}

// StartEnterpriseAuth 发起企业实名（自助）：
// 创建企业实名记录，复用上游人脸核身（FinAuth）对法人进行扫脸认证，返回跳转地址。
// 企业实名信息在法人扫脸通过后落地 user 表（is_kyc_verified=2）。
func (s *AuthService) StartEnterpriseAuth(
	userID int64,
	companyName, creditCode, legalName, legalIDCard, returnURL string,
) (*StartAuthResult, error) {
	// 已企业实名则无需重复认证
	if user, err := s.userRepo.GetUserByID(userID); err == nil && user.IsKYCVerified == 2 {
		return nil, errors.New("已企业实名，无需重复认证")
	}

	if returnURL == "" {
		returnURL = "/user/kyc"
	}

	bizNo := utils.GenerateRandomDigits(20)

	// 创建企业实名记录（状态：待法人扫脸）
	rec := &model.KycEnterprise{
		UserID:           userID,
		BizNo:            bizNo,
		CompanyName:      companyName,
		CreditCode:       creditCode,
		LegalName:        legalName,
		LegalIDCard:      legalIDCard,
		Source:           0, // 自助
		FourFactorStatus: 0,
		Status:           0,
	}
	if err := s.kycEntRepo.Create(rec); err != nil {
		return nil, fmt.Errorf("create kyc enterprise record failed: %w", err)
	}

	req := &upstream.GetTokenRequest{
		SignVersion:    upstream.SignVersionHMACSHA256,
		ReturnURL:      returnURL,
		NotifyURL:      finAuthNotifyURL,
		BizNo:          bizNo,
		SceneID:        s.finAuthCfg().SceneID,
		ComparisonType: "1", // 人脸核身模式
		UUID:           fmt.Sprintf("%d", userID),
		IDCardMode:     "0", // 直接传入法人姓名与身份证号
		IDCardName:     legalName,
		IDCardNumber:   legalIDCard,
	}

	tokenResp, err := s.finAuthClient().GetToken(req)
	if err != nil {
		_ = s.kycEntRepo.UpdateResult(rec.ID, 3, "TIMEOUT", "连接超时", "", nil)
		return nil, fmt.Errorf("get token failed: %w", err)
	}

	if err := s.kycEntRepo.UpdateUpstreamInfo(rec.ID, tokenResp.Token, tokenResp.BizID, tokenResp.RequestID); err != nil {
		_ = s.kycEntRepo.UpdateResult(rec.ID, 3, "TIMEOUT", "写入token失败", "", nil)
		return nil, fmt.Errorf("update kyc enterprise upstream info failed: %w", err)
	}

	return &StartAuthResult{
		AuthURL: s.BuildAuthURL(tokenResp.Token),
		Token:   tokenResp.Token,
		BizID:   tokenResp.BizID,
	}, nil
}

// GetEnterpriseAuthRecord 获取用户最新企业实名记录
func (s *AuthService) GetEnterpriseAuthRecord(userID int64) (*model.KycEnterprise, error) {
	return s.kycEntRepo.GetLatestByUserID(userID)
}

// BuildEnterpriseAuthURL 企业实名记录为「待法人扫脸」且已有上游 token 时返回继续认证地址
func (s *AuthService) BuildEnterpriseAuthURL(rec *model.KycEnterprise) string {
	if rec == nil || rec.Status != 1 || rec.UpToken == "" {
		return ""
	}
	return s.BuildAuthURL(rec.UpToken)
}

// applyEnterpriseCallback 落地企业实名回调结果：只有法人扫脸通过时才更新 user 表实名信息
func (s *AuthService) applyEnterpriseCallback(rec *model.KycEnterprise, notifyData *upstream.NotifyData) error {
	if isInProgressMessage(notifyData.ResultMessage) {
		return nil
	}

	status := 3
	switch notifyData.ResultCode {
	case 1000:
		status = 2 // 认证成功
	}
	resultCode := fmt.Sprintf("%d", notifyData.ResultCode)
	var verifiedAt *time.Time
	if status == 2 {
		t := time.Now()
		verifiedAt = &t
	}
	if err := s.kycEntRepo.UpdateResult(rec.ID, status, resultCode, notifyData.ResultMessage, "", verifiedAt); err != nil {
		return fmt.Errorf("update kyc enterprise result failed: %w", err)
	}
	if status == 2 {
		s.applyUserVerified(rec.UserID, 2, rec.CompanyName, rec.CreditCode)
	}
	return nil
}

// AdminCreateEnterpriseManual 后台人工企业实名（=公户验证）：
// 管理员录入企业名称与统一社会信用代码，直接为企业开通企业实名
func (s *AuthService) AdminCreateEnterpriseManual(userID int64, companyName, creditCode string, adminID int64) error {
	bizNo := utils.GenerateRandomDigits(20)
	rec := &model.KycEnterprise{
		UserID:      userID,
		BizNo:       bizNo,
		CompanyName: companyName,
		CreditCode:  creditCode,
		Source:      1, // 后台人工
		AdminID:     adminID,
		Status:      2, // 已通过
	}
	if err := s.kycEntRepo.Create(rec); err != nil {
		return err
	}
	s.applyUserVerified(userID, 2, companyName, creditCode)
	return nil
}

// ListEnterpriseRecords 企业实名记录列表（管理后台）
func (s *AuthService) ListEnterpriseRecords(page, pageSize int) ([]*model.KycEnterprise, int64, error) {
	return s.kycEntRepo.GetEnterpriseRecords(page, pageSize)
}

// GetLatestPendingOrder 获取用户最新进行中的认证订单（用于继续认证）
func (s *AuthService) GetLatestPendingOrder(userID int64) (*model.AuthOrder, error) {
	return s.orderRepo.GetLatestPendingOrder(userID)
}

// SyncKycRecord 同步用户最新进行中实名记录的上游结果（账户实名 source=1）
func (s *AuthService) SyncKycRecord(userID int64) (*model.KycPersonal, error) {
	record, err := s.kycRecordRepo.GetPendingByUserID(userID)
	if err != nil {
		return nil, err
	}
	s.syncKycRecordResult(record)
	// 重新查询，获取最新状态
	latest, _ := s.kycRecordRepo.GetLatestByUserID(userID)
	return latest, nil
}

// syncKycRecordResult 同步实名记录结果（从上游查询，账户实名 source=1）
func (s *AuthService) syncKycRecordResult(record *model.KycPersonal) {
	if record.UpBizID == "" {
		return
	}

	req := &upstream.GetResultRequest{
		BizID:       record.UpBizID,
		SignVersion: upstream.SignVersionHMACSHA256,
	}

	result, err := s.finAuthClient().GetResult(req)
	if err != nil {
		log.Printf("同步实名记录结果失败 [record_id=%d, biz_id=%s]: %v", record.ID, record.UpBizID, err)
		return
	}

	// 未开始/进行中：认证尚未完结，保持「认证中」，不结束流程
	if isInProgressMessage(result.ResultMessage) {
		log.Printf("认证尚未开始或进行中，保持认证中状态 [record_id=%d, result_message=%s]", record.ID, result.ResultMessage)
		return
	}

	// 判断结果状态
	status := record.Status
	switch result.ResultCode {
	case 1000:
		status = 2 // 认证成功
	default:
		if result.ResultCode != 0 {
			status = 3
		}
	}

	resultCode := fmt.Sprintf("%d", result.ResultCode)
	var verifiedAt *time.Time
	if status == 2 {
		t := time.Now()
		verifiedAt = &t
	}
	if err := s.kycRecordRepo.UpdateResult(record.ID, status, resultCode, result.ResultMessage, "", verifiedAt); err != nil {
		log.Printf("更新实名记录结果失败 [record_id=%d]: %v", record.ID, err)
		return
	}
	record.Status = status
	record.ResultCode = resultCode
	record.ResultMessage = result.ResultMessage

	// 实名成功：更新用户实名信息并生成下发 API Secret
	if status == 2 {
		s.applyUserVerified(record.UserID, 1, record.Name, record.IDCard)
	}
}

// applyRecordCallback 落地账户实名（source=1）回调结果：更新实名记录、实名成功时下发用户实名信息与 API Secret
func (s *AuthService) applyRecordCallback(record *model.KycPersonal, notifyData *upstream.NotifyData) error {
	// 未开始/进行中：认证尚未完结，忽略本次回调，保持「认证中」
	if isInProgressMessage(notifyData.ResultMessage) {
		log.Printf("回调表示实名认证尚未开始或进行中，忽略 [record_id=%d, biz_id=%s, result_message=%s]", record.ID, notifyData.BizInfo.BizID, notifyData.ResultMessage)
		return nil
	}

	// 判断认证结果
	status := record.Status
	switch notifyData.ResultCode {
	case 1000:
		status = 2 // 认证成功
	default:
		status = 3
	}

	resultCode := fmt.Sprintf("%d", notifyData.ResultCode)
	var verifiedAt *time.Time
	if status == 2 {
		t := time.Now()
		verifiedAt = &t
	}
	if err := s.kycRecordRepo.UpdateResult(record.ID, status, resultCode, notifyData.ResultMessage, "", verifiedAt); err != nil {
		return fmt.Errorf("update kyc record result failed: %w", err)
	}

	// 实名成功：更新用户实名信息并生成下发 API Secret
	if status == 2 {
		s.applyUserVerified(record.UserID, 1, record.Name, record.IDCard)
	}
	return nil
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

// GetUserAuthRecords 查询用户认证记录（来自系统库实名记录表）
func (s *AuthService) GetUserAuthRecords(userID int64, page, pageSize int) ([]*model.KycPersonal, int64, error) {
	records, total, err := s.kycRecordRepo.GetUserKycRecords(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	for _, r := range records {
		r.Cost = 0 // 账户实名免费，恒为0（仅用于列表展示）
	}
	return records, total, nil
}

// GetUserAuthCallStats 统计用户近 N 天的认证调用次数（按天，缺失日期补零）
func (s *AuthService) GetUserAuthCallStats(userID int64, days int) ([]string, []int64, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -(days - 1))

	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")

	stats, err := s.kycRecordRepo.GetUserDailyAuthCount(userID, startStr, endStr)
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

// applyUserVerified 实名成功后落地 user 表实名信息（仅成功后更新）
// verified: 1-个人实名 2-企业实名。
// 更换规则：已企业实名（is_kyc_verified=2）后不得降级为个人实名，企业实名结果保持不变；
// 未实名（0）或已个人实名（1）可升级为企业实名。
func (s *AuthService) applyUserVerified(userID int64, verified int, name, number string) {
	if verified == 1 {
		current, err := s.userRepo.GetUserByID(userID)
		if err != nil {
			log.Printf("查询用户失败，跳过个人实名落地 [user_id=%d]: %v", userID, err)
			return
		}
		// 已企业实名：不允许降级为个人实名，忽略本次个人实名结果
		if current.IsKYCVerified == 2 {
			log.Printf("用户已企业实名，忽略个人实名落地 [user_id=%d]", userID)
			return
		}
	}
	if err := s.userRepo.UpdateUserKYCInfo(userID, verified, name, number); err != nil {
		log.Printf("更新用户实名信息失败 [user_id=%d]: %v", userID, err)
		return
	}
	s.ensureAPISecret(userID)
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

	result, err := s.finAuthClient().GetResult(req)
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

	// 更新 kyc_personal 状态
	if status == 2 {
		kycRecord, err := s.kycRecordRepo.GetLatestByUserID(order.UserID)
		if err == nil && kycRecord != nil {
			now := time.Now()
			_ = s.kycRecordRepo.UpdateResult(kycRecord.ID, 2, resultCode, result.ResultMessage, "", &now)
		}
		if err == nil && kycRecord != nil {
			s.applyUserVerified(order.UserID, 1, kycRecord.Name, kycRecord.IDCard)
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
	if !s.finAuthClient().VerifySign(data, sign) {
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
		// 未查到订单：账户实名（source=1）不写认证订单，改按系统库实名记录处理（无订单、无计费）
		if record, rerr := s.kycRecordRepo.GetByUpBizID(notifyData.BizInfo.BizID); rerr == nil && record != nil {
			return s.applyRecordCallback(record, &notifyData)
		}
		// 企业实名（法人扫脸）也未写认证订单，按企业实名记录处理
		if ent, eerr := s.kycEntRepo.GetByUpBizID(notifyData.BizInfo.BizID); eerr == nil && ent != nil {
			return s.applyEnterpriseCallback(ent, &notifyData)
		}
		log.Printf("查找订单及实名记录均失败 [biz_id=%s]: %v", notifyData.BizInfo.BizID, err)
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
			now := time.Now()
			_ = s.kycRecordRepo.UpdateResult(kycRecord.ID, 2, resultCode, notifyData.ResultMessage, "", &now)
		}
		if err == nil && kycRecord != nil {
			s.applyUserVerified(order.UserID, 1, kycRecord.Name, kycRecord.IDCard)
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

// NotifyDownstream 通知下游：将认证结果 POST 到下游的 notify_url（携带 HMAC 签名，供下游校验防伪造）
func (s *AuthService) NotifyDownstream(order *model.AuthOrder) {
	if order.NotifyURL == "" {
		return
	}

	// 使用订单所属平台用户的 api_secret 生成签名（下游用同一 secret 校验）
	sign := ""
	if user, err := s.userRepo.GetUserByID(order.UserID); err == nil && user != nil && user.APISecret != "" {
		sign = buildNotifySign(user.APISecret, order)
	}

	payload := map[string]interface{}{
		"biz_no":         order.BizNo,
		"status":         order.Status,
		"result_code":    order.ResultCode,
		"result_message": order.ResultMessage,
		"cost":           order.Cost,
		"sign":           sign,
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

// buildNotifySign 构造下游回调签名：
// 对固定字段按 key 字典序拼接为 k=v&k=v... 的原始字符串（不做 URL 编码），
// 再以 HMAC-SHA256(api_secret, canonical) 计算十六进制小写签名。
// 下游（如 zjmf_v10 插件）用相同算法与自己的 api_secret 校验，杜绝伪造回调。
func buildNotifySign(apiSecret string, order *model.AuthOrder) string {
	fields := map[string]string{
		"biz_no":         order.BizNo,
		"cost":           strconv.FormatFloat(order.Cost, 'f', 2, 64),
		"result_code":    order.ResultCode,
		"result_message": order.ResultMessage,
		"status":         strconv.Itoa(order.Status),
	}

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte('&')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(fields[k])
	}

	mac := hmac.New(sha256.New, []byte(apiSecret))
	mac.Write([]byte(sb.String()))
	return hex.EncodeToString(mac.Sum(nil))
}

// revertStartAuth 发起认证失败时退还预扣费用并将结果标记为连接超时（免费流程不涉及退费）
func (s *AuthService) revertStartAuth(order *model.AuthOrder, kycRecord *model.KycPersonal, remark string) {
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

	// 账户实名（source=1）：同步处理中的实名记录（无订单）
	records, err := s.kycRecordRepo.GetPendingRecords()
	if err != nil {
		return err
	}
	for _, record := range records {
		s.syncKycRecordResult(record)
	}
	return nil
}
