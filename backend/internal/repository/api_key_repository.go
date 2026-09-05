package repository

import (
	"database/sql"
	"time"

	"starloftrpa/internal/model"
)

// ApiKeyRepository 平台 API 密钥存取（位于 sys 库 api 表）
type ApiKeyRepository struct {
	db *sql.DB
}

func NewApiKeyRepository(db *sql.DB) *ApiKeyRepository {
	return &ApiKeyRepository{db: db}
}

const apiKeyColumns = `id, user_id, api_key, api_secret, permission, created_at, updated_at`

func scanApiKey(row interface{ Scan(...interface{}) error }) (*model.ApiKey, error) {
	k := &model.ApiKey{}
	err := row.Scan(&k.ID, &k.UserID, &k.APIKey, &k.APISecret, &k.Permission, &k.CreatedAt, &k.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return k, nil
}

// GetByAPIKey 根据 API Key 查询密钥记录
func (r *ApiKeyRepository) GetByAPIKey(apiKey string) (*model.ApiKey, error) {
	query := `SELECT ` + apiKeyColumns + ` FROM ` + model.SysDB + `.api WHERE api_key = ?`
	k, err := scanApiKey(r.db.QueryRow(query, apiKey))
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return k, nil
}

// GetByUser 查询用户当前生效的 API 密钥（单密钥，取最新一条）
func (r *ApiKeyRepository) GetByUser(userID int64) (*model.ApiKey, error) {
	query := `SELECT ` + apiKeyColumns + ` FROM ` + model.SysDB + `.api WHERE user_id = ? ORDER BY id DESC LIMIT 1`
	k, err := scanApiKey(r.db.QueryRow(query, userID))
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return k, nil
}

// Create 新增一条 API 密钥
func (r *ApiKeyRepository) Create(k *model.ApiKey) error {
	query := `INSERT INTO ` + model.SysDB + `.api (user_id, api_key, api_secret, permission, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?)`
	res, err := r.db.Exec(query, k.UserID, k.APIKey, k.APISecret, k.Permission, time.Now(), time.Now())
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		k.ID = id
	}
	return nil
}

// DeleteByUser 删除用户全部 API 密钥（重置时先清空再重建）
func (r *ApiKeyRepository) DeleteByUser(userID int64) error {
	_, err := r.db.Exec(`DELETE FROM `+model.SysDB+`.api WHERE user_id = ?`, userID)
	return err
}

// SetPermission 更新用户 API 密钥的权限范围（all 或单个服务标识）
func (r *ApiKeyRepository) SetPermission(userID int64, permission string) error {
	_, err := r.db.Exec(`UPDATE `+model.SysDB+`.api SET permission = ?, updated_at = ? WHERE user_id = ?`,
		permission, time.Now(), userID)
	return err
}