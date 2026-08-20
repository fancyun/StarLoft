package repository

import (
	"database/sql"
	"time"

	"starloftrpa/internal/model"
)

type AdminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// GetAdminByUsername 根据用户名查询管理员
func (r *AdminRepository) GetAdminByUsername(username string) (*model.AdminUser, error) {
	query := `SELECT id, username, password_hash, nickname, status, last_login_at, created_at, updated_at 
		FROM admin_user WHERE username = ?`

	admin := &model.AdminUser{}
	err := r.db.QueryRow(query, username).Scan(
		&admin.ID,
		&admin.Username,
		&admin.PasswordHash,
		&admin.Nickname,
		&admin.Status,
		&admin.LastLoginAt,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return admin, nil
}

// UpdateLastLoginTime 更新最后登录时间
func (r *AdminRepository) UpdateLastLoginTime(adminID int64) error {
	query := `UPDATE admin_user SET last_login_at = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, time.Now(), time.Now(), adminID)
	return err
}
