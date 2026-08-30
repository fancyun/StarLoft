package repository

import (
	"database/sql"
	"time"

	"starloftrpa/internal/model"
)

type KycRecordRepository struct {
	db *sql.DB
}

func NewKycRecordRepository(db *sql.DB) *KycRecordRepository {
	return &KycRecordRepository{db: db}
}

// Create 创建实名认证记录
func (r *KycRecordRepository) Create(record *model.KycRecord) error {
	query := `INSERT INTO kyc_record 
		(user_id, auth_order_id, name, id_card, status, result_code, 
		result_message, result_data, verified_at, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		record.UserID,
		record.AuthOrderID,
		record.Name,
		record.IDCard,
		record.Status,
		record.ResultCode,
		record.ResultMessage,
		record.ResultData,
		record.VerifiedAt,
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
	record.ID = id
	return nil
}

// GetLatestByUserID 获取用户最近一次认证记录
func (r *KycRecordRepository) GetLatestByUserID(userID int64) (*model.KycRecord, error) {
	query := `SELECT id, user_id, auth_order_id, name, id_card, status, 
		result_code, result_message, result_data, verified_at, created_at, updated_at 
		FROM kyc_record WHERE user_id = ? 
		ORDER BY created_at DESC LIMIT 1`

	record := &model.KycRecord{}
	err := r.db.QueryRow(query, userID).Scan(
		&record.ID,
		&record.UserID,
		&record.AuthOrderID,
		&record.Name,
		&record.IDCard,
		&record.Status,
		&record.ResultCode,
		&record.ResultMessage,
		&record.ResultData,
		&record.VerifiedAt,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return record, nil
}

// UpdateResult 更新认证结果
func (r *KycRecordRepository) UpdateResult(id int64, status int, resultCode, resultMessage, resultData string, verifiedAt *time.Time) error {
	query := `UPDATE kyc_record 
			SET status = ?, result_code = ?, result_message = ?, result_data = ?, 
			verified_at = ?, updated_at = ? 
			WHERE id = ?`
	_, err := r.db.Exec(query, status, resultCode, resultMessage, resultData, verifiedAt, time.Now(), id)
	return err
}

// Cancel 取消认证记录（将状态设为 3-认证失败/取消）
func (r *KycRecordRepository) Cancel(id int64) error {
	query := `UPDATE kyc_record SET status = 3, result_message = '用户取消认证', updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, time.Now(), id)
	return err
}
