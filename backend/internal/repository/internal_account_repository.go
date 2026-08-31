package repository

import (
	"database/sql"
	"errors"
	"time"

	"starloftrpa/internal/model"
)

// InternalAccountRepository 内部账号仓库
type InternalAccountRepository struct {
	db *sql.DB
}

func NewInternalAccountRepository(db *sql.DB) *InternalAccountRepository {
	return &InternalAccountRepository{db: db}
}

const internalAccountColumns = `id, name, remark, api_key, api_secret, status, created_at, updated_at`

func scanInternalAccount(row *sql.Row) (*model.InternalAccount, error) {
	acc := &model.InternalAccount{}
	err := row.Scan(&acc.ID, &acc.Name, &acc.Remark, &acc.APIKey, &acc.APISecret, &acc.Status, &acc.CreatedAt, &acc.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

// Create 创建内部账号
func (r *InternalAccountRepository) Create(acc *model.InternalAccount) error {
	query := `INSERT INTO ` + model.SysDB + `.internal_account (name, remark, api_key, api_secret, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	result, err := r.db.Exec(query, acc.Name, acc.Remark, acc.APIKey, acc.APISecret, acc.Status, time.Now(), time.Now())
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	acc.ID = id
	return nil
}

// GetByID 根据ID查询
func (r *InternalAccountRepository) GetByID(id int64) (*model.InternalAccount, error) {
	query := `SELECT ` + internalAccountColumns + ` FROM ` + model.SysDB + `.internal_account WHERE id = ?`
	return scanInternalAccount(r.db.QueryRow(query, id))
}

// GetByAPIKey 根据API Key查询
func (r *InternalAccountRepository) GetByAPIKey(apiKey string) (*model.InternalAccount, error) {
	query := `SELECT ` + internalAccountColumns + ` FROM ` + model.SysDB + `.internal_account WHERE api_key = ?`
	return scanInternalAccount(r.db.QueryRow(query, apiKey))
}

// GetByName 根据名称查询（用于名称唯一性校验）
func (r *InternalAccountRepository) GetByName(name string) (*model.InternalAccount, error) {
	query := `SELECT ` + internalAccountColumns + ` FROM ` + model.SysDB + `.internal_account WHERE name = ?`
	return scanInternalAccount(r.db.QueryRow(query, name))
}

// List 分页查询内部账号列表
func (r *InternalAccountRepository) List(keyword string, page, pageSize int) ([]*model.InternalAccount, int64, error) {
	whereClause := ""
	args := []interface{}{}
	if keyword != "" {
		whereClause = ` WHERE name LIKE ? OR remark LIKE ?`
		like := "%" + keyword + "%"
		args = append(args, like, like)
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM ` + model.SysDB + `.internal_account` + whereClause
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `SELECT ` + internalAccountColumns + ` FROM ` + model.SysDB + `.internal_account` + whereClause + ` ORDER BY id DESC LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	accounts := make([]*model.InternalAccount, 0)
	for rows.Next() {
		acc := &model.InternalAccount{}
		if err := rows.Scan(&acc.ID, &acc.Name, &acc.Remark, &acc.APIKey, &acc.APISecret, &acc.Status, &acc.CreatedAt, &acc.UpdatedAt); err != nil {
			return nil, 0, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, total, nil
}

// UpdateStatus 启用/禁用内部账号
func (r *InternalAccountRepository) UpdateStatus(id int64, status int) error {
	query := `UPDATE ` + model.SysDB + `.internal_account SET status = ?, updated_at = ? WHERE id = ?`
	result, err := r.db.Exec(query, status, time.Now(), id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("internal account not found")
	}
	return nil
}

// ResetAPIKey 重置内部账号API密钥（Key 与 Secret 一并重新生成）
func (r *InternalAccountRepository) ResetAPIKey(id int64, apiKey, apiSecret string) error {
	query := `UPDATE ` + model.SysDB + `.internal_account SET api_key = ?, api_secret = ?, updated_at = ? WHERE id = ?`
	result, err := r.db.Exec(query, apiKey, apiSecret, time.Now(), id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("internal account not found")
	}
	return nil
}
