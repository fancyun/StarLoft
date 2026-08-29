package model

import "time"

// InternalAccount 内部账号（本司其他系统专用，无需实名与计费，不可在用户端登录）
type InternalAccount struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`      // 账号名称（标识本司系统，唯一）
	Remark    string    `json:"remark"`    // 备注
	APIKey    string    `json:"api_key"`   // API Key
	APISecret string    `json:"api_secret"` // API Secret
	Status    int       `json:"status"`    // 1-启用 0-禁用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
