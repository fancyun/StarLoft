package service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"starloftrpa/internal/model"
	"starloftrpa/internal/repository"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
)

type BalanceService struct {
	userRepo         *repository.UserRepository
	balanceLogRepo   *repository.BalanceLogRepository
	paymentRepo      *repository.PaymentOrderRepository
	resourcePackRepo *repository.ResourcePackRepository
	configRepo       *repository.SystemConfigRepository
	db               *sql.DB
}

func NewBalanceService(
	userRepo *repository.UserRepository,
	balanceLogRepo *repository.BalanceLogRepository,
	paymentRepo *repository.PaymentOrderRepository,
	resourcePackRepo *repository.ResourcePackRepository,
	configRepo *repository.SystemConfigRepository,
	db *sql.DB,
) *BalanceService {
	return &BalanceService{
		userRepo:         userRepo,
		balanceLogRepo:   balanceLogRepo,
		paymentRepo:      paymentRepo,
		resourcePackRepo: resourcePackRepo,
		configRepo:       configRepo,
		db:               db,
	}
}

// DeductBalance 扣除余额（预扣费）
func (s *BalanceService) DeductBalance(userID int64, amount float64, orderID int64, remark string) error {
	// 开启事务
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 在事务中锁定并查询用户当前余额
	balance, err := s.userRepo.GetBalanceForUpdateTx(tx, userID)
	if err != nil {
		return err
	}

	// 检查余额是否充足
	if balance < amount {
		return ErrInsufficientBalance
	}

	// 扣除余额
	newBalance := balance - amount
	if err := s.userRepo.UpdateUserBalanceTx(tx, userID, newBalance); err != nil {
		return err
	}

	// 记录余额流水
	log := &model.BalanceLog{
		UserID:        userID,
		OrderID:       orderID,
		Type:          2, // 消费
		Amount:        amount,
		BalanceBefore: balance,
		BalanceAfter:  newBalance,
		Remark:        remark,
	}
	if err := s.balanceLogRepo.CreateLogTx(tx, log); err != nil {
		return err
	}

	return tx.Commit()
}

// CreateRecharge 创建充值订单（channel：alipay-支付宝 wechat-微信）
func (s *BalanceService) CreateRecharge(userID int64, amount float64, channel string) (*model.PaymentOrder, error) {
	if channel != "alipay" && channel != "wechat" {
		return nil, fmt.Errorf("不支持的支付渠道: %s", channel)
	}

	// 生成支付订单号
	payOrderNo := fmt.Sprintf("R%s%06d", time.Now().Format("20060102150405"), time.Now().UnixNano()%1000000)

	// 创建支付订单
	expireTime := time.Now().Add(30 * time.Minute)
	order := &model.PaymentOrder{
		PayOrderNo:   payOrderNo,
		UserID:       userID,
		Amount:       amount,
		Channel:      channel,
		Status:       0, // 待支付
		RefundStatus: 0,
		ExpireTime:   &expireTime,
	}

	err := s.paymentRepo.CreateOrder(order)
	if err != nil {
		return nil, err
	}

	return order, nil
}

// GetPaymentOrder 根据支付订单号查询订单
func (s *BalanceService) GetPaymentOrder(payOrderNo string) (*model.PaymentOrder, error) {
	return s.paymentRepo.GetOrderByPayOrderNo(payOrderNo)
}

// creditBalance 增加用户余额并记录流水（通用入账，type 区分充值等）
func (s *BalanceService) creditBalance(userID int64, amount float64, logType int, remark, bankSerialNo string) error {
	// 开启事务
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 在事务中锁定并查询用户当前余额
	balance, err := s.userRepo.GetBalanceForUpdateTx(tx, userID)
	if err != nil {
		return err
	}

	// 增加余额
	newBalance := balance + amount
	if err := s.userRepo.UpdateUserBalanceTx(tx, userID, newBalance); err != nil {
		return err
	}

	// 记录余额流水
	log := &model.BalanceLog{
		UserID:        userID,
		Type:          logType,
		Amount:        amount,
		BalanceBefore: balance,
		BalanceAfter:  newBalance,
		BankSerialNo:  bankSerialNo,
		Remark:        remark,
	}
	if err := s.balanceLogRepo.CreateLogTx(tx, log); err != nil {
		return err
	}

	return tx.Commit()
}

