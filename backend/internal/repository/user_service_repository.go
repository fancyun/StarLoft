package repository

import (
	"database/sql"

	"starloftrpa/internal/model"
)

type UserServiceRepository struct {
	db *sql.DB
}

func NewUserServiceRepository(db *sql.DB) *UserServiceRepository {
	return &UserServiceRepository{db: db}
}

// Create 开通服务（用户+服务唯一，重复开通返回已存在）
func (r *UserServiceRepository) Create(userID int64, serviceCode string) (*model.UserService, error) {
	query := `INSERT INTO ` + model.SysDB + `.user_service (user_id, service_code, status, opened_at, created_at, updated_at) 
		VALUES (?, ?, 1, NOW(), NOW(), NOW())`
	if _, err := r.db.Exec(query, userID, serviceCode); err != nil {
		return nil, err
	}
	return r.GetByUserAndService(userID, serviceCode)
}

// GetByUserAndService 查询指定用户某服务的开通记录
func (r *UserServiceRepository) GetByUserAndService(userID int64, serviceCode string) (*model.UserService, error) {
	query := `SELECT id, user_id, service_code, status, opened_at, created_at, updated_at 
		FROM ` + model.SysDB + `.user_service WHERE user_id = ? AND service_code = ? ORDER BY id DESC LIMIT 1`
	us := &model.UserService{}
	err := r.db.QueryRow(query, userID, serviceCode).Scan(
		&us.ID, &us.UserID, &us.ServiceCode, &us.Status, &us.OpenedAt, &us.CreatedAt, &us.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return us, nil
}

// IsOpen 判断指定用户服务是否已开通（存在且状态为已开通）
func (r *UserServiceRepository) IsOpen(userID int64, serviceCode string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM ` + model.SysDB + `.user_service WHERE user_id = ? AND service_code = ? AND status = 1`
	if err := r.db.QueryRow(query, userID, serviceCode).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListByUser 查询指定用户已开通的所有服务
func (r *UserServiceRepository) ListByUser(userID int64) ([]*model.UserService, error) {
	query := `SELECT id, user_id, service_code, status, opened_at, created_at, updated_at 
		FROM ` + model.SysDB + `.user_service WHERE user_id = ? AND status = 1 ORDER BY opened_at DESC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*model.UserService, 0)
	for rows.Next() {
		us := &model.UserService{}
		if err := rows.Scan(&us.ID, &us.UserID, &us.ServiceCode, &us.Status, &us.OpenedAt, &us.CreatedAt, &us.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, us)
	}
	return list, nil
}
