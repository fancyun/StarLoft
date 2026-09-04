package repository

import (
	"database/sql"
	"time"

	"starloftrpa/internal/model"
)

type AuthOrderRepository struct {
	db *sql.DB
}

func NewAuthOrderRepository(db *sql.DB) *AuthOrderRepository {
	return &AuthOrderRepository{db: db}
}

// CreateOrder 创建认证订单
func (r *AuthOrderRepository) CreateOrder(order *model.AuthOrder) error {
	query := `INSERT INTO ` + model.KycDB + `.auth_order
		(platform_biz_no, biz_no, user_id, return_url, notify_url, 
		biz_extra_data, status, cost, source, pay_type, user_pack_id, is_refunded, notify_times, 
			notify_status, created_at, updated_at) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		order.PlatformBizNo,
		order.BizNo,
		order.UserID,
		order.ReturnURL,
		order.NotifyURL,
		order.BizExtraData,
		order.Status,
		order.Cost,
		order.Source,
		order.PayType,
		order.UserPackID,
		order.IsRefunded,
		order.NotifyTimes,
		order.NotifyStatus,
		time.Now(),
		time.Now(),
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

// GetOrderByPlatformBizNo 根据平台流水号查询订单
func (r *AuthOrderRepository) GetOrderByPlatformBizNo(platformBizNo string) (*model.AuthOrder, error) {
	query := `SELECT id, platform_biz_no, COALESCE(biz_no, ''), user_id, 
		COALESCE(return_url, ''), COALESCE(notify_url, ''), COALESCE(biz_extra_data, ''), 
		COALESCE(up_token, ''), COALESCE(up_biz_id, ''), COALESCE(up_request_id, ''), 
		COALESCE(result_code, ''), COALESCE(result_message, ''),
		status, cost, source, pay_type, COALESCE(user_pack_id, 0), is_refunded, notify_times, 
		notify_status, created_at, updated_at, finished_at 
		FROM ` + model.KycDB + `.auth_order WHERE platform_biz_no = ?`

	order := &model.AuthOrder{}
	err := r.db.QueryRow(query, platformBizNo).Scan(
		&order.ID,
		&order.PlatformBizNo,
		&order.BizNo,
		&order.UserID,
		&order.ReturnURL,
		&order.NotifyURL,
		&order.BizExtraData,
		&order.UpToken,
		&order.UpBizID,
		&order.UpRequestID,
		&order.ResultCode,
		&order.ResultMessage,
		&order.Status,
		&order.Cost,
		&order.Source,
		&order.PayType,
		&order.UserPackID,
		&order.IsRefunded,
		&order.NotifyTimes,
		&order.NotifyStatus,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.FinishedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return order, nil
}

// GetOrderByID 根据订单ID查询订单
func (r *AuthOrderRepository) GetOrderByID(orderID int64) (*model.AuthOrder, error) {
	query := `SELECT id, platform_biz_no, COALESCE(biz_no, ''), user_id, 
		COALESCE(return_url, ''), COALESCE(notify_url, ''), COALESCE(biz_extra_data, ''), 
		COALESCE(up_token, ''), COALESCE(up_biz_id, ''), COALESCE(up_request_id, ''), 
		COALESCE(result_code, ''), COALESCE(result_message, ''), COALESCE(result_data, ''),
		status, cost, source, pay_type, COALESCE(user_pack_id, 0), is_refunded, notify_times, 
		notify_status, created_at, updated_at, finished_at 
		FROM ` + model.KycDB + `.auth_order WHERE id = ?`

	order := &model.AuthOrder{}
	err := r.db.QueryRow(query, orderID).Scan(
		&order.ID,
		&order.PlatformBizNo,
		&order.BizNo,
		&order.UserID,
		&order.ReturnURL,
		&order.NotifyURL,
		&order.BizExtraData,
		&order.UpToken,
		&order.UpBizID,
		&order.UpRequestID,
		&order.ResultCode,
		&order.ResultMessage,
		&order.ResultData,
		&order.Status,
		&order.Cost,
		&order.Source,
		&order.PayType,
		&order.UserPackID,
		&order.IsRefunded,
		&order.NotifyTimes,
		&order.NotifyStatus,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.FinishedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return order, nil
}

// RecentAuthOrder 最近认证订单（含用户手机号和姓名）
type RecentAuthOrder struct {
	PlatformBizNo string
	UserPhone     string
	Name          string
	Status        int
	Cost          float64
	CreatedAt     time.Time
}

// GetRecentOrders 获取最近认证订单列表（含用户手机号和姓名）
func (r *AuthOrderRepository) GetRecentOrders(limit int) ([]*RecentAuthOrder, error) {
	query := `SELECT ao.platform_biz_no, u.phone, 
		COALESCE((SELECT kr.name FROM ` + model.SysDB + `.kyc_record kr WHERE kr.user_id = ao.user_id ORDER BY kr.id DESC LIMIT 1), ''), 
		ao.status, ao.cost, ao.created_at
		FROM ` + model.KycDB + `.auth_order ao
		JOIN ` + model.SysDB + `.platform_user u ON u.id = ao.user_id
		ORDER BY ao.created_at DESC LIMIT ?`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]*RecentAuthOrder, 0)
	for rows.Next() {
		o := &RecentAuthOrder{}
		err := rows.Scan(&o.PlatformBizNo, &o.UserPhone, &o.Name, &o.Status, &o.Cost, &o.CreatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

// GetDailyOrderStats 按天统计认证订单数
func (r *AuthOrderRepository) GetDailyOrderStats(startDate, endDate string) (map[string]int64, error) {
	query := `SELECT DATE_FORMAT(created_at, '%Y-%m-%d') AS d, COUNT(*) AS c
		FROM ` + model.KycDB + `.auth_order
		WHERE DATE(created_at) >= ? AND DATE(created_at) <= ?
		GROUP BY DATE_FORMAT(created_at, '%Y-%m-%d')`

	rows, err := r.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int64)
	for rows.Next() {
		var d string
		var c int64
		if err := rows.Scan(&d, &c); err != nil {
			return nil, err
		}
		result[d] = c
	}
	return result, nil
}

// GetDailyIncomeStats 按天统计认证收入（仅统计已完成的认证订单）
func (r *AuthOrderRepository) GetDailyIncomeStats(startDate, endDate string) (map[string]float64, error) {
	query := `SELECT DATE_FORMAT(created_at, '%Y-%m-%d') AS d, COALESCE(SUM(cost), 0) AS amount
		FROM ` + model.KycDB + `.auth_order
		WHERE status = 2 AND DATE(created_at) >= ? AND DATE(created_at) <= ?
		GROUP BY DATE_FORMAT(created_at, '%Y-%m-%d')`

	rows, err := r.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var d string
		var amount float64
		if err := rows.Scan(&d, &amount); err != nil {
			return nil, err
		}
		result[d] = amount
	}
	return result, nil
}

// GetOrderByBizNo 根据用户业务流水号查询订单
func (r *AuthOrderRepository) GetOrderByBizNo(userID int64, bizNo string) (*model.AuthOrder, error) {
	query := `SELECT id, platform_biz_no, COALESCE(biz_no, ''), user_id, 
		COALESCE(return_url, ''), COALESCE(notify_url, ''), COALESCE(biz_extra_data, ''), 
		COALESCE(up_token, ''), COALESCE(up_biz_id, ''), COALESCE(up_request_id, ''), 
		COALESCE(result_code, ''), COALESCE(result_message, ''),
		status, cost, source, pay_type, COALESCE(user_pack_id, 0), is_refunded, notify_times, 
		notify_status, created_at, updated_at, finished_at 
		FROM ` + model.KycDB + `.auth_order WHERE user_id = ? AND biz_no = ? ORDER BY id DESC LIMIT 1`

	order := &model.AuthOrder{}
	err := r.db.QueryRow(query, userID, bizNo).Scan(
		&order.ID,
		&order.PlatformBizNo,
		&order.BizNo,
		&order.UserID,
		&order.ReturnURL,
		&order.NotifyURL,
		&order.BizExtraData,
		&order.UpToken,
		&order.UpBizID,
		&order.UpRequestID,
		&order.ResultCode,
		&order.ResultMessage,
		&order.Status,
		&order.Cost,
		&order.Source,
		&order.PayType,
		&order.UserPackID,
		&order.IsRefunded,
		&order.NotifyTimes,
		&order.NotifyStatus,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.FinishedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return order, nil
}

// UpdateOrderUpstreamInfo 更新订单上游信息
func (r *AuthOrderRepository) UpdateOrderUpstreamInfo(orderID int64, token, bizID, requestID string) error {
	query := `UPDATE ` + model.KycDB + `.auth_order 
		SET up_token = ?, up_biz_id = ?, up_request_id = ?, status = 1, updated_at = ? 
		WHERE id = ?`
	_, err := r.db.Exec(query, token, bizID, requestID, time.Now(), orderID)
	return err
}

// GetOrderByUpBizID 根据上游业务ID查询订单
func (r *AuthOrderRepository) GetOrderByUpBizID(upBizID string) (*model.AuthOrder, error) {
	query := `SELECT id, platform_biz_no, COALESCE(biz_no, ''), user_id, 
			COALESCE(return_url, ''), COALESCE(notify_url, ''), COALESCE(biz_extra_data, ''), 
			COALESCE(up_token, ''), COALESCE(up_biz_id, ''), COALESCE(up_request_id, ''), 
			COALESCE(result_code, ''), COALESCE(result_message, ''),
			status, cost, source, pay_type, COALESCE(user_pack_id, 0), is_refunded, notify_times, 
			notify_status, created_at, updated_at, finished_at 
			FROM ` + model.KycDB + `.auth_order WHERE up_biz_id = ?`

	order := &model.AuthOrder{}
	err := r.db.QueryRow(query, upBizID).Scan(
		&order.ID,
		&order.PlatformBizNo,
		&order.BizNo,
		&order.UserID,
		&order.ReturnURL,
		&order.NotifyURL,
		&order.BizExtraData,
		&order.UpToken,
		&order.UpBizID,
		&order.UpRequestID,
		&order.ResultCode,
		&order.ResultMessage,
		&order.Status,
		&order.Cost,
		&order.Source,
		&order.PayType,
		&order.UserPackID,
		&order.IsRefunded,
		&order.NotifyTimes,
		&order.NotifyStatus,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.FinishedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return order, nil
}

// UpdateOrderPayType 更新订单扣费方式（资源包并发耗尽时回退到余额扣费使用）
func (r *AuthOrderRepository) UpdateOrderPayType(orderID int64, payType int, userPackID int64) error {
	query := `UPDATE ` + model.KycDB + `.auth_order SET pay_type = ?, user_pack_id = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, payType, userPackID, time.Now(), orderID)
	return err
}

