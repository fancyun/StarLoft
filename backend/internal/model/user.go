package model

import (
	"database/sql"
	"time"
)

// PlatformUser 平台用户（使用 Web 前端的用户）
type PlatformUser struct {
	ID            int64          `json:"id"`
	Phone         string         `json:"phone"`
	PasswordHash  string         `json:"-"`
	Balance       float64        `json:"balance"`
	APIKey        string         `json:"api_key"`
	APISecret     string         `json:"-"`
	IsKYCVerified int            `json:"is_kyc_verified"`
	KYCName       sql.NullString `json:"kyc_name,omitempty"`
	KYCIDCard     sql.NullString `json:"kyc_id_card,omitempty"`
	KYCPrice      float64        `json:"kyc_price"` // KYC认证单价（元），默认为系统价格，可单独调整
	Status        int            `json:"status"`
	LastLoginAt   *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// AdminUser 管理员用户
type AdminUser struct {
	ID           int64      `json:"id"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"`
	Nickname     string     `json:"nickname"`
	Status       int        `json:"status"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// BalanceLog 余额流水
type BalanceLog struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	OrderID       int64     `json:"order_id,omitempty"` // 关联订单ID（认证订单或支付订单）
	Type          int       `json:"type"`               // 类型：1-充值 2-消费 3-退款 4-赠送
	Amount        float64   `json:"amount"`             // 变动金额
	BalanceBefore float64   `json:"balance_before"`     // 变动前余额
	BalanceAfter  float64   `json:"balance_after"`      // 变动后余额
	Remark        string    `json:"remark,omitempty"`   // 备注
	CreatedAt     time.Time `json:"created_at"`
}

// AuthOrder 认证订单
type AuthOrder struct {
	ID            int64      `json:"id"`
	PlatformBizNo string     `json:"platform_biz_no"`  // 平台业务流水号
	BizNo         string     `json:"biz_no,omitempty"` // 用户业务流水号
	UserID        int64      `json:"user_id"`
	UserPhone     string     `json:"user_phone,omitempty"`     // 联表查询时的用户手机号（管理后台订单列表用）
	ReturnURL     string     `json:"return_url,omitempty"`     // 认证完成后跳转的URL
	NotifyURL     string     `json:"notify_url,omitempty"`     // 异步通知回调URL
	BizExtraData  string     `json:"biz_extra_data,omitempty"` // 额外业务数据
	UpToken       string     `json:"up_token,omitempty"`       // 上游返回的token
	UpBizID       string     `json:"up_biz_id,omitempty"`      // 上游返回的biz_id
	UpRequestID   string     `json:"up_request_id,omitempty"`  // 上游返回的request_id
	ResultCode    string     `json:"result_code,omitempty"`    // 认证结果码
	ResultMessage string     `json:"result_message,omitempty"` // 认证结果消息
	ResultData    string     `json:"result_data,omitempty"`    // 认证结果完整数据（JSON）
	Status        int        `json:"status"`                   // 0-待认证 1-认证中 2-已完成 3-失败 4-已取消 5-超时（已退款）
	Cost          float64    `json:"cost"`                     // 本次认证消耗金额
	IsRefunded    int        `json:"is_refunded"`              // 超时是否已退款：0-否 1-是
	NotifyTimes   int        `json:"notify_times"`             // 回调用户次数
	NotifyStatus  int        `json:"notify_status"`            // 回调用户状态：0-待通知 1-通知成功 2-通知失败
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"` // 完成时间
}

// PaymentOrder 支付订单（充值订单）
type PaymentOrder struct {
	ID             int64      `json:"id"`
	PayOrderNo     string     `json:"pay_order_no"` // 支付流水号
	UserID         int64      `json:"user_id"`
	Amount         float64    `json:"amount"`                     // 充值金额（元）
	Channel        string     `json:"channel"`                    // alipay / wechat / unionpay
	ChannelTradeNo string     `json:"channel_trade_no,omitempty"` // 银联商务交易号（seqId）
	Status         int        `json:"status"`                     // 0-待支付 1-已支付 2-已退款 3-已关闭
	RefundStatus   int        `json:"refund_status"`              // 退款状态：0-未退款 1-部分退款 2-全额退款
	RefundAmount   float64    `json:"refund_amount"`              // 退款金额
	ExpireTime     *time.Time `json:"expire_time,omitempty"`      // 过期时间
	PaidAt         *time.Time `json:"paid_at,omitempty"`          // 支付时间
	RefundedAt     *time.Time `json:"refunded_at,omitempty"`      // 退款时间
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// SystemConfig 系统配置
type SystemConfig struct {
	ID          int64     `json:"id"`
	ConfigKey   string    `json:"config_key"`
	ConfigValue string    `json:"config_value"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}
