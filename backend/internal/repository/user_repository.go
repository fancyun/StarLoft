package repository

import (
	"database/sql"
	"errors"
	"time"

	"starloftrpa/internal/model"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidCredential = errors.New("invalid credential")
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser 创建用户
func (r *UserRepository) CreateUser(user *model.PlatformUser) error {
	query := `INSERT INTO platform_user 
		(phone, password_hash, balance, api_key, api_secret, is_kyc_verified, kyc_price, status, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		user.Phone,
		user.PasswordHash,
		user.Balance,
		user.APIKey,
		user.APISecret,
		user.IsKYCVerified,
		user.KYCPrice, // KYC单价，注册时设置
		user.Status,
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
	user.ID = id
	return nil
}

// GetUserByPhone 根据手机号查询用户
func (r *UserRepository) GetUserByPhone(phone string) (*model.PlatformUser, error) {
	query := `SELECT id, phone, password_hash, balance, api_key, api_secret, is_kyc_verified, kyc_name, kyc_id_card, kyc_price, status, last_login_at, created_at, updated_at 
		FROM platform_user WHERE phone = ?`

	user := &model.PlatformUser{}
	err := r.db.QueryRow(query, phone).Scan(
		&user.ID,
		&user.Phone,
		&user.PasswordHash,
		&user.Balance,
		&user.APIKey,
		&user.APISecret,
		&user.IsKYCVerified,
		&user.KYCName,
		&user.KYCIDCard,
		&user.KYCPrice,
		&user.Status,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByID 根据用户ID查询用户
func (r *UserRepository) GetUserByID(id int64) (*model.PlatformUser, error) {
	query := `SELECT id, phone, password_hash, balance, api_key, api_secret, 
		is_kyc_verified, kyc_name, kyc_id_card, kyc_price, status, last_login_at, created_at, updated_at 
		FROM platform_user WHERE id = ?`

	user := &model.PlatformUser{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Phone,
		&user.PasswordHash,
		&user.Balance,
		&user.APIKey,
		&user.APISecret,
		&user.IsKYCVerified,
		&user.KYCName,
		&user.KYCIDCard,
		&user.KYCPrice,
		&user.Status,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByAPIKey 根据API Key查询用户
func (r *UserRepository) GetByAPIKey(apiKey string) (*model.PlatformUser, error) {
	query := `SELECT id, phone, password_hash, balance, api_key, api_secret, is_kyc_verified, kyc_name, kyc_id_card, kyc_price, status, last_login_at, created_at, updated_at 
		FROM platform_user WHERE api_key = ?`

	user := &model.PlatformUser{}
	err := r.db.QueryRow(query, apiKey).Scan(
		&user.ID,
		&user.Phone,
		&user.PasswordHash,
		&user.Balance,
		&user.APIKey,
		&user.APISecret,
		&user.IsKYCVerified,
		&user.KYCName,
		&user.KYCIDCard,
		&user.KYCPrice,
		&user.Status,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUserPassword 更新用户密码
func (r *UserRepository) UpdateUserPassword(userID int64, passwordHash string) error {
	query := `UPDATE platform_user SET password_hash = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, passwordHash, time.Now(), userID)
	return err
}

// UpdateUserAPIKey 更新用户API密钥（Key 与 Secret 一并更新）
func (r *UserRepository) UpdateUserAPIKey(userID int64, apiKey, apiSecret string) error {
	query := `UPDATE platform_user SET api_key = ?, api_secret = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, apiKey, apiSecret, time.Now(), userID)
	return err
}

// UpdateUserAPISecret 更新用户API Secret（实名成功后单独生成下发）
func (r *UserRepository) UpdateUserAPISecret(userID int64, apiSecret string) error {
	query := `UPDATE platform_user SET api_secret = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, apiSecret, time.Now(), userID)
	return err
}

// UpdateUserKYCInfo 更新用户实名信息
func (r *UserRepository) UpdateUserKYCInfo(userID int64, name, idCard string) error {
	query := `UPDATE platform_user 
		SET is_kyc_verified = 1, kyc_name = ?, kyc_id_card = ?, updated_at = ? 
		WHERE id = ?`
	_, err := r.db.Exec(query, name, idCard, time.Now(), userID)
	return err
}

