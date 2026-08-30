package model

import (
	"database/sql"
	"time"
)

// PlatformUser 平台用户（使用 Web 前端的用户）
type PlatformUser struct {
	ID            int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	Phone         string         `json:"phone" gorm:"size:20;not null;uniqueIndex"`
	Username      string         `json:"username" gorm:"size:50;not null;uniqueIndex"`
	Email         string         `json:"email" gorm:"size:100;not null;uniqueIndex"`
	PasswordHash  string         `json:"-" gorm:"size:255;not null"`
	Balance       float64        `json:"balance" gorm:"type:decimal(10,2);not null;default:0"`
	APIKey        string         `json:"api_key" gorm:"size:64;not null;uniqueIndex"`
	APISecret     string         `json:"-" gorm:"size:64;not null"`
	IsKYCVerified int            `json:"is_kyc_verified" gorm:"type:tinyint;not null;default:0"`
	KYCName       sql.NullString `json:"kyc_name,omitempty" gorm:"size:100"`
	KYCIDCard     sql.NullString `json:"kyc_id_card,omitempty" gorm:"size:100"`
	KYCPrice      float64        `json:"-" gorm:"type:decimal(10,2);not null;default:1"` // 已废弃的个人KYC单价（资源包上线后统一按平台价格扣费），仅保留字段以兼容数据库列
	Status        int            `json:"status" gorm:"type:tinyint;not null;default:1;index"`
	LastLoginAt   *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

// AdminUser 管理员用户
type AdminUser struct {
	ID           int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	Username     string     `json:"username" gorm:"size:50;not null;uniqueIndex"`
	PasswordHash string     `json:"-" gorm:"size:255;not null"`
	Nickname     string     `json:"nickname" gorm:"size:50"`
	Status       int        `json:"status" gorm:"type:tinyint;not null;default:1"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// BalanceLog 余额流水
type BalanceLog struct {
	ID            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID        int64     `json:"user_id" gorm:"not null;index"`
	OrderID       int64     `json:"order_id,omitempty" gorm:"index"` // 关联订单ID（认证订单或支付订单）
	Type          int       `json:"type" gorm:"type:tinyint;not null"` // 类型：1-充值 2-消费 3-退款
	Amount        float64   `json:"amount" gorm:"type:decimal(10,2);not null"` // 变动金额
	BalanceBefore float64   `json:"balance_before" gorm:"type:decimal(10,2);not null"` // 变动前余额
	BalanceAfter  float64   `json:"balance_after" gorm:"type:decimal(10,2);not null"` // 变动后余额
	BankSerialNo  string    `json:"bank_serial_no,omitempty" gorm:"size:100"` // 银行流水单号（人工充值）
	Remark        string    `json:"remark,omitempty" gorm:"size:255"` // 备注
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// AuthOrder 认证订单
type AuthOrder struct {
	ID            int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	PlatformBizNo string     `json:"platform_biz_no" gorm:"size:50;not null;uniqueIndex"` // 平台业务流水号
	BizNo         string     `json:"biz_no,omitempty" gorm:"size:50;index"`               // 用户业务流水号
	UserID        int64      `json:"user_id" gorm:"not null;index"`
	UserPhone     string     `json:"user_phone,omitempty" gorm:"-"` // 联表查询时的用户手机号（管理后台订单列表用）
	ReturnURL     string     `json:"return_url,omitempty" gorm:"size:500"` // 认证完成后跳转的URL
	NotifyURL     string     `json:"notify_url,omitempty" gorm:"size:500"` // 异步通知回调URL
	BizExtraData  string     `json:"biz_extra_data,omitempty" gorm:"type:text"` // 额外业务数据
	UpToken       string     `json:"up_token,omitempty" gorm:"size:100"` // 上游返回的token
	UpBizID       string     `json:"up_biz_id,omitempty" gorm:"size:50"` // 上游返回的biz_id
	UpRequestID   string     `json:"up_request_id,omitempty" gorm:"size:50"` // 上游返回的request_id
	ResultCode    string     `json:"result_code,omitempty" gorm:"size:20"` // 认证结果码
	ResultMessage string     `json:"result_message,omitempty" gorm:"size:255"` // 认证结果消息
	ResultData    string     `json:"result_data,omitempty" gorm:"type:text"` // 认证结果完整数据（JSON）
	Status        int        `json:"status" gorm:"type:tinyint;not null;default:0;index"` // 0-待认证 1-认证中 2-已完成 3-失败 4-已取消 5-超时（已退款）
	Cost          float64    `json:"cost" gorm:"type:decimal(10,2);not null;default:0"` // 本次认证消耗金额
	Source        int        `json:"source" gorm:"type:tinyint;not null;default:2"` // 来源：1-账户实名 2-API调用
	PayType       int        `json:"pay_type" gorm:"type:tinyint;not null;default:0"` // 扣费方式：0-免费 1-余额 2-资源包
	UserPackID    int64      `json:"user_pack_id,omitempty"` // 使用的用户资源包ID（pay_type=2 时）
	IsRefunded    int        `json:"is_refunded" gorm:"type:tinyint;not null;default:0"` // 超时是否已退款：0-否 1-是
	NotifyTimes   int        `json:"notify_times" gorm:"not null;default:0"` // 回调用户次数
	NotifyStatus  int        `json:"notify_status" gorm:"type:tinyint;not null;default:0"` // 回调用户状态：0-待通知 1-通知成功 2-通知失败
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"` // 完成时间
}

// PaymentOrder 支付订单（充值订单）
type PaymentOrder struct {
	ID             int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	PayOrderNo     string     `json:"pay_order_no" gorm:"size:50;not null;uniqueIndex"` // 支付流水号
	UserID         int64      `json:"user_id" gorm:"not null;index"`
	Amount         float64    `json:"amount" gorm:"type:decimal(10,2);not null"` // 充值金额（元）
	Channel        string     `json:"channel" gorm:"size:20;not null"` // 支付渠道：alipay-支付宝 wechat-微信
	ChannelTradeNo string     `json:"channel_trade_no,omitempty" gorm:"size:100"` // 渠道交易号（支付宝 trade_no / 微信 transaction_id）
	Status         int        `json:"status" gorm:"type:tinyint;not null;default:0;index"` // 0-待支付 1-已支付 2-已退款 3-已关闭
	RefundStatus   int        `json:"refund_status" gorm:"type:tinyint;not null;default:0"` // 退款状态：0-未退款 1-部分退款 2-全额退款
	RefundAmount   float64    `json:"refund_amount" gorm:"type:decimal(10,2);default:0"` // 退款金额
	ExpireTime     *time.Time `json:"expire_time,omitempty"` // 过期时间
	PaidAt         *time.Time `json:"paid_at,omitempty"` // 支付时间
	RefundedAt     *time.Time `json:"refunded_at,omitempty"` // 退款时间
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}
