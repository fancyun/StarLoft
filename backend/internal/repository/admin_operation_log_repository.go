package repository

import (
	"database/sql"

	"starloftrpa/internal/model"
)

// AdminOperationLogRepository 管理员操作日志仓储
type AdminOperationLogRepository struct {
	db *sql.DB
}

func NewAdminOperationLogRepository(db *sql.DB) *AdminOperationLogRepository {
	return &AdminOperationLogRepository{db: db}
}

// InsertOperationLog 写入一条管理员操作日志
func (r *AdminOperationLogRepository) InsertOperationLog(log *model.AdminOperationLog) error {
	query := `INSERT INTO admin_operation_log 
		(admin_id, operation, resource_type, resource_id, details, ip_address) 
		VALUES (?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		log.AdminID,
		log.Operation,
		nullIfEmpty(log.ResourceType),
		nullIfZero(log.ResourceID),
		nullIfEmpty(log.Details),
		nullIfEmpty(log.IPAddress),
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

// GetOperationLogs 分页查询管理员操作日志（联表获取管理员用户名）
// adminID 为 0 表示不过滤；operation、adminName 为空表示不过滤
func (r *AdminOperationLogRepository) GetOperationLogs(page, pageSize int, adminID int64, operation, adminName string) ([]*model.AdminOperationLog, int64, error) {
	offset := (page - 1) * pageSize

	// 构建动态 WHERE 条件（参数化查询，防 SQL 注入）
	where := "WHERE 1=1"
	args := make([]interface{}, 0, 3)
	if adminID > 0 {
		where += " AND l.admin_id = ?"
		args = append(args, adminID)
	}
	if operation != "" {
		where += " AND l.operation = ?"
		args = append(args, operation)
	}
	if adminName != "" {
		where += " AND a.username LIKE ?"
		args = append(args, "%"+adminName+"%")
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM admin_operation_log l LEFT JOIN admin_user a ON a.id = l.admin_id ` + where
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT l.id, l.admin_id, a.username, l.operation, l.resource_type, l.resource_id, l.details, l.ip_address, l.created_at 
		FROM admin_operation_log l 
		LEFT JOIN admin_user a ON a.id = l.admin_id 
		` + where + ` 
		ORDER BY l.id DESC LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	logs := make([]*model.AdminOperationLog, 0)
	for rows.Next() {
		log := &model.AdminOperationLog{}
		if err := rows.Scan(
			&log.ID,
			&log.AdminID,
			&log.AdminName,
			&log.Operation,
			&log.ResourceType,
			&log.ResourceID,
			&log.Details,
			&log.IPAddress,
			&log.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}

// nullIfEmpty 空字符串转为 NULL
func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// nullIfZero 0 值转为 NULL
func nullIfZero(v int64) interface{} {
	if v == 0 {
		return nil
	}
	return v
}