// UpdateUserBalance 更新用户余额
func (r *UserRepository) UpdateUserBalance(userID int64, balance float64) error {
	query := `UPDATE platform_user SET balance = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, balance, time.Now(), userID)
	return err
}

// GetBalanceForUpdateTx 在事务中锁定用户余额行并返回余额
func (r *UserRepository) GetBalanceForUpdateTx(tx *sql.Tx, userID int64) (float64, error) {
	var balance float64
	err := tx.QueryRow(`SELECT balance FROM platform_user WHERE id = ? FOR UPDATE`, userID).Scan(&balance)
	if err == sql.ErrNoRows {
		return 0, ErrUserNotFound
	}
	if err != nil {
		return 0, err
	}
	return balance, nil
}

// UpdateUserBalanceTx 在事务中更新用户余额
func (r *UserRepository) UpdateUserBalanceTx(tx *sql.Tx, userID int64, balance float64) error {
	query := `UPDATE platform_user SET balance = ?, updated_at = ? WHERE id = ?`
	_, err := tx.Exec(query, balance, time.Now(), userID)
	return err
}

// UpdateLastLoginTime 更新最后登录时间
func (r *UserRepository) UpdateLastLoginTime(userID int64) error {
	query := `UPDATE platform_user SET last_login_at = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, time.Now(), time.Now(), userID)
	return err
}

// CheckPhoneExists 检查手机号是否已存在
func (r *UserRepository) CheckPhoneExists(phone string) (bool, error) {
	query := `SELECT COUNT(*) FROM platform_user WHERE phone = ?`
	var count int
	err := r.db.QueryRow(query, phone).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetAllUsers 获取用户列表（带分页和搜索）- 修复SQL注入漏洞
func (r *UserRepository) GetAllUsers(page, pageSize int, phone string) ([]*model.PlatformUser, int64, error) {
	offset := (page - 1) * pageSize

	// 使用参数化查询防止SQL注入
	var total int64
	var rows *sql.Rows
	var err error

	if phone != "" {
		// 查询总数
		countQuery := "SELECT COUNT(*) FROM platform_user WHERE phone LIKE ?"
		err = r.db.QueryRow(countQuery, "%"+phone+"%").Scan(&total)
		if err != nil {
			return nil, 0, err
		}

		// 查询列表
		query := `SELECT id, phone, password_hash, balance, api_key, api_secret, 
			is_kyc_verified, kyc_name, kyc_id_card, kyc_price, status, last_login_at, created_at, updated_at 
			FROM platform_user WHERE phone LIKE ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
		rows, err = r.db.Query(query, "%"+phone+"%", pageSize, offset)
	} else {
		// 查询总数
		countQuery := "SELECT COUNT(*) FROM platform_user"
		err = r.db.QueryRow(countQuery).Scan(&total)
		if err != nil {
			return nil, 0, err
		}

		// 查询列表
		query := `SELECT id, phone, password_hash, balance, api_key, api_secret, 
			is_kyc_verified, kyc_name, kyc_id_card, kyc_price, status, last_login_at, created_at, updated_at 
			FROM platform_user ORDER BY created_at DESC LIMIT ? OFFSET ?`
		rows, err = r.db.Query(query, pageSize, offset)
	}

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]*model.PlatformUser, 0)
	for rows.Next() {
		user := &model.PlatformUser{}
		err := rows.Scan(
			&user.ID,
			&user.Phone,
			&user.PasswordHash,
			&user.Balance,
			&user.APIKey,
			&user.APISecret,
			&user.IsKYCVerified,
			&user.KYCName,
			&user.KYCIDCard,
			&user.KYCPrice,
			&user.Status,
			&user.LastLoginAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, total, nil
}

// UpdateUserKYCPrice 更新用户KYC单价
func (r *UserRepository) UpdateUserKYCPrice(userID int64, kycPrice float64) error {
	query := `UPDATE platform_user SET kyc_price = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, kycPrice, time.Now(), userID)
	return err
}

// UpdateUserStatus 更新用户状态
func (r *UserRepository) UpdateUserStatus(userID int64, status int) error {
	query := `UPDATE platform_user SET status = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, status, time.Now(), userID)
	return err
}

// DeleteUser 删除用户
func (r *UserRepository) DeleteUser(userID int64) error {
	query := `DELETE FROM platform_user WHERE id = ?`
	_, err := r.db.Exec(query, userID)
	return err
}

// UpdateUserDiscount 更新用户折扣
func (r *UserRepository) UpdateUserDiscount(userID int64, discount float64) error {
	query := `UPDATE platform_user SET kyc_price = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, discount, time.Now(), userID)
	return err
}