// UpdateOrderResult 更新订单认证结果
func (r *AuthOrderRepository) UpdateOrderResult(orderID int64, resultCode, resultMessage string, status int) error {
	query := `UPDATE ` + model.KycDB + `.auth_order 
		SET result_code = ?, result_message = ?, status = ?, finished_at = ?, updated_at = ? 
		WHERE id = ?`
	_, err := r.db.Exec(query, resultCode, resultMessage, status, time.Now(), time.Now(), orderID)
	return err
}

// UpdateOrderRefundFlag 仅标记订单已退款（不改变状态）
func (r *AuthOrderRepository) UpdateOrderRefundFlag(orderID int64) error {
	query := `UPDATE ` + model.KycDB + `.auth_order SET is_refunded = 1, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, time.Now(), orderID)
	return err
}

// GetPendingOrders 查询所有处理中且未退款的订单（供定时任务主动同步上游结果）
func (r *AuthOrderRepository) GetPendingOrders() ([]*model.AuthOrder, error) {
	query := `SELECT id, platform_biz_no, COALESCE(biz_no, ''), user_id, 
		COALESCE(return_url, ''), COALESCE(notify_url, ''), COALESCE(biz_extra_data, ''), 
		COALESCE(up_token, ''), COALESCE(up_biz_id, ''), COALESCE(up_request_id, ''), 
		COALESCE(result_code, ''), COALESCE(result_message, ''),
		status, cost, source, pay_type, COALESCE(user_pack_id, 0), is_refunded, notify_times, 
		notify_status, created_at, updated_at, finished_at 
		FROM ` + model.KycDB + `.auth_order WHERE status IN (0, 1) AND is_refunded = 0 AND up_biz_id IS NOT NULL AND up_biz_id != ''`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]*model.AuthOrder, 0)
	for rows.Next() {
		order := &model.AuthOrder{}
		err := rows.Scan(
			&order.ID,
			&order.PlatformBizNo,
			&order.BizNo,
			&order.UserID,
			&order.ReturnURL,
			&order.NotifyURL,
			&order.BizExtraData,
			&order.UpToken,
			&order.UpBizID,
			&order.UpRequestID,
			&order.ResultCode,
			&order.ResultMessage,
			&order.Status,
			&order.Cost,
			&order.Source,
			&order.PayType,
			&order.UserPackID,
			&order.IsRefunded,
			&order.NotifyTimes,
			&order.NotifyStatus,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.FinishedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// GetLatestPendingOrder 获取用户最新进行中的订单（用于继续认证）
func (r *AuthOrderRepository) GetLatestPendingOrder(userID int64) (*model.AuthOrder, error) {
	query := `SELECT id, platform_biz_no, COALESCE(biz_no, ''), user_id, 
			COALESCE(return_url, ''), COALESCE(notify_url, ''), COALESCE(biz_extra_data, ''), 
			COALESCE(up_token, ''), COALESCE(up_biz_id, ''), COALESCE(up_request_id, ''), 
			COALESCE(result_code, ''), COALESCE(result_message, ''), 
			status, cost, source, pay_type, COALESCE(user_pack_id, 0), is_refunded, notify_times, 
			notify_status, created_at, updated_at, finished_at 
			FROM ` + model.KycDB + `.auth_order WHERE user_id = ? AND status IN (0, 1) 
			ORDER BY created_at DESC LIMIT 1`

	order := &model.AuthOrder{}
	err := r.db.QueryRow(query, userID).Scan(
		&order.ID,
		&order.PlatformBizNo,
		&order.BizNo,
		&order.UserID,
		&order.ReturnURL,
		&order.NotifyURL,
		&order.BizExtraData,
		&order.UpToken,
		&order.UpBizID,
		&order.UpRequestID,
		&order.ResultCode,
		&order.ResultMessage,
		&order.Status,
		&order.Cost,
		&order.Source,
		&order.PayType,
		&order.UserPackID,
		&order.IsRefunded,
		&order.NotifyTimes,
		&order.NotifyStatus,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.FinishedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return order, nil
}

// GetAllOrders 获取所有认证订单列表（管理员，带分页和筛选，含用户手机号）
func (r *AuthOrderRepository) GetAllOrders(page, pageSize int, status *int, userID *int64) ([]*model.AuthOrder, int64, error) {
	offset := (page - 1) * pageSize

	// 构建查询条件（带 ao. 前缀，避免与 platform_user 联表后的列名歧义）
	whereClause := ""
	args := []interface{}{}

	if status != nil {
		whereClause = "WHERE ao.status = ?"
		args = append(args, *status)
	}

	if userID != nil {
		if whereClause == "" {
			whereClause = "WHERE ao.user_id = ?"
		} else {
			whereClause += " AND ao.user_id = ?"
		}
		args = append(args, *userID)
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM ` + model.KycDB + `.auth_order ao JOIN ` + model.SysDB + `.platform_user u ON u.id = ao.user_id " + whereClause
	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表
	query := `SELECT ao.id, ao.platform_biz_no, COALESCE(ao.biz_no, ''), ao.user_id, 
		COALESCE(ao.return_url, ''), COALESCE(ao.notify_url, ''), COALESCE(ao.biz_extra_data, ''), 
		COALESCE(ao.up_token, ''), COALESCE(ao.up_biz_id, ''), COALESCE(ao.up_request_id, ''), 
		COALESCE(ao.result_code, ''), COALESCE(ao.result_message, ''),
		ao.status, ao.cost, ao.is_refunded, ao.notify_times, 
		ao.notify_status, ao.created_at, ao.updated_at, ao.finished_at, u.phone 
		FROM ` + model.KycDB + `.auth_order ao 
		JOIN ` + model.SysDB + `.platform_user u ON u.id = ao.user_id
		` + whereClause + ` ORDER BY ao.created_at DESC LIMIT ? OFFSET ?`

	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders := make([]*model.AuthOrder, 0)
	for rows.Next() {
		order := &model.AuthOrder{}
		err := rows.Scan(
			&order.ID,
			&order.PlatformBizNo,
			&order.BizNo,
			&order.UserID,
			&order.ReturnURL,
			&order.NotifyURL,
			&order.BizExtraData,
			&order.UpToken,
			&order.UpBizID,
			&order.UpRequestID,
			&order.ResultCode,
			&order.ResultMessage,
			&order.Status,
			&order.Cost,
			&order.IsRefunded,
			&order.NotifyTimes,
			&order.NotifyStatus,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.FinishedAt,
			&order.UserPhone,
		)
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, order)
	}

	return orders, total, nil
}

// GetUserAuthOrders 获取指定用户的认证订单列表
func (r *AuthOrderRepository) GetUserAuthOrders(userID int64, page, pageSize int) ([]*model.AuthOrder, int64, error) {
	// 查询总数
	countQuery := `SELECT COUNT(*) FROM ` + model.KycDB + `.auth_order WHERE user_id = ?`
	var total int64
	err := r.db.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询订单列表
	offset := (page - 1) * pageSize
	query := `SELECT 
		id, platform_biz_no, COALESCE(biz_no, ''), user_id, 
		COALESCE(return_url, ''), COALESCE(notify_url, ''), COALESCE(biz_extra_data, ''), 
		COALESCE(up_token, ''), COALESCE(up_biz_id, ''), COALESCE(up_request_id, ''), 
		COALESCE(result_code, ''), COALESCE(result_message, ''), status, cost, source, pay_type, COALESCE(user_pack_id, 0), is_refunded, 
		notify_times, notify_status, created_at, updated_at, finished_at
		FROM ` + model.KycDB + `.auth_order 
		WHERE user_id = ?
		ORDER BY created_at DESC 
		LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders := make([]*model.AuthOrder, 0)
	for rows.Next() {
		order := &model.AuthOrder{}
		err := rows.Scan(
			&order.ID,
			&order.PlatformBizNo,
			&order.BizNo,
			&order.UserID,
			&order.ReturnURL,
			&order.NotifyURL,
			&order.BizExtraData,
			&order.UpToken,
			&order.UpBizID,
			&order.UpRequestID,
			&order.ResultCode,
			&order.ResultMessage,
			&order.Status,
			&order.Cost,
			&order.Source,
			&order.PayType,
			&order.UserPackID,
			&order.IsRefunded,
			&order.NotifyTimes,
			&order.NotifyStatus,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.FinishedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, order)
	}

	return orders, total, nil
}
