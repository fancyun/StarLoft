package model

import "time"

// InternalAccount 内部账号（本司其他系统专用，无需实名与计费，不可在用户端登录）
type InternalAccount struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"size:64;not null;uniqueIndex"` // 账号名称（标识本司系统，唯一）
	Remark    string    `json:"remark" gorm:"size:255;not null;default:''"` // 备注
	APIKey    string    `json:"api_key" gorm:"size:64;not null;uniqueIndex"` // API Key
	APISecret string    `json:"api_secret" gorm:"size:64;not null"` // API Secret
	Status    int       `json:"status" gorm:"type:tinyint;not null;default:1"` // 1-启用 0-禁用
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
