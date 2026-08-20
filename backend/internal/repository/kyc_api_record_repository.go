package repository

import (
	"database/sql"
	"time"

	"starloftrpa/internal/model"
)

type KycApiRecordRepository struct {
	db *sql.DB
}

func NewKycApiRecordRepository(db *sql.DB) *KycApiRecordRepository {
	return &KycApiRecordRepository{db: db}
}

// Create 创建API调用记录
func (r *KycApiRecordRepository) Create(record *model.KycApiRecord) error {
	query := `INSERT INTO kyc_api_record 
		(user_id, auth_order_id, api_type, request_data, response_data, 
		http_status, cost, duration_ms, error_message, ip_address, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		record.UserID,
		record.AuthOrderID,
		record.ApiType,
		record.RequestData,
		record.ResponseData,
		record.HttpStatus,
		record.Cost,
		record.DurationMs,
		record.ErrorMessage,
		record.IPAddress,
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

// GetByID 根据ID查询记录
func (r *KycApiRecordRepository) GetByID(id int64) (*model.KycApiRecord, error) {
	query := `SELECT id, user_id, auth_order_id, api_type, request_data, response_data, 
		http_status, cost, duration_ms, error_message, ip_address, created_at 
		FROM kyc_api_record WHERE id = ?`

	record := &model.KycApiRecord{}
	err := r.db.QueryRow(query, id).Scan(
		&record.ID,
		&record.UserID,
		&record.AuthOrderID,
		&record.ApiType,
		&record.RequestData,
		&record.ResponseData,
		&record.HttpStatus,
		&record.Cost,
		&record.DurationMs,
		&record.ErrorMessage,
		&record.IPAddress,
		&record.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return record, nil
}

// GetByUserID 查询用户的所有API调用记录（分页）
func (r *KycApiRecordRepository) GetByUserID(userID int64, page, pageSize int) ([]*model.KycApiRecord, int64, error) {
	countQuery := `SELECT COUNT(*) FROM kyc_api_record WHERE user_id = ?`
	var total int64
	err := r.db.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `SELECT id, user_id, auth_order_id, api_type, request_data, response_data, 
		http_status, cost, duration_ms, error_message, ip_address, created_at 
		FROM kyc_api_record WHERE user_id = ? 
		ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	records := make([]*model.KycApiRecord, 0)
	for rows.Next() {
		record := &model.KycApiRecord{}
		err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.AuthOrderID,
			&record.ApiType,
			&record.RequestData,
			&record.ResponseData,
			&record.HttpStatus,
			&record.Cost,
			&record.DurationMs,
			&record.ErrorMessage,
			&record.IPAddress,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		records = append(records, record)
	}

	return records, total, nil
}

// GetByOrderID 根据订单ID查询关联的API调用记录
func (r *KycApiRecordRepository) GetByOrderID(orderID int64) ([]*model.KycApiRecord, error) {
	query := `SELECT id, user_id, auth_order_id, api_type, request_data, response_data, 
		http_status, cost, duration_ms, error_message, ip_address, created_at 
		FROM kyc_api_record WHERE auth_order_id = ? 
		ORDER BY created_at ASC`

	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]*model.KycApiRecord, 0)
	for rows.Next() {
		record := &model.KycApiRecord{}
		err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.AuthOrderID,
			&record.ApiType,
			&record.RequestData,
			&record.ResponseData,
			&record.HttpStatus,
			&record.Cost,
			&record.DurationMs,
			&record.ErrorMessage,
			&record.IPAddress,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// GetAllRecords 管理员查询所有API调用记录（分页+筛选）
func (r *KycApiRecordRepository) GetAllRecords(page, pageSize int, userID *int64, apiType *string) ([]*model.KycApiRecord, int64, error) {
	offset := (page - 1) * pageSize

	whereClause := ""
	args := []interface{}{}

	if userID != nil {
		whereClause = "WHERE user_id = ?"
		args = append(args, *userID)
	}

	if apiType != nil && *apiType != "" {
		if whereClause == "" {
			whereClause = "WHERE api_type = ?"
		} else {
			whereClause += " AND api_type = ?"
		}
		args = append(args, *apiType)
	}

	countQuery := "SELECT COUNT(*) FROM kyc_api_record " + whereClause
	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT id, user_id, auth_order_id, api_type, request_data, response_data, 
		http_status, cost, duration_ms, error_message, ip_address, created_at 
		FROM kyc_api_record ` + whereClause + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`

	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	records := make([]*model.KycApiRecord, 0)
	for rows.Next() {
		record := &model.KycApiRecord{}
		err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.AuthOrderID,
			&record.ApiType,
			&record.RequestData,
			&record.ResponseData,
			&record.HttpStatus,
			&record.Cost,
			&record.DurationMs,
			&record.ErrorMessage,
			&record.IPAddress,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		records = append(records, record)
	}

	return records, total, nil
}
