package repository

import (
	"database/sql"
	"time"

	"starloftrpa/internal/model"
)

type KycEnterpriseRepository struct {
	db *sql.DB
}

func NewKycEnterpriseRepository(db *sql.DB) *KycEnterpriseRepository {
	return &KycEnterpriseRepository{db: db}
}

// kycEnterpriseColumns 企业实名记录常用查询列
const kycEnterpriseColumns = `id, user_id, biz_no, 
	COALESCE(company_name, ''), COALESCE(credit_code, ''), COALESCE(legal_name, ''), COALESCE(legal_id_card, ''), 
	source, COALESCE(admin_id, 0), four_factor_status, COALESCE(four_factor_data, ''), 
	COALESCE(up_token, ''), COALESCE(up_biz_id, ''), COALESCE(up_request_id, ''), 
	status, COALESCE(result_code, ''), COALESCE(result_message, ''), COALESCE(result_data, ''), 
	verified_at, created_at, updated_at`

// scanKycEnterprise 将查询结果扫描到 KycEnterprise
func scanKycEnterprise(row interface{ Scan(...interface{}) error }) (*model.KycEnterprise, error) {
	rec := &model.KycEnterprise{}
	err := row.Scan(
		&rec.ID,
		&rec.UserID,
		&rec.BizNo,
		&rec.CompanyName,
		&rec.CreditCode,
		&rec.LegalName,
		&rec.LegalIDCard,
		&rec.Source,
		&rec.AdminID,
		&rec.FourFactorStatus,
		&rec.FourFactorData,
		&rec.UpToken,
		&rec.UpBizID,
		&rec.UpRequestID,
		&rec.Status,
		&rec.ResultCode,
		&rec.ResultMessage,
		&rec.ResultData,
		&rec.VerifiedAt,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return rec, nil
}

// Create 创建企业实名记录
func (r *KycEnterpriseRepository) Create(rec *model.KycEnterprise) error {
	query := `INSERT INTO ` + model.SysDB + `.kyc_enterprise 
		(user_id, biz_no, company_name, credit_code, legal_name, legal_id_card, source, admin_id, 
		four_factor_status, four_factor_data, up_token, up_biz_id, up_request_id, 
		status, result_code, result_message, result_data, verified_at, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		rec.UserID,
		rec.BizNo,
		rec.CompanyName,
		rec.CreditCode,
		rec.LegalName,
		rec.LegalIDCard,
		rec.Source,
		rec.AdminID,
		rec.FourFactorStatus,
		rec.FourFactorData,
		rec.UpToken,
		rec.UpBizID,
		rec.UpRequestID,
		rec.Status,
		rec.ResultCode,
		rec.ResultMessage,
		rec.ResultData,
		rec.VerifiedAt,
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
	rec.ID = id
	return nil
}

// GetByBizNo 根据业务流水号查询企业实名记录
func (r *KycEnterpriseRepository) GetByBizNo(bizNo string) (*model.KycEnterprise, error) {
	query := `SELECT ` + kycEnterpriseColumns + `
		FROM ` + model.SysDB + `.kyc_enterprise WHERE biz_no = ?`
	rec, err := scanKycEnterprise(r.db.QueryRow(query, bizNo))
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return rec, nil
}

// GetByUpBizID 根据上游业务ID查询企业实名记录
func (r *KycEnterpriseRepository) GetByUpBizID(upBizID string) (*model.KycEnterprise, error) {
	query := `SELECT ` + kycEnterpriseColumns + `
		FROM ` + model.SysDB + `.kyc_enterprise WHERE up_biz_id = ? ORDER BY id DESC LIMIT 1`
	rec, err := scanKycEnterprise(r.db.QueryRow(query, upBizID))
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return rec, nil
}

// GetLatestByUserID 获取用户最近一次企业实名记录
func (r *KycEnterpriseRepository) GetLatestByUserID(userID int64) (*model.KycEnterprise, error) {
	query := `SELECT ` + kycEnterpriseColumns + `
		FROM ` + model.SysDB + `.kyc_enterprise WHERE user_id = ? ORDER BY created_at DESC LIMIT 1`
	rec, err := scanKycEnterprise(r.db.QueryRow(query, userID))
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return rec, nil
}

// GetPendingByUserID 获取用户最新待法人扫脸（status=1）的企业实名记录
func (r *KycEnterpriseRepository) GetPendingByUserID(userID int64) (*model.KycEnterprise, error) {
	query := `SELECT ` + kycEnterpriseColumns + `
		FROM ` + model.SysDB + `.kyc_enterprise WHERE user_id = ? AND status = 1 ORDER BY created_at DESC LIMIT 1`
	rec, err := scanKycEnterprise(r.db.QueryRow(query, userID))
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return rec, nil
}

// UpdateUpstreamInfo 更新法人扫脸上游信息
func (r *KycEnterpriseRepository) UpdateUpstreamInfo(id int64, token, bizID, requestID string) error {
	query := `UPDATE ` + model.SysDB + `.kyc_enterprise 
		SET up_token = ?, up_biz_id = ?, up_request_id = ?, status = 1, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, token, bizID, requestID, time.Now(), id)
	return err
}

// UpdateFourFactor 更新工商四要素核验结果
func (r *KycEnterpriseRepository) UpdateFourFactor(id int64, status int, data string) error {
	query := `UPDATE ` + model.SysDB + `.kyc_enterprise 
		SET four_factor_status = ?, four_factor_data = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, status, data, time.Now(), id)
	return err
}

// UpdateResult 更新企业实名结果（成功后置状态 2 及通过时间）
func (r *KycEnterpriseRepository) UpdateResult(id int64, status int, resultCode, resultMessage, resultData string, verifiedAt *time.Time) error {
	query := `UPDATE ` + model.SysDB + `.kyc_enterprise 
		SET status = ?, result_code = ?, result_message = ?, result_data = ?, verified_at = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, status, resultCode, resultMessage, resultData, verifiedAt, time.Now(), id)
	return err
}

// GetEnterpriseRecords 分页查询企业实名记录（管理后台）
func (r *KycEnterpriseRepository) GetEnterpriseRecords(page, pageSize int) ([]*model.KycEnterprise, int64, error) {
	countQuery := `SELECT COUNT(*) FROM ` + model.SysDB + `.kyc_enterprise`
	var total int64
	if err := r.db.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `SELECT ` + kycEnterpriseColumns + `
		FROM ` + model.SysDB + `.kyc_enterprise ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := r.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]*model.KycEnterprise, 0)
	for rows.Next() {
		rec, err := scanKycEnterprise(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, rec)
	}
	return list, total, nil
}
