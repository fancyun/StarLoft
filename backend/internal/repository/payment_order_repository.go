package repository

import (
	"database/sql"
	"errors"
	"starloftrpa/internal/model"
	"time"
)

var (
	ErrPaymentOrderNotFound = errors.New("payment order not found")
)

type PaymentOrderRepository struct {
	db *sql.DB
}

func NewPaymentOrderRepository(db *sql.DB) *PaymentOrderRepository {
	return &PaymentOrderRepository{db: db}
}

// CreateOrder 创建支付订单
func (r *PaymentOrderRepository) CreateOrder(order *model.PaymentOrder) error {
	query := `INSERT INTO ` + model.SysDB + `.payment_order 
		(pay_order_no, user_id, amount, channel, status, 
		expire_time, created_at, updated_at, refund_status, intent, biz_no, balance_amount, stock_reserved) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		order.PayOrderNo,
		order.UserID,
		order.Amount,
		order.Channel,
		order.Status,
		order.ExpireTime,
		time.Now(),
		time.Now(),
		order.RefundStatus,
		order.Intent,
		order.BizNo,
		order.BalanceAmount,
		order.StockReserved,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	order.ID = id
	return nil
}

// CreateOrderTx 在事务中创建支付订单（用于购买资源包组合支付的原子落库）
func (r *PaymentOrderRepository) CreateOrderTx(tx *sql.Tx, order *model.PaymentOrder) error {
	query := `INSERT INTO ` + model.SysDB + `.payment_order 
		(pay_order_no, user_id, amount, channel, status, 
		expire_time, created_at, updated_at, refund_status, intent, biz_no, balance_amount, stock_reserved) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := tx.Exec(query,
		order.PayOrderNo,
		order.UserID,
		order.Amount,
		order.Channel,
		order.Status,
		order.ExpireTime,
		time.Now(),
		time.Now(),
		order.RefundStatus,
		order.Intent,
		order.BizNo,
		order.BalanceAmount,
		order.StockReserved,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	order.ID = id
	return nil
}

// GetOrderByID 根据ID查询支付订单
func (r *PaymentOrderRepository) GetOrderByID(orderID int64) (*model.PaymentOrder, error) {
	query := `SELECT id, pay_order_no, user_id, amount, 
		channel, COALESCE(channel_trade_no, ''), status, expire_time, paid_at, created_at, updated_at, 
		refund_status, COALESCE(refund_amount, 0), refunded_at, intent, COALESCE(biz_no, ''), 
		COALESCE(balance_amount, 0), COALESCE(stock_reserved, 0) 
		FROM ` + model.SysDB + `.payment_order WHERE id = ?`

	order := &model.PaymentOrder{}
	err := r.db.QueryRow(query, orderID).Scan(
		&order.ID,
		&order.PayOrderNo,
		&order.UserID,
		&order.Amount,
		&order.Channel,
		&order.ChannelTradeNo,
		&order.Status,
		&order.ExpireTime,
		&order.PaidAt,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.RefundStatus,
		&order.RefundAmount,
		&order.RefundedAt,
		&order.Intent,
		&order.BizNo,
		&order.BalanceAmount,
		&order.StockReserved,
	)
	if err == sql.ErrNoRows {
		return nil, ErrPaymentOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return order, nil
}

// GetOrderByPayOrderNo 根据支付流水号查询订单
func (r *PaymentOrderRepository) GetOrderByPayOrderNo(payOrderNo string) (*model.PaymentOrder, error) {
	query := `SELECT id, pay_order_no, user_id, amount, 
		channel, COALESCE(channel_trade_no, ''), status, expire_time, paid_at, created_at, updated_at, 
		refund_status, COALESCE(refund_amount, 0), refunded_at, intent, COALESCE(biz_no, ''), 
		COALESCE(balance_amount, 0), COALESCE(stock_reserved, 0) 
		FROM ` + model.SysDB + `.payment_order WHERE pay_order_no = ?`

	order := &model.PaymentOrder{}
	err := r.db.QueryRow(query, payOrderNo).Scan(
		&order.ID,
		&order.PayOrderNo,
		&order.UserID,
		&order.Amount,
		&order.Channel,
		&order.ChannelTradeNo,
		&order.Status,
		&order.ExpireTime,
		&order.PaidAt,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.RefundStatus,
		&order.RefundAmount,
		&order.RefundedAt,
		&order.Intent,
		&order.BizNo,
		&order.BalanceAmount,
		&order.StockReserved,
	)
	if err == sql.ErrNoRows {
		return nil, ErrPaymentOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return order, nil
}

// MarkOrderPaidIfPending 仅当订单仍为待支付时更新为已支付（幂等）
// 返回是否发生了状态变更
func (r *PaymentOrderRepository) MarkOrderPaidIfPending(orderID int64, channelTradeNo string) (bool, error) {
	query := `UPDATE ` + model.SysDB + `.payment_order 
		SET status = 1, channel_trade_no = ?, paid_at = ?, updated_at = ? 
		WHERE id = ? AND status = 0`
	result, err := r.db.Exec(query, channelTradeNo, time.Now(), time.Now(), orderID)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// MarkOrderPaidIfPendingTx 在事务中标记支付单为已支付（幂等，供资源包等组合支付落地使用）
func (r *PaymentOrderRepository) MarkOrderPaidIfPendingTx(tx *sql.Tx, orderID int64, channelTradeNo string) (bool, error) {
	query := `UPDATE ` + model.SysDB + `.payment_order 
		SET status = 1, channel_trade_no = ?, paid_at = ?, updated_at = ? 
		WHERE id = ? AND status = 0`
	result, err := tx.Exec(query, channelTradeNo, time.Now(), time.Now(), orderID)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// CloseOrderIfPendingTx 在事务中关闭待支付订单（状态 0→3，幂等，供资源包超时释放库存使用）
func (r *PaymentOrderRepository) CloseOrderIfPendingTx(tx *sql.Tx, orderID int64) (bool, error) {
	query := `UPDATE ` + model.SysDB + `.payment_order 
		SET status = 3, updated_at = ? 
		WHERE id = ? AND status = 0`
	result, err := tx.Exec(query, time.Now(), orderID)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// CloseOrderIfPending 非事务关闭待支付订单（状态 0→3，幂等，供对账单边关闭使用）
func (r *PaymentOrderRepository) CloseOrderIfPending(orderID int64) (bool, error) {
	query := `UPDATE ` + model.SysDB + `.payment_order 
		SET status = 3, updated_at = ? 
		WHERE id = ? AND status = 0`
	result, err := r.db.Exec(query, time.Now(), orderID)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// GetPendingOrdersForReconcile 查询早于指定时间创建、仍待支付的支付订单（用于每日对账）
// olderThan 用于排除刚创建仍在正常支付流程中的订单
func (r *PaymentOrderRepository) GetPendingOrdersForReconcile(olderThan time.Time) ([]*model.PaymentOrder, error) {
	query := `SELECT id, pay_order_no, user_id, amount, 
		channel, COALESCE(channel_trade_no, ''), status, expire_time, paid_at, created_at, updated_at, 
		refund_status, COALESCE(refund_amount, 0), refunded_at, intent, COALESCE(biz_no, ''), 
		COALESCE(balance_amount, 0), COALESCE(stock_reserved, 0) 
		FROM ` + model.SysDB + `.payment_order 
		WHERE status = 0 AND created_at < ?`

	rows, err := r.db.Query(query, olderThan)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]*model.PaymentOrder, 0)
	for rows.Next() {
		order := &model.PaymentOrder{}
		if err := rows.Scan(
			&order.ID,
			&order.PayOrderNo,
			&order.UserID,
			&order.Amount,
			&order.Channel,
			&order.ChannelTradeNo,
			&order.Status,
			&order.ExpireTime,
			&order.PaidAt,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.RefundStatus,
			&order.RefundAmount,
			&order.RefundedAt,
			&order.Intent,
			&order.BizNo,
			&order.BalanceAmount,
			&order.StockReserved,
		); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

// GetExpiredPendingResourcePackOrders 查询已过期且仍待支付的资源包支付订单（用于定时释放库存）
func (r *PaymentOrderRepository) GetExpiredPendingResourcePackOrders(now time.Time) ([]*model.PaymentOrder, error) {
	query := `SELECT id, pay_order_no, user_id, amount, 
		channel, COALESCE(channel_trade_no, ''), status, expire_time, paid_at, created_at, updated_at, 
		refund_status, COALESCE(refund_amount, 0), refunded_at, intent, COALESCE(biz_no, ''), 
		COALESCE(balance_amount, 0), COALESCE(stock_reserved, 0) 
		FROM ` + model.SysDB + `.payment_order 
		WHERE intent = 'resource_pack' AND status = 0 AND expire_time IS NOT NULL AND expire_time < ?`

	rows, err := r.db.Query(query, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]*model.PaymentOrder, 0)
	for rows.Next() {
		order := &model.PaymentOrder{}
		if err := rows.Scan(
			&order.ID,
			&order.PayOrderNo,
			&order.UserID,
			&order.Amount,
			&order.Channel,
			&order.ChannelTradeNo,
			&order.Status,
			&order.ExpireTime,
			&order.PaidAt,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.RefundStatus,
			&order.RefundAmount,
			&order.RefundedAt,
			&order.Intent,
			&order.BizNo,
			&order.BalanceAmount,
			&order.StockReserved,
		); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

// GetAllOrders 获取所有支付订单列表（管理员，带分页和筛选）
func (r *PaymentOrderRepository) GetAllOrders(page, pageSize int, status *int, userID *int64) ([]*model.PaymentOrder, int64, error) {
	offset := (page - 1) * pageSize

	// 构建查询条件
	whereClause := ""
	args := []interface{}{}

	if status != nil {
		whereClause = "WHERE status = ?"
		args = append(args, *status)
	}

	if userID != nil {
		if whereClause == "" {
			whereClause = "WHERE user_id = ?"
		} else {
			whereClause += " AND user_id = ?"
		}
		args = append(args, *userID)
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM ` + model.SysDB + `.payment_order " + whereClause
	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表
	query := `SELECT id, pay_order_no, user_id, amount, 
		channel, COALESCE(channel_trade_no, ''), status, expire_time, paid_at, created_at, updated_at, 
		refund_status, COALESCE(refund_amount, 0), refunded_at, intent, COALESCE(biz_no, ''), 
		COALESCE(balance_amount, 0), COALESCE(stock_reserved, 0) 
		FROM ` + model.SysDB + `.payment_order ` + whereClause + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`

	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders := make([]*model.PaymentOrder, 0)
	for rows.Next() {
		order := &model.PaymentOrder{}
		err := rows.Scan(
			&order.ID,
			&order.PayOrderNo,
			&order.UserID,
			&order.Amount,
			&order.Channel,
			&order.ChannelTradeNo,
			&order.Status,
			&order.ExpireTime,
			&order.PaidAt,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.RefundStatus,
			&order.RefundAmount,
			&order.RefundedAt,
			&order.Intent,
			&order.BizNo,
			&order.BalanceAmount,
			&order.StockReserved,
		)
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, order)
	}

	return orders, total, nil
}
