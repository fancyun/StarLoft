package model

import (
	"database/sql"
	"time"
)

// KycRecord 账户实名认证记录表
// 记录平台用户每次实名认证的详细信息，包括认证状态、结果等
type KycRecord struct {
	ID            int64         `json:"id"`
	UserID        int64         `json:"user_id"`                  // 用户ID
	AuthOrderID   sql.NullInt64 `json:"auth_order_id,omitempty"`  // 关联认证订单ID（通过下单完成实名时关联）
	Name          string        `json:"name"`                     // 实名姓名
	IDCard        string        `json:"id_card"`                  // 身份证号
	Status        int           `json:"status"`                   // 0-待认证 1-认证中 2-认证成功 3-认证失败 4-已更换
	ResultCode    string        `json:"result_code,omitempty"`    // 上游认证结果码
	ResultMessage string        `json:"result_message,omitempty"` // 认证结果消息
	ResultData    string        `json:"result_data,omitempty"`    // 认证结果完整数据（JSON）
	VerifiedAt    *time.Time    `json:"verified_at,omitempty"`    // 认证通过时间
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}
