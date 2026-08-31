package repository

import (
	"database/sql"
	"time"

	"starloftrpa/internal/model"
)

// SystemConfigRepository 系统配置仓库
type SystemConfigRepository struct {
	db *sql.DB
}

func NewSystemConfigRepository(db *sql.DB) *SystemConfigRepository {
	return &SystemConfigRepository{db: db}
}

// GetConfig 根据配置键查询配置
func (r *SystemConfigRepository) GetConfig(key string) (string, error) {
	query := `SELECT config_value FROM ` + model.SysDB + `.system_config WHERE config_key = ?`
	var value string
	err := r.db.QueryRow(query, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

// GetAllConfigs 获取所有配置
func (r *SystemConfigRepository) GetAllConfigs() (map[string]string, error) {
	query := `SELECT config_key, config_value FROM ` + model.SysDB + `.system_config`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	configs := make(map[string]string)
	for rows.Next() {
		var key, value string
		err := rows.Scan(&key, &value)
		if err != nil {
			return nil, err
		}
		configs[key] = value
	}

	return configs, nil
}

// BatchUpdateConfigs 批量更新配置
func (r *SystemConfigRepository) BatchUpdateConfigs(configs map[string]string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `UPDATE ` + model.SysDB + `.system_config SET config_value = ?, updated_at = ? WHERE config_key = ?`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for key, value := range configs {
		_, err := stmt.Exec(value, time.Now(), key)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
