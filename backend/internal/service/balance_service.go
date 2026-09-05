package service

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
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
	db               *sql.DB
}

func NewBalanceService(
	userRepo *repository.UserRepository,
	balanceLogRepo *repository.BalanceLogRepository,
	paymentRepo *repository.PaymentOrderRepository,
	resourcePackRepo *repository.ResourcePackRepository,
	db *sql.DB,
) *BalanceService {
	return &BalanceService{
		userRepo:         userRepo,
		balanceLogRepo:   balanceLogRepo,
		paymentRepo:      paymentRepo,
		resourcePackRepo: resourcePackRepo,
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

	order := s.newPaymentOrder(userID, amount, channel, paymentIntentRecharge, "", 0, 0)

	err := s.paymentRepo.CreateOrder(order)
	if err != nil {
		return nil, err
	}

	return order, nil
}

// 支付用途：充值 / 购买资源包
const (
	paymentIntentRecharge     = "recharge"
	paymentIntentResourcePack = "resource_pack"
	// paymentExpireMinutes 待支付订单的过期时间（分钟），超时未支付则关闭（资源包需释放库存）
	paymentExpireMinutes = 30
)

// newPaymentOrder 构造待支付支付单（未落库）
func (s *BalanceService) newPaymentOrder(userID int64, amount float64, channel, intent, bizNo string, balanceAmount float64, stockReserved int) *model.PaymentOrder {
	payOrderNo := fmt.Sprintf("R%s%06d", time.Now().Format("20060102150405"), time.Now().UnixNano()%1000000)
	expireTime := time.Now().Add(paymentExpireMinutes * time.Minute)
	return &model.PaymentOrder{
		PayOrderNo:    payOrderNo,
		UserID:        userID,
		Amount:        amount,
		Channel:       channel,
		Status:        0, // 待支付
		RefundStatus:  0,
		ExpireTime:    &expireTime,
		Intent:        intent,
		BizNo:         bizNo,
		BalanceAmount: balanceAmount,
		StockReserved: stockReserved,
	}
}

// PurchaseResourcePackResult 购买资源包结果
type PurchaseResourcePackResult struct {
	UserPack            *model.UserResourcePack // 全部由余额支付时直接发放的资源包
	PaymentOrder        *model.PaymentOrder     // 需要外部支付时的待支付单
	FullyPaidByBalance  bool                    // 是否已由余额全额支付
	BalanceAmount       float64                 // 本次支付的余额部分
	ExternalAmount      float64                 // 本次支付的外部支付部分（支付宝/微信）
}

// PurchaseResourcePackWithPay 购买资源包（支持组合支付：余额支付一部分，支付宝/微信支付剩余部分，无需先充值再购买）
// 下单即占库存，待支付单 30 分钟未支付由定时任务释放库存并退还余额支付部分。
func (s *BalanceService) PurchaseResourcePackWithPay(userID, packID int64, channel string) (*PurchaseResourcePackResult, error) {
	if channel != "alipay" && channel != "wechat" {
		return nil, fmt.Errorf("不支持的支付渠道: %s", channel)
	}

	pack, err := s.resourcePackRepo.GetPackByID(packID)
	if err != nil {
		return nil, err
	}
	if pack.Status != 1 {
		return nil, repository.ErrPackOffSale
	}

	// 开启事务：占库存 + 扣余额 + 发放/下单，原子完成
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 下单即占库存（有限库存扣减，不限量不占）
	reserved, err := s.resourcePackRepo.ReserveStockTx(tx, packID)
	if err != nil {
		return nil, err
	}

	// 锁定余额，计算组合支付各部分
	balance, err := s.userRepo.GetBalanceForUpdateTx(tx, userID)
	if err != nil {
		return nil, err
	}
	balancePart := pack.Price
	if balance < balancePart {
		balancePart = balance
	}
	externalPart := pack.Price - balancePart

	// 扣除余额部分并记录流水
	if balancePart > 0 {
		newBalance := balance - balancePart
		if err := s.userRepo.UpdateUserBalanceTx(tx, userID, newBalance); err != nil {
			return nil, err
		}
		log := &model.BalanceLog{
			UserID:        userID,
			Type:          2, // 消费
			Amount:        balancePart,
			BalanceBefore: balance,
			BalanceAfter:  newBalance,
			Remark:        fmt.Sprintf("购买资源包：%s（余额支付）", pack.Name),
		}
		if err := s.balanceLogRepo.CreateLogTx(tx, log); err != nil {
			return nil, err
		}
	}

	// 余额已全额覆盖：直接发放资源包
	if externalPart <= 0 {
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
		return &PurchaseResourcePackResult{
			UserPack:           up,
			FullyPaidByBalance: true,
			BalanceAmount:      balancePart,
		}, nil
	}

	// 需要外部支付：创建待支付单（资源包已占库存）
	order := s.newPaymentOrder(userID, externalPart, channel, paymentIntentResourcePack, fmt.Sprintf("%d", pack.ID), balancePart, boolToInt(reserved))
	if err := s.paymentRepo.CreateOrderTx(tx, order); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &PurchaseResourcePackResult{
		PaymentOrder:   order,
		BalanceAmount:  balancePart,
		ExternalAmount: externalPart,
	}, nil
}

// SettleResourcePackPaid 支付成功落地资源包订单：原子完成「标记已支付 + 发放资源包」，幂等（同时刻仅一次成功）
func (s *BalanceService) SettleResourcePackPaid(orderID int64, channelTradeNo string) error {
	order, err := s.paymentRepo.GetOrderByID(orderID)
	if err != nil {
		return err
	}
	packID, err := strconv.ParseInt(order.BizNo, 10, 64)
	if err != nil {
		return fmt.Errorf("资源包支付单关联单号非法: %w", err)
	}
	pack, err := s.resourcePackRepo.GetPackByID(packID)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	changed, err := s.paymentRepo.MarkOrderPaidIfPendingTx(tx, orderID, channelTradeNo)
	if err != nil {
		return err
	}
	if !changed {
		// 已处理过（幂等），直接成功
		return tx.Commit()
	}

	up := &model.UserResourcePack{
		UserID:         order.UserID,
		PackID:         pack.ID,
		PackName:       pack.Name,
		TotalCount:     pack.TotalCount,
		RemainingCount: pack.TotalCount,
		Status:         1,
	}
	if err := s.resourcePackRepo.CreateUserPackTx(tx, up); err != nil {
		return err
	}
	return tx.Commit()
}

// ReleaseExpiredResourcePackOrders 释放超时未支付的资源包订单（定时任务调用）：
// 关闭订单、退还已占库存、退还余额支付部分。
func (s *BalanceService) ReleaseExpiredResourcePackOrders() error {
	orders, err := s.paymentRepo.GetExpiredPendingResourcePackOrders(time.Now())
	if err != nil {
		return err
	}
	for _, o := range orders {
		s.releaseResourcePackOrder(o)
	}
	return nil
}

// releaseResourcePackOrder 释放单个超时资源包订单，事务内原子关闭（幂等）
func (s *BalanceService) releaseResourcePackOrder(order *model.PaymentOrder) {
	tx, err := s.db.Begin()
	if err != nil {
		log.Printf("释放资源包订单开启事务失败 [order_id=%d]: %v", order.ID, err)
		return
	}
	defer tx.Rollback()

	// 仅关闭仍为待支付的订单，避免与支付回调竞争导致重复退款/释放
	changed, err := s.paymentRepo.CloseOrderIfPendingTx(tx, order.ID)
	if err != nil {
		log.Printf("关闭超时资源包订单失败 [order_id=%d]: %v", order.ID, err)
		return
	}
	if !changed {
		return
	}

	// 加回已占库存
	if order.StockReserved == 1 {
		packID, perr := strconv.ParseInt(order.BizNo, 10, 64)
		if perr != nil {
			log.Printf("资源包订单关联单号非法，跳过库存释放 [order_id=%d, biz_no=%s]: %v", order.ID, order.BizNo, perr)
		} else {
			if err := s.resourcePackRepo.ReleaseStockTx(tx, packID); err != nil {
				log.Printf("释放资源包库存失败 [order_id=%d, pack_id=%d]: %v", order.ID, packID, err)
				return
			}
		}
	}

	// 退还余额支付部分
	if order.BalanceAmount > 0 {
		balance, err := s.userRepo.GetBalanceForUpdateTx(tx, order.UserID)
		if err != nil {
			log.Printf("查询余额失败（资源包超时退款）[order_id=%d]: %v", order.ID, err)
			return
		}
		newBalance := balance + order.BalanceAmount
		if err := s.userRepo.UpdateUserBalanceTx(tx, order.UserID, newBalance); err != nil {
			log.Printf("退还余额失败（资源包超时）[order_id=%d]: %v", order.ID, err)
			return
		}
		logEntry := &model.BalanceLog{
			UserID:        order.UserID,
			OrderID:       order.ID,
			Type:          3, // 退款
			Amount:        order.BalanceAmount,
			BalanceBefore: balance,
			BalanceAfter:  newBalance,
			Remark:        fmt.Sprintf("购买资源包订单超时退款：%s", order.BizNo),
		}
		if err := s.balanceLogRepo.CreateLogTx(tx, logEntry); err != nil {
			log.Printf("写退款流水失败（资源包超时）[order_id=%d]: %v", order.ID, err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("释放资源包订单提交失败 [order_id=%d]: %v", order.ID, err)
	}
}

// boolToInt 布尔转 0/1
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
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
