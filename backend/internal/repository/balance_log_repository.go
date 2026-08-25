package repository

import (
	"database/sql"
	"time"

	"starloftrpa/internal/model"
)

type BalanceLogRepository struct {
	db *sql.DB
}

func NewBalanceLogRepository(db *sql.DB) *BalanceLogRepository {
	return &BalanceLogRepository{db: db}
}

// CreateLog 创建余额流水记录
func (r *BalanceLogRepository) CreateLog(log *model.BalanceLog) error {
	query := `INSERT INTO balance_log 
		(user_id, order_id, type, amount, balance_before, balance_after, bank_serial_no, remark, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		log.UserID,
		log.OrderID,
		log.Type,
		log.Amount,
		log.BalanceBefore,
		log.BalanceAfter,
		log.BankSerialNo,
		log.Remark,
		time.Now(),
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	log.ID = id
	return nil
}

// CreateLogTx 在事务中创建余额流水记录
func (r *BalanceLogRepository) CreateLogTx(tx *sql.Tx, log *model.BalanceLog) error {
	query := `INSERT INTO balance_log 
		(user_id, order_id, type, amount, balance_before, balance_after, bank_serial_no, remark, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := tx.Exec(query,
		log.UserID,
		log.OrderID,
		log.Type,
		log.Amount,
		log.BalanceBefore,
		log.BalanceAfter,
		log.BankSerialNo,
		log.Remark,
		time.Now(),
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	log.ID = id
	return nil
}

// GetUserBalanceLogs 查询用户余额流水（分页）
func (r *BalanceLogRepository) GetUserBalanceLogs(userID int64, page, pageSize int) ([]*model.BalanceLog, int64, error) {
	// 查询总数
	countQuery := `SELECT COUNT(*) FROM balance_log WHERE user_id = ?`
	var total int64
	err := r.db.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询流水列表
	offset := (page - 1) * pageSize
	query := `SELECT id, user_id, order_id, type, amount, balance_before, balance_after, COALESCE(bank_serial_no, ''), remark, created_at 
		FROM balance_log WHERE user_id = ? 
		ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	logs := make([]*model.BalanceLog, 0)
	for rows.Next() {
		log := &model.BalanceLog{}
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.OrderID,
			&log.Type,
			&log.Amount,
			&log.BalanceBefore,
			&log.BalanceAfter,
			&log.BankSerialNo,
			&log.Remark,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}

// GetUserFinanceStats 获取用户财务统计
func (r *BalanceLogRepository) GetUserFinanceStats(userID int64) (map[string]interface{}, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN type = 1 THEN amount ELSE 0 END), 0) as total_recharge,
			COALESCE(SUM(CASE WHEN type = 2 THEN amount ELSE 0 END), 0) as total_consume,
			COALESCE(SUM(CASE WHEN type = 3 THEN amount ELSE 0 END), 0) as total_refund
		FROM balance_log 
		WHERE user_id = ?`

	var totalRecharge, totalConsume, totalRefund float64
	err := r.db.QueryRow(query, userID).Scan(&totalRecharge, &totalConsume, &totalRefund)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"totalRecharge": totalRecharge,
		"totalConsume":  totalConsume,
		"totalRefund":   totalRefund,
	}, nil
}
