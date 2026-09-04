package model

import (
	"database/sql"
	"time"
)

// KycRecord 账户实名认证记录表
// 记录平台用户每次实名认证的详细信息，包括认证状态、结果等
func (KycRecord) TableName() string { return SysDB + ".kyc_record" }

type KycRecord struct {
	ID             int64         `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID         int64         `json:"user_id" gorm:"not null;index"`          // 用户ID
	Source         int           `json:"source" gorm:"type:tinyint;not null;default:2"` // 来源：1-账户实名 2-API调用
	AuthOrderID    sql.NullInt64 `json:"auth_order_id,omitempty" gorm:"index"`   // 关联认证订单ID（通过下单完成实名时关联；账户实名不关联订单）
	PlatformBizNo  string        `json:"platform_biz_no,omitempty" gorm:"size:50;index"` // 平台业务流水号
	BizNo          string        `json:"biz_no,omitempty" gorm:"size:50;index"`            // 用户业务流水号
	ReturnURL      string        `json:"return_url,omitempty" gorm:"size:500"`  // 认证完成后跳转的URL
	NotifyURL      string        `json:"notify_url,omitempty" gorm:"size:500"`  // 异步通知回调URL
	BizExtraData   string        `json:"biz_extra_data,omitempty" gorm:"type:text"` // 额外业务数据
	UpToken        string        `json:"up_token,omitempty" gorm:"size:100"`    // 上游返回的token
	UpBizID        string        `json:"up_biz_id,omitempty" gorm:"size:50;index"` // 上游返回的biz_id
	UpRequestID    string        `json:"up_request_id,omitempty" gorm:"size:50"`   // 上游返回的request_id
	Name           string        `json:"name" gorm:"size:50;not null"`             // 实名姓名
	IDCard         string        `json:"id_card" gorm:"size:18;not null"`          // 身份证号
	Status         int           `json:"status" gorm:"type:tinyint;not null;default:0;index"` // 0-待认证 1-认证中 2-认证成功 3-认证失败 4-已更换
	ResultCode     string        `json:"result_code,omitempty" gorm:"size:20"`     // 上游认证结果码
	ResultMessage  string        `json:"result_message,omitempty" gorm:"size:255"` // 认证结果消息
	ResultData     string        `json:"result_data,omitempty" gorm:"type:text"`   // 认证结果完整数据（JSON）
	Cost           float64       `json:"cost" gorm:"-"`                            // 本次认证消耗金额（账户实名免费，恒为0；仅用于列表展示）
	VerifiedAt     *time.Time    `json:"verified_at,omitempty"`                    // 认证通过时间
	CreatedAt      time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
}
