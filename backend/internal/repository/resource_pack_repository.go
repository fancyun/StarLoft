package repository

import (
	"database/sql"
	"errors"
	"time"

	"starloftrpa/internal/model"
)

var (
	ErrPackNotFound   = errors.New("resource pack not found")
	ErrPackSoldOut    = errors.New("resource pack sold out")
	ErrPackOffSale    = errors.New("resource pack off sale")
	ErrPackCountEmpty = errors.New("resource pack count empty")
)

type ResourcePackRepository struct {
	db *sql.DB
}

func NewResourcePackRepository(db *sql.DB) *ResourcePackRepository {
	return &ResourcePackRepository{db: db}
}

// ---------- 资源包定义 ----------

// CreatePack 创建资源包
func (r *ResourcePackRepository) CreatePack(pack *model.ResourcePack) error {
	query := `INSERT INTO resource_pack (name, total_count, price, stock, status, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := r.db.Exec(query,
		pack.Name, pack.TotalCount, pack.Price, pack.Stock, pack.Status, pack.Description,
		time.Now(), time.Now(),
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	pack.ID = id
	return nil
}

// UpdatePack 更新资源包
func (r *ResourcePackRepository) UpdatePack(pack *model.ResourcePack) error {
	query := `UPDATE resource_pack SET name = ?, total_count = ?, price = ?, stock = ?, status = ?, description = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query,
		pack.Name, pack.TotalCount, pack.Price, pack.Stock, pack.Status, pack.Description,
		time.Now(), pack.ID,
	)
	return err
}

// GetPackByID 根据ID查询资源包
func (r *ResourcePackRepository) GetPackByID(id int64) (*model.ResourcePack, error) {
	query := `SELECT id, name, total_count, price, stock, status, description, created_at, updated_at FROM resource_pack WHERE id = ?`
	pack := &model.ResourcePack{}
	err := r.db.QueryRow(query, id).Scan(
		&pack.ID, &pack.Name, &pack.TotalCount, &pack.Price, &pack.Stock,
		&pack.Status, &pack.Description, &pack.CreatedAt, &pack.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrPackNotFound
	}
	if err != nil {
		return nil, err
	}
	return pack, nil
}

// ListPacks 查询资源包列表（status 为 nil 时查询全部；否则按状态过滤）
func (r *ResourcePackRepository) ListPacks(status *int) ([]*model.ResourcePack, error) {
	query := `SELECT id, name, total_count, price, stock, status, description, created_at, updated_at FROM resource_pack`
	args := make([]interface{}, 0)
	if status != nil {
		query += ` WHERE status = ?`
		args = append(args, *status)
	}
	query += ` ORDER BY id ASC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	packs := make([]*model.ResourcePack, 0)
	for rows.Next() {
		pack := &model.ResourcePack{}
		if err := rows.Scan(
			&pack.ID, &pack.Name, &pack.TotalCount, &pack.Price, &pack.Stock,
			&pack.Status, &pack.Description, &pack.CreatedAt, &pack.UpdatedAt,
		); err != nil {
			return nil, err
		}
		packs = append(packs, pack)
	}
	return packs, nil
}

// DecrementStockTx 扣减资源包库存（在购买事务中调用）
// 返回 false 表示库存不足/已售罄
func (r *ResourcePackRepository) DecrementStockTx(tx *sql.Tx, packID int64) (bool, error) {
	// 先锁定并读取库存
	query := `SELECT stock FROM resource_pack WHERE id = ? FOR UPDATE`
	var stock int
	err := tx.QueryRow(query, packID).Scan(&stock)
	if err == sql.ErrNoRows {
		return false, ErrPackNotFound
	}
	if err != nil {
		return false, err
	}

	if stock == -1 {
		// 不限量
		return true, nil
	}
	if stock <= 0 {
		return false, ErrPackSoldOut
	}

	_, err = tx.Exec(`UPDATE resource_pack SET stock = stock - 1, updated_at = ? WHERE id = ?`, time.Now(), packID)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ---------- 用户资源包 ----------

// CreateUserPackTx 在事务中创建用户资源包
func (r *ResourcePackRepository) CreateUserPackTx(tx *sql.Tx, up *model.UserResourcePack) error {
	query := `INSERT INTO user_resource_pack (user_id, pack_id, pack_name, total_count, remaining_count, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := tx.Exec(query,
		up.UserID, up.PackID, up.PackName, up.TotalCount, up.RemainingCount, up.Status,
		time.Now(), time.Now(),
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	up.ID = id
	return nil
}

// GetUserActivePack 获取用户第一个有效资源包（剩余次数 > 0），按购买时间升序
func (r *ResourcePackRepository) GetUserActivePack(userID int64) (*model.UserResourcePack, error) {
	query := `SELECT id, user_id, pack_id, pack_name, total_count, remaining_count, status, created_at, updated_at
		FROM user_resource_pack
		WHERE user_id = ? AND status = 1 AND remaining_count > 0
		ORDER BY created_at ASC
		LIMIT 1`
	up := &model.UserResourcePack{}
	err := r.db.QueryRow(query, userID).Scan(
		&up.ID, &up.UserID, &up.PackID, &up.PackName, &up.TotalCount,
		&up.RemainingCount, &up.Status, &up.CreatedAt, &up.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return up, nil
}

// DeductUserPackCountTx 在事务中扣减用户资源包次数，返回是否扣减成功
func (r *ResourcePackRepository) DeductUserPackCountTx(tx *sql.Tx, id, userID int64) (bool, error) {
	query := `UPDATE user_resource_pack
		SET remaining_count = remaining_count - 1,
		    status = IF(remaining_count - 1 <= 0, 0, 1),
		    updated_at = ?
		WHERE id = ? AND user_id = ? AND remaining_count > 0 AND status = 1`
	result, err := tx.Exec(query, time.Now(), id, userID)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

// DeductUserPackCount 扣减用户资源包次数（单条原子 UPDATE，无需事务）
// 返回 false 表示资源包已耗尽/无剩余次数（并发安全，由 remaining_count > 0 条件保证）
func (r *ResourcePackRepository) DeductUserPackCount(id, userID int64) (bool, error) {
	query := `UPDATE user_resource_pack
		SET remaining_count = remaining_count - 1,
		    status = IF(remaining_count - 1 <= 0, 0, 1),
		    updated_at = ?
		WHERE id = ? AND user_id = ? AND remaining_count > 0 AND status = 1`
	result, err := r.db.Exec(query, time.Now(), id, userID)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

// RefundUserPackCount 退还资源包次数（退款/不计费时加回）
func (r *ResourcePackRepository) RefundUserPackCount(id int64) error {
	query := `UPDATE user_resource_pack SET remaining_count = remaining_count + 1, status = 1, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, time.Now(), id)
	return err
}

// ListUserPacks 查询用户全部资源包（含已耗尽），按购买时间倒序
func (r *ResourcePackRepository) ListUserPacks(userID int64) ([]*model.UserResourcePack, error) {
	query := `SELECT id, user_id, pack_id, pack_name, total_count, remaining_count, status, created_at, updated_at
		FROM user_resource_pack
		WHERE user_id = ?
		ORDER BY id DESC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*model.UserResourcePack, 0)
	for rows.Next() {
		up := &model.UserResourcePack{}
		if err := rows.Scan(
			&up.ID, &up.UserID, &up.PackID, &up.PackName, &up.TotalCount,
			&up.RemainingCount, &up.Status, &up.CreatedAt, &up.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, up)
	}
	return list, nil
}
