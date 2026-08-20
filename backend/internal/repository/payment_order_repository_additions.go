package repository

import (
	"database/sql"
	"starloftrpa/internal/model"
)

// GetOrderByPayOrderNoForUpdate 在事务中锁定并查询支付订单
func (r *PaymentOrderRepository) GetOrderByPayOrderNoForUpdate(tx *sql.Tx, payOrderNo string) (*model.PaymentOrder, error) {
	query := `SELECT id, pay_order_no, user_id, amount, channel, COALESCE(channel_trade_no, ''), status, 
		refund_status, COALESCE(refund_amount, 0), expire_time, paid_at, refunded_at, created_at, updated_at 
		FROM payment_order WHERE pay_order_no = ? FOR UPDATE`

	order := &model.PaymentOrder{}
	err := tx.QueryRow(query, payOrderNo).Scan(
		&order.ID,
		&order.PayOrderNo,
		&order.UserID,
		&order.Amount,
		&order.Channel,
		&order.ChannelTradeNo,
		&order.Status,
		&order.RefundStatus,
		&order.RefundAmount,
		&order.ExpireTime,
		&order.PaidAt,
		&order.RefundedAt,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrPaymentOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return order, nil
}

// GetOrderConsumeAmountTx 在事务中查询订单消费金额
func (r *PaymentOrderRepository) GetOrderConsumeAmountTx(tx *sql.Tx, orderID int64) (float64, error) {
	var consumeAmount float64
	query := `SELECT COALESCE(SUM(amount), 0) FROM balance_log 
		WHERE order_id = ? AND type = 2`
	err := tx.QueryRow(query, orderID).Scan(&consumeAmount)
	if err != nil {
		return 0, err
	}
	return consumeAmount, nil
}
