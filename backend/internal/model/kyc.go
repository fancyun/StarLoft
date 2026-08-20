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

// KycApiRecord 用户API调用记录表
// 记录平台用户每次调用实名认证API的请求和响应信息
type KycApiRecord struct {
	ID           int64         `json:"id"`
	UserID       int64         `json:"user_id"`                 // 调用用户ID
	AuthOrderID  sql.NullInt64 `json:"auth_order_id,omitempty"` // 关联认证订单ID
	ApiType      string        `json:"api_type"`                // API类型：get_token / get_result / create_order
	RequestData  string        `json:"request_data,omitempty"`  // 请求数据（脱敏后）
	ResponseData string        `json:"response_data,omitempty"` // 响应数据
	HttpStatus   int           `json:"http_status"`             // HTTP状态码
	Cost         float64       `json:"cost"`                    // 本次调用消耗金额
	DurationMs   int           `json:"duration_ms"`             // 接口耗时（毫秒）
	ErrorMessage string        `json:"error_message,omitempty"` // 错误信息
	IPAddress    string        `json:"ip_address,omitempty"`    // 请求IP地址
	CreatedAt    time.Time     `json:"created_at"`
}