// ManualRechargeBalance 人工充值（type=1，增加余额）
// bankSerialNo 为银行流水单号（必填，用于对账）
func (s *BalanceService) ManualRechargeBalance(userID int64, amount float64, remark, bankSerialNo string) error {
	return s.creditBalance(userID, amount, 1, remark, bankSerialNo)
}

// PurchaseResourcePack 使用余额购买资源包（从余额扣费，不支持直接为资源包付费）
// 在单个事务中完成：校验/扣减库存 → 锁定余额并扣费 → 记录余额流水 → 创建用户资源包
func (s *BalanceService) PurchaseResourcePack(userID, packID int64) (*model.UserResourcePack, error) {
	// 查询资源包
	pack, err := s.resourcePackRepo.GetPackByID(packID)
	if err != nil {
		return nil, err
	}
	if pack.Status != 1 {
		return nil, repository.ErrPackOffSale
	}

	// 开启事务
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 扣减库存（锁定库存行，返回 false 表示已售罄）
	ok, err := s.resourcePackRepo.DecrementStockTx(tx, packID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, repository.ErrPackSoldOut
	}

	// 锁定余额并校验是否充足
	balance, err := s.userRepo.GetBalanceForUpdateTx(tx, userID)
	if err != nil {
		return nil, err
	}
	if balance < pack.Price {
		return nil, ErrInsufficientBalance
	}

	// 扣除余额
	newBalance := balance - pack.Price
	if err := s.userRepo.UpdateUserBalanceTx(tx, userID, newBalance); err != nil {
		return nil, err
	}

	// 记录余额流水
	log := &model.BalanceLog{
		UserID:        userID,
		Type:          2, // 消费
		Amount:        pack.Price,
		BalanceBefore: balance,
		BalanceAfter:  newBalance,
		Remark:        fmt.Sprintf("购买资源包：%s", pack.Name),
	}
	if err := s.balanceLogRepo.CreateLogTx(tx, log); err != nil {
		return nil, err
	}

	// 创建用户资源包（次数快照）
	up := &model.UserResourcePack{
		UserID:         userID,
		PackID:         pack.ID,
		PackName:       pack.Name,
		TotalCount:     pack.TotalCount,
		RemainingCount: pack.TotalCount,
		Status:         1,
	}
	if err := s.resourcePackRepo.CreateUserPackTx(tx, up); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return up, nil
}

// RefundBalance 退还余额
func (s *BalanceService) RefundBalance(userID int64, amount float64, orderID int64, remark string) error {
	// 开启事务
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 在事务中锁定并查询用户当前余额
	balance, err := s.userRepo.GetBalanceForUpdateTx(tx, userID)
	if err != nil {
		return err
	}

	// 增加余额
	newBalance := balance + amount
	if err := s.userRepo.UpdateUserBalanceTx(tx, userID, newBalance); err != nil {
		return err
	}

	// 记录余额流水
	log := &model.BalanceLog{
		UserID:        userID,
		OrderID:       orderID,
		Type:          3, // 退款
		Amount:        amount,
		BalanceBefore: balance,
		BalanceAfter:  newBalance,
		Remark:        remark,
	}
	if err := s.balanceLogRepo.CreateLogTx(tx, log); err != nil {
		return err
	}

	return tx.Commit()
}

// RechargeBalance 充值到账
func (s *BalanceService) RechargeBalance(userID int64, amount float64, orderID int64) error {
	// 开启事务
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 在事务中锁定并查询用户当前余额
	balance, err := s.userRepo.GetBalanceForUpdateTx(tx, userID)
	if err != nil {
		return err
	}

	// 增加余额
	newBalance := balance + amount
	if err := s.userRepo.UpdateUserBalanceTx(tx, userID, newBalance); err != nil {
		return err
	}

	// 记录充值流水
	log := &model.BalanceLog{
		UserID:        userID,
		OrderID:       orderID,
		Type:          1, // 充值
		Amount:        amount,
		BalanceBefore: balance,
		BalanceAfter:  newBalance,
		Remark:        "充值到账",
	}
	if err := s.balanceLogRepo.CreateLogTx(tx, log); err != nil {
		return err
	}

	return tx.Commit()
}
