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

// kycRecordColumns 实名认证记录常用查询列
const kycRecordColumns = `id, user_id, source, auth_order_id, COALESCE(platform_biz_no, ''), 
	COALESCE(biz_no, ''), COALESCE(return_url, ''), COALESCE(notify_url, ''), COALESCE(biz_extra_data, ''), 
	COALESCE(up_token, ''), COALESCE(up_biz_id, ''), COALESCE(up_request_id, ''), 
	name, id_card, status, COALESCE(result_code, ''), COALESCE(result_message, ''), COALESCE(result_data, ''), 
	verified_at, created_at, updated_at`

// scanKycRecord 将查询结果扫描到 KycRecord
func scanKycRecord(row interface{ Scan(...interface{}) error }) (*model.KycRecord, error) {
	record := &model.KycRecord{}
	err := row.Scan(
		&record.ID,
		&record.UserID,
		&record.Source,
		&record.AuthOrderID,
		&record.PlatformBizNo,
		&record.BizNo,
		&record.ReturnURL,
		&record.NotifyURL,
		&record.BizExtraData,
		&record.UpToken,
		&record.UpBizID,
		&record.UpRequestID,
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
	if err != nil {
		return nil, err
	}
	return record, nil
}

// Create 创建实名认证记录
func (r *KycRecordRepository) Create(record *model.KycRecord) error {
	query := `INSERT INTO ` + model.SysDB + `.kyc_record 
		(user_id, source, auth_order_id, platform_biz_no, biz_no, return_url, notify_url, 
		biz_extra_data, up_token, up_biz_id, up_request_id, name, id_card, status, result_code, 
		result_message, result_data, verified_at, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		record.UserID,
		record.Source,
		record.AuthOrderID,
		record.PlatformBizNo,
		record.BizNo,
		record.ReturnURL,
		record.NotifyURL,
		record.BizExtraData,
		record.UpToken,
		record.UpBizID,
		record.UpRequestID,
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
	query := `SELECT ` + kycRecordColumns + `
		FROM ` + model.SysDB + `.kyc_record WHERE user_id = ? 
		ORDER BY created_at DESC LIMIT 1`

	record, err := scanKycRecord(r.db.QueryRow(query, userID))
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return record, nil
}

// GetPendingByUserID 获取用户最新进行中（status=1）的认证记录
func (r *KycRecordRepository) GetPendingByUserID(userID int64) (*model.KycRecord, error) {
	query := `SELECT ` + kycRecordColumns + `
		FROM ` + model.SysDB + `.kyc_record WHERE user_id = ? AND status = 1 
		ORDER BY created_at DESC LIMIT 1`

	record, err := scanKycRecord(r.db.QueryRow(query, userID))
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return record, nil
}

// GetByUpBizID 根据上游业务ID查询认证记录
func (r *KycRecordRepository) GetByUpBizID(upBizID string) (*model.KycRecord, error) {
	query := `SELECT ` + kycRecordColumns + `
		FROM ` + model.SysDB + `.kyc_record WHERE up_biz_id = ? 
		ORDER BY created_at DESC LIMIT 1`

	record, err := scanKycRecord(r.db.QueryRow(query, upBizID))
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return record, nil
}

// GetPendingRecords 查询所有处理中（status=1）且已获取上游 biz_id 的认证记录（供定时任务主动同步上游结果）
func (r *KycRecordRepository) GetPendingRecords() ([]*model.KycRecord, error) {
	query := `SELECT ` + kycRecordColumns + `
		FROM ` + model.SysDB + `.kyc_record 
		WHERE status = 1 AND up_biz_id IS NOT NULL AND up_biz_id != '' 
		ORDER BY created_at ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]*model.KycRecord, 0)
	for rows.Next() {
		record, err := scanKycRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

// UpdateUpstreamInfo 更新认证记录的上游信息（发起认证成功后写入 token/biz_id/request_id）
func (r *KycRecordRepository) UpdateUpstreamInfo(id int64, token, bizID, requestID string) error {
	query := `UPDATE ` + model.SysDB + `.kyc_record 
		SET up_token = ?, up_biz_id = ?, up_request_id = ?, status = 1, updated_at = ? 
		WHERE id = ?`
	_, err := r.db.Exec(query, token, bizID, requestID, time.Now(), id)
	return err
}

// UpdateResult 更新认证结果
func (r *KycRecordRepository) UpdateResult(id int64, status int, resultCode, resultMessage, resultData string, verifiedAt *time.Time) error {
	query := `UPDATE ` + model.SysDB + `.kyc_record 
			SET status = ?, result_code = ?, result_message = ?, result_data = ?, 
			verified_at = ?, updated_at = ? 
			WHERE id = ?`
	_, err := r.db.Exec(query, status, resultCode, resultMessage, resultData, verifiedAt, time.Now(), id)
	return err
}

// Cancel 取消认证记录（将状态设为 3-认证失败/取消）
func (r *KycRecordRepository) Cancel(id int64) error {
	query := `UPDATE ` + model.SysDB + `.kyc_record SET status = 3, result_message = '用户取消认证', updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, time.Now(), id)
	return err
}

// GetUserKycRecords 分页查询用户认证记录
func (r *KycRecordRepository) GetUserKycRecords(userID int64, page, pageSize int) ([]*model.KycRecord, int64, error) {
	countQuery := `SELECT COUNT(*) FROM ` + model.SysDB + `.kyc_record WHERE user_id = ?`
	var total int64
	if err := r.db.QueryRow(countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `SELECT ` + kycRecordColumns + `
		FROM ` + model.SysDB + `.kyc_record WHERE user_id = ? 
		ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	records := make([]*model.KycRecord, 0)
	for rows.Next() {
		record, err := scanKycRecord(rows)
		if err != nil {
			return nil, 0, err
		}
		records = append(records, record)
	}
	return records, total, nil
}

// GetUserDailyAuthCount 按天统计用户认证次数
func (r *KycRecordRepository) GetUserDailyAuthCount(userID int64, startDate, endDate string) (map[string]int64, error) {
	query := `SELECT DATE_FORMAT(created_at, '%Y-%m-%d') AS d, COUNT(*) AS c
		FROM ` + model.SysDB + `.kyc_record
		WHERE user_id = ? AND DATE(created_at) >= ? AND DATE(created_at) <= ?
		GROUP BY DATE_FORMAT(created_at, '%Y-%m-%d')`

	rows, err := r.db.Query(query, userID, startDate, endDate)
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

// CountUserFreeFailures 统计账户实名（source=1）已失败（status=3）的认证次数（账号终身累计）
func (r *KycRecordRepository) CountUserFreeFailures(userID int64) (int, error) {
	query := `SELECT COUNT(*) FROM ` + model.SysDB + `.kyc_record 
		WHERE user_id = ? AND source = 1 AND status = 3`
	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}