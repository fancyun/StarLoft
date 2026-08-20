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
	userRepo       *repository.UserRepository
	balanceLogRepo *repository.BalanceLogRepository
	paymentRepo    *repository.PaymentOrderRepository
	configRepo     *repository.SystemConfigRepository
	db             *sql.DB
}

func NewBalanceService(
	userRepo *repository.UserRepository,
	balanceLogRepo *repository.BalanceLogRepository,
	paymentRepo *repository.PaymentOrderRepository,
	configRepo *repository.SystemConfigRepository,
	db *sql.DB,
) *BalanceService {
	return &BalanceService{
		userRepo:       userRepo,
		balanceLogRepo: balanceLogRepo,
		paymentRepo:    paymentRepo,
		configRepo:     configRepo,
		db:             db,
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

// CreateRecharge 创建充值订单
func (s *BalanceService) CreateRecharge(userID int64, amount float64) (*model.PaymentOrder, error) {
	// 生成支付订单号
	payOrderNo := fmt.Sprintf("R%s%06d", time.Now().Format("20060102150405"), time.Now().UnixNano()%1000000)

	// 创建支付订单
	expireTime := time.Now().Add(30 * time.Minute)
	order := &model.PaymentOrder{
		PayOrderNo:   payOrderNo,
		UserID:       userID,
		Amount:       amount,
		Channel:      "unionpay",
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

// creditBalance 增加用户余额并记录流水（通用入账，type 区分充值/赠送等）
func (s *BalanceService) creditBalance(userID int64, amount float64, logType int, remark string) error {
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
		Remark:        remark,
	}
	if err := s.balanceLogRepo.CreateLogTx(tx, log); err != nil {
		return err
	}

	return tx.Commit()
}

// ManualRechargeBalance 人工充值（type=1，增加余额）
func (s *BalanceService) ManualRechargeBalance(userID int64, amount float64, remark string) error {
	return s.creditBalance(userID, amount, 1, remark)
}

// GiftBalance 赠送余额（type=4，增加余额）
func (s *BalanceService) GiftBalance(userID int64, amount float64, remark string) error {
	return s.creditBalance(userID, amount, 4, remark)
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
